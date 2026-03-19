package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/models"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/pitcher"

	homerun "github.com/stuttgart-things/homerun-library/v3"
)

// NewGrafanaPitchHandler creates a handler that accepts Grafana webhook payloads
// and converts each alert into a homerun.Message for pitching.
func NewGrafanaPitchHandler(p pitcher.Pitcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload models.GrafanaWebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid Grafana webhook payload")
			return
		}

		if len(payload.Alerts) == 0 {
			respondWithError(w, http.StatusBadRequest, "No alerts in payload")
			return
		}

		var results []models.PitchResponse
		var errors []string

		for _, alert := range payload.Alerts {
			msg := grafanaAlertToMessage(alert, payload)

			objectID, streamID, err := p.Pitch(msg)
			if err != nil {
				slog.Error("failed to pitch grafana alert", "error", err, "fingerprint", alert.Fingerprint)
				errors = append(errors, fmt.Sprintf("alert %s: %v", alert.Fingerprint, err))
				continue
			}

			results = append(results, models.PitchResponse{
				ObjectID: objectID,
				StreamID: streamID,
				Status:   "success",
				Message:  fmt.Sprintf("Alert %s enqueued", alert.Fingerprint),
			})

			slog.Info("grafana alert pitched", "objectID", objectID, "streamID", streamID, "fingerprint", alert.Fingerprint)
		}

		if len(errors) > 0 && len(results) == 0 {
			respondWithJSON(w, http.StatusServiceUnavailable, map[string]any{
				"status":  "error",
				"message": "Failed to enqueue all alerts",
				"errors":  errors,
			})
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]any{
			"status":   "success",
			"message":  fmt.Sprintf("%d of %d alerts enqueued", len(results), len(payload.Alerts)),
			"results":  results,
			"errors":   errors,
		})
	}
}

// grafanaAlertToMessage maps a Grafana alert to a homerun.Message.
func grafanaAlertToMessage(alert models.GrafanaAlert, payload models.GrafanaWebhookPayload) homerun.Message {
	// Build title from alert labels (alertname is the convention)
	title := alert.Labels["alertname"]
	if title == "" {
		title = payload.Title
	}
	if title == "" {
		title = "Grafana Alert"
	}

	// Use the summary annotation as message, fall back to description, then payload message
	message := alert.Annotations["summary"]
	if message == "" {
		message = alert.Annotations["description"]
	}
	if message == "" {
		message = payload.Message
	}
	if message == "" {
		message = fmt.Sprintf("Alert %s is %s", title, alert.Status)
	}

	// Map Grafana status to severity
	severity := "info"
	switch alert.Status {
	case "firing":
		severity = mapGrafanaSeverity(alert.Labels["severity"])
	case "resolved":
		severity = "info"
	}

	// Use startsAt as timestamp, fall back to now
	timestamp := alert.StartsAt
	if timestamp == "" {
		timestamp = time.Now().Format(time.RFC3339)
	}

	// Build tags from alert labels (excluding alertname and severity which are mapped elsewhere)
	var tags []string
	for k, v := range alert.Labels {
		if k != "alertname" && k != "severity" {
			tags = append(tags, k+"="+v)
		}
	}

	// Build URL: prefer dashboardURL, then panelURL, then generatorURL
	url := alert.DashboardURL
	if url == "" {
		url = alert.PanelURL
	}
	if url == "" {
		url = alert.GeneratorURL
	}

	return homerun.Message{
		Title:     title,
		Message:   message,
		Severity:  severity,
		Author:    "grafana",
		Timestamp: timestamp,
		System:    payload.Receiver,
		Tags:      strings.Join(tags, ","),
		Url:       url,
	}
}

// mapGrafanaSeverity maps Grafana severity labels to homerun severity levels.
func mapGrafanaSeverity(severity string) string {
	switch strings.ToLower(severity) {
	case "critical", "page":
		return "critical"
	case "warning", "warn":
		return "warning"
	case "info", "informational", "none", "":
		return "info"
	default:
		return severity
	}
}
