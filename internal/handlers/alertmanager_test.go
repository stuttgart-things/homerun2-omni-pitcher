package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	homerun "github.com/stuttgart-things/homerun-library/v4"
	"github.com/stuttgart-things/homerun-library/v4/routing"
)

// amPayload is an Alertmanager v4 webhook body as kube-prometheus-stack sends
// it: one group, one firing and one resolved alert.
const amPayload = `{
  "version": "4",
  "groupKey": "{}/{severity=~\"warning|critical\"}:{alertname=\"KubePodCrashLooping\"}",
  "truncatedAlerts": 0,
  "status": "firing",
  "receiver": "msteams",
  "groupLabels": {"alertname": "KubePodCrashLooping"},
  "commonLabels": {"alertname": "KubePodCrashLooping", "namespace": "homerun2", "prometheus": "monitoring/kube-prometheus"},
  "commonAnnotations": {"runbook_url": "https://runbooks.prometheus-operator.dev/runbooks/kubernetes/kubepodcrashlooping"},
  "externalURL": "http://alertmanager.monitoring:9093",
  "alerts": [
    {
      "status": "firing",
      "labels": {"alertname": "KubePodCrashLooping", "severity": "warning", "namespace": "homerun2", "pod": "omni-pitcher-0", "container": "a,b"},
      "annotations": {"summary": "Pod is crash looping.", "description": "Pod homerun2/omni-pitcher-0 is in waiting state."},
      "startsAt": "2026-10-22T08:00:00Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": "http://prometheus.monitoring:9090/graph?g0.expr=x",
      "fingerprint": "c0ffee01"
    },
    {
      "status": "resolved",
      "labels": {"alertname": "KubePodCrashLooping", "severity": "critical", "namespace": "homerun2", "pod": "omni-pitcher-1"},
      "annotations": {"description": "Pod homerun2/omni-pitcher-1 is in waiting state."},
      "startsAt": "2026-10-22T07:00:00Z",
      "endsAt": "2026-10-22T07:30:00Z",
      "generatorURL": "",
      "fingerprint": "c0ffee02"
    }
  ]
}`

func postAlertmanager(t *testing.T, p *recordingPitcher, routes *routing.StreamRoutes, query, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, PitchPathAlertmanager+query, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	NewAlertmanagerPitchHandler(p, routes).ServeHTTP(rr, req)
	return rr
}

func TestAlertmanagerPitchPerAlert(t *testing.T) {
	rp := &recordingPitcher{}
	rr := postAlertmanager(t, rp, nil, "", amPayload)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body)
	}
	if len(rp.messages) != 2 {
		t.Fatalf("pitched %d messages, want 2", len(rp.messages))
	}

	firing, resolved := rp.messages[0], rp.messages[1]
	want := homerun.Message{
		Title:     "KubePodCrashLooping",
		Message:   "Pod is crash looping.",
		Severity:  "warning",
		Author:    "alertmanager",
		Timestamp: "2026-10-22T08:00:00Z",
		System:    "msteams",
		Tags:      "status=firing,fingerprint=c0ffee01,container=a;b,namespace=homerun2,pod=omni-pitcher-0",
		URL:       "http://prometheus.monitoring:9090/graph?g0.expr=x",
	}
	if firing != want {
		t.Errorf("firing:\n got %+v\nwant %+v", firing, want)
	}

	// Resolved: success, the prefix, the end time, the external URL as the
	// fallback, the description when there is no summary.
	if resolved.Severity != "success" || resolved.Message != "Resolved: Pod homerun2/omni-pitcher-1 is in waiting state." ||
		resolved.Timestamp != "2026-10-22T07:30:00Z" || resolved.URL != "http://alertmanager.monitoring:9093" ||
		!strings.HasPrefix(resolved.Tags, "status=resolved,fingerprint=c0ffee02,") {
		t.Errorf("resolved: %+v", resolved)
	}

	var body struct {
		Message string `json:"message"`
		Results []struct {
			ObjectID string `json:"objectId"`
		} `json:"results"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Message != "2 of 2 alerts enqueued" || len(body.Results) != 2 || body.Results[0].ObjectID != "obj-1" {
		t.Errorf("response %s", rr.Body)
	}
}

func TestAlertmanagerPitchGroup(t *testing.T) {
	rp := &recordingPitcher{}
	if rr := postAlertmanager(t, rp, nil, "?group=true", amPayload); rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body)
	}
	if len(rp.messages) != 1 {
		t.Fatalf("pitched %d messages, want 1", len(rp.messages))
	}
	m := rp.messages[0]
	// The resolved critical alert does not raise the group's severity.
	if m.Title != "KubePodCrashLooping" || m.Severity != "warning" || m.URL != "http://alertmanager.monitoring:9093" ||
		m.Message != "Pod is crash looping.; Pod homerun2/omni-pitcher-1 is in waiting state. (2 alerts)" ||
		m.Tags != "status=firing,alerts=2,namespace=homerun2,prometheus=monitoring/kube-prometheus" {
		t.Errorf("group: %+v", m)
	}

	resolved := strings.ReplaceAll(amPayload, `"status": "firing",
  "receiver"`, `"status": "resolved",
  "receiver"`)
	rp = &recordingPitcher{}
	postAlertmanager(t, rp, nil, "?group=1", resolved)
	if m := rp.messages[0]; m.Severity != "success" || !strings.HasPrefix(m.Message, "Resolved: ") || !strings.HasPrefix(m.Tags, "status=resolved,") {
		t.Errorf("resolved group: %+v", m)
	}
}

func TestAlertmanagerPitchRejects(t *testing.T) {
	for name, tc := range map[string]struct {
		method, body string
		want         int
	}{
		"GET":           {http.MethodGet, "", http.StatusMethodNotAllowed},
		"not JSON":      {http.MethodPost, "nope", http.StatusBadRequest},
		"no alerts":     {http.MethodPost, `{"version":"4","status":"firing","alerts":[]}`, http.StatusBadRequest},
		"wrong version": {http.MethodPost, `{"version":"3","alerts":[{"status":"firing"}]}`, http.StatusBadRequest},
	} {
		t.Run(name, func(t *testing.T) {
			rp := &recordingPitcher{}
			req := httptest.NewRequest(tc.method, PitchPathAlertmanager, bytes.NewBufferString(tc.body))
			rr := httptest.NewRecorder()
			NewAlertmanagerPitchHandler(rp, nil).ServeHTTP(rr, req)
			if rr.Code != tc.want || len(rp.messages) != 0 {
				t.Errorf("status %d, %d pitched; want %d, none", rr.Code, len(rp.messages), tc.want)
			}
		})
	}
}

func TestAlertmanagerPitchBackendDown(t *testing.T) {
	rp := &recordingPitcher{err: errors.New("redis down")}
	if rr := postAlertmanager(t, rp, nil, "", amPayload); rr.Code != http.StatusServiceUnavailable {
		t.Errorf("status %d, want 503", rr.Code)
	}
}

// A route on the new endpoint picks its stream; one on /pitch/grafana does not.
func TestAlertmanagerPitchRoutesByEndpoint(t *testing.T) {
	routes := &routing.StreamRoutes{
		Streams:       []string{"messages", "alerts", "grafana"},
		DefaultStream: "messages",
		Routes: []routing.StreamRoute{
			{Match: routing.RouteMatch{Endpoint: routing.PitchPathGrafana}, Stream: "grafana"},
			{Match: routing.RouteMatch{Endpoint: PitchPathAlertmanager}, Stream: "alerts"},
		},
	}
	if err := routes.Validate(); err != nil {
		t.Fatal(err)
	}
	cp := &capturingPitcher{}
	req := httptest.NewRequest(http.MethodPost, PitchPathAlertmanager, bytes.NewBufferString(amPayload))
	rr := httptest.NewRecorder()
	NewAlertmanagerPitchHandler(cp, routes).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || cp.gotOverride != "alerts" {
		t.Errorf("status %d, stream %q; want the alerts stream", rr.Code, cp.gotOverride)
	}
}
