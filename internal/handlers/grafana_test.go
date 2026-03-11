package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/models"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/pitcher"

	homerun "github.com/stuttgart-things/homerun-library/v2"
)

// pitchedMessage records a message that was pitched (for test assertions).
type recordingPitcher struct {
	messages []homerun.Message
	err      error
}

func (r *recordingPitcher) Pitch(msg homerun.Message) (string, string, error) {
	if r.err != nil {
		return "", "", r.err
	}
	r.messages = append(r.messages, msg)
	return fmt.Sprintf("obj-%d", len(r.messages)), "stream-1", nil
}

var _ pitcher.Pitcher = (*recordingPitcher)(nil)

func TestGrafanaPitchHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		payload        any
		pitcher        *recordingPitcher
		expectedStatus int
		validateResp   func(t *testing.T, body map[string]any)
	}{
		{
			name:           "Method not allowed - GET",
			method:         http.MethodGet,
			pitcher:        &recordingPitcher{},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			payload:        "not-json",
			pitcher:        &recordingPitcher{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Empty alerts array",
			method: http.MethodPost,
			payload: models.GrafanaWebhookPayload{
				Status: "firing",
				Alerts: []models.GrafanaAlert{},
			},
			pitcher:        &recordingPitcher{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Single firing alert",
			method: http.MethodPost,
			payload: models.GrafanaWebhookPayload{
				Receiver: "test-receiver",
				Status:   "firing",
				Alerts: []models.GrafanaAlert{
					{
						Status:      "firing",
						Labels:      map[string]string{"alertname": "HighCPU", "severity": "critical", "instance": "node-1"},
						Annotations: map[string]string{"summary": "CPU usage is above 90%"},
						StartsAt:    "2026-03-11T10:00:00Z",
						Fingerprint: "abc123",
						GeneratorURL: "http://grafana.local/alerting/abc123",
					},
				},
			},
			pitcher:        &recordingPitcher{},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body map[string]any) {
				if body["status"] != "success" {
					t.Errorf("expected status 'success', got '%s'", body["status"])
				}
			},
		},
		{
			name:   "Multiple alerts",
			method: http.MethodPost,
			payload: models.GrafanaWebhookPayload{
				Receiver: "webhook-receiver",
				Status:   "firing",
				Alerts: []models.GrafanaAlert{
					{
						Status:      "firing",
						Labels:      map[string]string{"alertname": "HighCPU"},
						Annotations: map[string]string{"summary": "CPU high"},
						Fingerprint: "aaa",
					},
					{
						Status:      "resolved",
						Labels:      map[string]string{"alertname": "DiskFull"},
						Annotations: map[string]string{"summary": "Disk ok now"},
						Fingerprint: "bbb",
					},
				},
			},
			pitcher:        &recordingPitcher{},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body map[string]any) {
				msg := body["message"].(string)
				if msg != "2 of 2 alerts enqueued" {
					t.Errorf("expected '2 of 2 alerts enqueued', got '%s'", msg)
				}
			},
		},
		{
			name:   "Backend failure",
			method: http.MethodPost,
			payload: models.GrafanaWebhookPayload{
				Status: "firing",
				Alerts: []models.GrafanaAlert{
					{
						Status:      "firing",
						Labels:      map[string]string{"alertname": "Test"},
						Fingerprint: "fail1",
					},
				},
			},
			pitcher:        &recordingPitcher{err: fmt.Errorf("redis down")},
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.payload == nil {
				req, err = http.NewRequest(tt.method, "/pitch/grafana", nil)
			} else if s, ok := tt.payload.(string); ok {
				req, err = http.NewRequest(tt.method, "/pitch/grafana", bytes.NewBufferString(s))
			} else {
				data, _ := json.Marshal(tt.payload)
				req, err = http.NewRequest(tt.method, "/pitch/grafana", bytes.NewBuffer(data))
			}
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := NewGrafanaPitchHandler(tt.pitcher)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d (body: %s)", tt.expectedStatus, rr.Code, rr.Body.String())
			}

			if tt.validateResp != nil {
				var body map[string]any
				if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				tt.validateResp(t, body)
			}
		})
	}
}

func TestGrafanaAlertToMessage(t *testing.T) {
	payload := models.GrafanaWebhookPayload{
		Receiver: "prod-alerts",
		Title:    "Fallback Title",
		Message:  "Fallback Message",
	}

	t.Run("maps firing alert correctly", func(t *testing.T) {
		alert := models.GrafanaAlert{
			Status:       "firing",
			Labels:       map[string]string{"alertname": "HighMemory", "severity": "warning", "namespace": "prod"},
			Annotations:  map[string]string{"summary": "Memory above 80%", "description": "Detailed description"},
			StartsAt:     "2026-03-11T12:00:00Z",
			Fingerprint:  "xyz",
			DashboardURL: "http://grafana.local/d/abc",
		}

		msg := grafanaAlertToMessage(alert, payload)

		if msg.Title != "HighMemory" {
			t.Errorf("expected title 'HighMemory', got '%s'", msg.Title)
		}
		if msg.Message != "Memory above 80%" {
			t.Errorf("expected message 'Memory above 80%%', got '%s'", msg.Message)
		}
		if msg.Severity != "warning" {
			t.Errorf("expected severity 'warning', got '%s'", msg.Severity)
		}
		if msg.Author != "grafana" {
			t.Errorf("expected author 'grafana', got '%s'", msg.Author)
		}
		if msg.Timestamp != "2026-03-11T12:00:00Z" {
			t.Errorf("expected timestamp '2026-03-11T12:00:00Z', got '%s'", msg.Timestamp)
		}
		if msg.System != "prod-alerts" {
			t.Errorf("expected system 'prod-alerts', got '%s'", msg.System)
		}
		if msg.Url != "http://grafana.local/d/abc" {
			t.Errorf("expected url 'http://grafana.local/d/abc', got '%s'", msg.Url)
		}
	})

	t.Run("resolved alert gets info severity", func(t *testing.T) {
		alert := models.GrafanaAlert{
			Status:      "resolved",
			Labels:      map[string]string{"alertname": "HighCPU", "severity": "critical"},
			Annotations: map[string]string{"summary": "All good now"},
		}

		msg := grafanaAlertToMessage(alert, payload)

		if msg.Severity != "info" {
			t.Errorf("expected severity 'info' for resolved alert, got '%s'", msg.Severity)
		}
	})

	t.Run("falls back to payload title and message", func(t *testing.T) {
		alert := models.GrafanaAlert{
			Status: "firing",
			Labels: map[string]string{},
		}

		msg := grafanaAlertToMessage(alert, payload)

		if msg.Title != "Fallback Title" {
			t.Errorf("expected fallback title, got '%s'", msg.Title)
		}
		if msg.Message != "Fallback Message" {
			t.Errorf("expected fallback message, got '%s'", msg.Message)
		}
	})

	t.Run("falls back to default title when nothing set", func(t *testing.T) {
		emptyPayload := models.GrafanaWebhookPayload{}
		alert := models.GrafanaAlert{
			Status: "firing",
			Labels: map[string]string{},
		}

		msg := grafanaAlertToMessage(alert, emptyPayload)

		if msg.Title != "Grafana Alert" {
			t.Errorf("expected default title 'Grafana Alert', got '%s'", msg.Title)
		}
	})
}

func TestMapGrafanaSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"critical", "critical"},
		{"Critical", "critical"},
		{"page", "critical"},
		{"warning", "warning"},
		{"warn", "warning"},
		{"info", "info"},
		{"informational", "info"},
		{"none", "info"},
		{"", "info"},
		{"custom", "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := mapGrafanaSeverity(tt.input)
			if got != tt.expected {
				t.Errorf("mapGrafanaSeverity(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGrafanaPitchHandlerRecordedMessages(t *testing.T) {
	rp := &recordingPitcher{}

	payload := models.GrafanaWebhookPayload{
		Receiver: "test",
		Status:   "firing",
		Alerts: []models.GrafanaAlert{
			{
				Status:      "firing",
				Labels:      map[string]string{"alertname": "TestAlert", "severity": "critical"},
				Annotations: map[string]string{"summary": "Something broke"},
				StartsAt:    "2026-03-11T10:00:00Z",
				Fingerprint: "rec1",
			},
		},
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/pitch/grafana", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := NewGrafanaPitchHandler(rp)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	if len(rp.messages) != 1 {
		t.Fatalf("expected 1 pitched message, got %d", len(rp.messages))
	}

	msg := rp.messages[0]
	if msg.Title != "TestAlert" {
		t.Errorf("expected title 'TestAlert', got '%s'", msg.Title)
	}
	if msg.Severity != "critical" {
		t.Errorf("expected severity 'critical', got '%s'", msg.Severity)
	}
	if msg.Author != "grafana" {
		t.Errorf("expected author 'grafana', got '%s'", msg.Author)
	}
}
