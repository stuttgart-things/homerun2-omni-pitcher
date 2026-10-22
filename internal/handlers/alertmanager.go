package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/metrics"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/models"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/pitcher"

	homerun "github.com/stuttgart-things/homerun-library/v4"
	"github.com/stuttgart-things/homerun-library/v4/routing"
)

// PitchPathAlertmanager takes Alertmanager's native webhook payload (#210).
// Not in homerun-library's routing.PitchPaths yet; a route matches it with
// `endpoint: /pitch/alertmanager` all the same (endpoint is a substring test).
const PitchPathAlertmanager = "/pitch/alertmanager"

// Alert states, as Alertmanager sends them and as the status tag carries them.
const (
	alertFiring   = "firing"
	alertResolved = "resolved"
)

// NewAlertmanagerPitchHandler creates a handler that accepts Alertmanager
// webhook payloads (version 4) and pitches one homerun.Message per alert, or
// one per notification group with ?group=true.
//
// Each message carries the alert's state as the tag status=firing|resolved
// and its fingerprint as fingerprint=<fp>, so a consumer can tell a
// resolution from the alert and pair the two
// (homerun2-notification-catcher#71).
func NewAlertmanagerPitchHandler(p pitcher.Pitcher, routes *routing.StreamRoutes) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		limitBody(w, r)

		fail := func(status int, msg string) {
			metrics.RecordPitch(metrics.SourceAlertmanager, "", metrics.StatusError)
			metrics.ObservePitchDuration(metrics.SourceAlertmanager, start)
			respondWithError(w, status, msg)
		}

		var payload models.AlertmanagerWebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			if isBodyTooLarge(err) {
				fail(http.StatusRequestEntityTooLarge, "Request body too large")
				return
			}
			fail(http.StatusBadRequest, "Invalid Alertmanager webhook payload")
			return
		}
		if payload.Version != "" && payload.Version != "4" {
			fail(http.StatusBadRequest, fmt.Sprintf("Unsupported Alertmanager webhook version %q (want 4)", payload.Version))
			return
		}
		if len(payload.Alerts) == 0 {
			fail(http.StatusBadRequest, "No alerts in payload")
			return
		}

		var msgs []homerun.Message
		var ids []string // per message, for the log and the response
		if group, _ := strconv.ParseBool(r.URL.Query().Get("group")); group {
			msgs = []homerun.Message{alertmanagerGroupToMessage(payload)}
			ids = []string{payload.GroupKey}
		} else {
			for _, a := range payload.Alerts {
				msgs = append(msgs, alertmanagerAlertToMessage(a, payload))
				ids = append(ids, a.Fingerprint)
			}
		}

		var results []models.PitchResponse
		var errors []string
		for i, msg := range msgs {
			stream, _ := routes.Resolve(r.URL.Path, msg)
			objectID, streamID, err := p.Pitch(msg, stream)
			if err != nil {
				metrics.RecordPitch(metrics.SourceAlertmanager, msg.Severity, metrics.StatusError)
				slog.Error("failed to pitch alertmanager alert", "error", err, "id", ids[i])
				errors = append(errors, fmt.Sprintf("alert %s: %v", ids[i], err))
				continue
			}
			metrics.RecordPitch(metrics.SourceAlertmanager, msg.Severity, metrics.StatusSuccess)
			results = append(results, models.PitchResponse{
				ObjectID: objectID,
				StreamID: streamID,
				Status:   "success",
				Message:  fmt.Sprintf("Alert %s enqueued", ids[i]),
			})
			slog.Info("alertmanager alert pitched", "objectID", objectID, "streamID", streamID, "id", ids[i], "status", alertStatus(msg))
		}

		metrics.ObservePitchDuration(metrics.SourceAlertmanager, start)
		if len(errors) > 0 && len(results) == 0 {
			respondWithJSON(w, http.StatusServiceUnavailable, map[string]any{
				"status":  "error",
				"message": "Failed to enqueue all alerts",
				"errors":  errors,
			})
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]any{
			"status":  "success",
			"message": fmt.Sprintf("%d of %d alerts enqueued", len(results), len(msgs)),
			"results": results,
			"errors":  errors,
		})
	}
}

// alertmanagerAlertToMessage maps one alert. Title is the alertname, the
// message its summary (then description, then the group's), the severity
// its severity label -- or success when resolved, as on /pitch/grafana.
func alertmanagerAlertToMessage(a models.AlertmanagerAlert, p models.AlertmanagerWebhookPayload) homerun.Message {
	status := a.Status
	if status == "" {
		status = p.Status
	}
	title := firstNonEmpty(a.Labels["alertname"], p.CommonLabels["alertname"], "Alertmanager Alert")
	message := firstNonEmpty(
		a.Annotations["summary"], a.Annotations["description"],
		p.CommonAnnotations["summary"], p.CommonAnnotations["description"],
		fmt.Sprintf("Alert %s is %s", title, status),
	)

	// A resolution is stamped with when it ended, an alert with when it began.
	timestamp := a.StartsAt
	if status == alertResolved && a.EndsAt != "" && !strings.HasPrefix(a.EndsAt, "0001-") {
		timestamp = a.EndsAt
	}

	tags := []string{"status=" + status}
	if a.Fingerprint != "" {
		tags = append(tags, "fingerprint="+a.Fingerprint)
	}
	return homerun.Message{
		Title:     title,
		Message:   withResolvedPrefix(message, status),
		Severity:  alertSeverity(status, a.Labels["severity"]),
		Author:    "alertmanager",
		Timestamp: firstNonEmpty(timestamp, time.Now().Format(time.RFC3339)),
		System:    p.Receiver,
		Tags:      strings.Join(append(tags, labelTags(a.Labels)...), ","),
		URL:       firstNonEmpty(a.GeneratorURL, p.ExternalURL),
	}
}

// alertmanagerGroupToMessage maps a whole notification group to one message:
// the common labels and annotations, the highest severity of its firing
// alerts, and alerts=<n>.
func alertmanagerGroupToMessage(p models.AlertmanagerWebhookPayload) homerun.Message {
	title := firstNonEmpty(p.CommonLabels["alertname"], p.GroupLabels["alertname"], "Alertmanager")

	message := firstNonEmpty(p.CommonAnnotations["summary"], p.CommonAnnotations["description"])
	if message == "" {
		var parts []string
		for _, a := range p.Alerts {
			if s := firstNonEmpty(a.Annotations["summary"], a.Annotations["description"]); s != "" {
				parts = append(parts, s)
			}
		}
		message = strings.Join(parts, "; ")
	}
	n := len(p.Alerts) + p.TruncatedAlerts
	if message == "" {
		message = fmt.Sprintf("%d alerts %s", n, p.Status)
	} else if n > 1 {
		message = fmt.Sprintf("%s (%d alerts)", message, n)
	}

	severity := "success"
	if p.Status != alertResolved {
		severity = "info"
		for _, a := range p.Alerts {
			if a.Status == alertResolved {
				continue
			}
			if s := alertSeverity(alertFiring, a.Labels["severity"]); severityRank(s) > severityRank(severity) {
				severity = s
			}
		}
	}

	tags := []string{"status=" + p.Status, "alerts=" + strconv.Itoa(n)}
	return homerun.Message{
		Title:     title,
		Message:   withResolvedPrefix(message, p.Status),
		Severity:  severity,
		Author:    "alertmanager",
		Timestamp: time.Now().Format(time.RFC3339),
		System:    p.Receiver,
		Tags:      strings.Join(append(tags, labelTags(p.CommonLabels)...), ","),
		URL:       p.ExternalURL,
	}
}

// alertSeverity is the severity label mapped like /pitch/grafana's, or
// success for a resolved alert.
func alertSeverity(status, label string) string {
	if status == alertResolved {
		return "success"
	}
	return mapGrafanaSeverity(label)
}

func severityRank(s string) int {
	switch s {
	case "critical":
		return 3
	case "warning":
		return 2
	case "info":
		return 1
	default:
		return 0
	}
}

func withResolvedPrefix(message, status string) string {
	if status == alertResolved && !strings.HasPrefix(message, resolvedPrefix) {
		return resolvedPrefix + message
	}
	return message
}

// labelTags turns labels into sorted key=value tags, without the ones the
// message carries elsewhere (alertname, severity) or that the handler sets
// (status, fingerprint, alerts). A comma in a value would split the tag, so
// it becomes a semicolon.
func labelTags(labels map[string]string) []string {
	var tags []string
	for k, v := range labels {
		switch k {
		case "alertname", "severity", "status", "fingerprint", "alerts":
			continue
		}
		tags = append(tags, k+"="+strings.ReplaceAll(v, ",", ";"))
	}
	sort.Strings(tags)
	return tags
}

// alertStatus reads the status tag back, for the log line.
func alertStatus(msg homerun.Message) string {
	for _, t := range strings.Split(msg.Tags, ",") {
		if v, ok := strings.CutPrefix(t, "status="); ok {
			return v
		}
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
