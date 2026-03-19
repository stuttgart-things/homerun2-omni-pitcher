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
	homerun "github.com/stuttgart-things/homerun-library/v3"
)

// mockPitcher is a test pitcher that returns configurable results.
type mockPitcher struct {
	objectID string
	streamID string
	err      error
}

func (m *mockPitcher) Pitch(_ homerun.Message) (string, string, error) {
	return m.objectID, m.streamID, m.err
}

// Compile-time check that mockPitcher implements Pitcher.
var _ pitcher.Pitcher = (*mockPitcher)(nil)

func TestPitchHandler(t *testing.T) {
	successPitcher := &mockPitcher{objectID: "obj-1", streamID: "stream-1"}
	failPitcher := &mockPitcher{err: fmt.Errorf("backend error")}

	tests := []struct {
		name           string
		method         string
		payload        interface{}
		pitcher        pitcher.Pitcher
		expectedStatus int
		validateResp   func(t *testing.T, resp models.PitchResponse)
	}{
		{
			name:           "Method not allowed - GET",
			method:         http.MethodGet,
			payload:        nil,
			pitcher:        successPitcher,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Method not allowed - PUT",
			method:         http.MethodPut,
			payload:        nil,
			pitcher:        successPitcher,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Method not allowed - DELETE",
			method:         http.MethodDelete,
			payload:        nil,
			pitcher:        successPitcher,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid JSON payload",
			method:         http.MethodPost,
			payload:        "invalid-json",
			pitcher:        successPitcher,
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, resp models.PitchResponse) {
				if resp.Message != "Invalid JSON payload" {
					t.Errorf("expected message 'Invalid JSON payload', got '%s'", resp.Message)
				}
			},
		},
		{
			name:   "Missing title",
			method: http.MethodPost,
			payload: homerun.Message{
				Message: "test message",
			},
			pitcher:        successPitcher,
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, resp models.PitchResponse) {
				if resp.Message != "Title is required" {
					t.Errorf("expected message 'Title is required', got '%s'", resp.Message)
				}
			},
		},
		{
			name:   "Missing message",
			method: http.MethodPost,
			payload: homerun.Message{
				Title: "test title",
			},
			pitcher:        successPitcher,
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, resp models.PitchResponse) {
				if resp.Message != "Message is required" {
					t.Errorf("expected message 'Message is required', got '%s'", resp.Message)
				}
			},
		},
		{
			name:   "Successful pitch",
			method: http.MethodPost,
			payload: homerun.Message{
				Title:   "test title",
				Message: "test message",
			},
			pitcher:        successPitcher,
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, resp models.PitchResponse) {
				if resp.Status != "success" {
					t.Errorf("expected status 'success', got '%s'", resp.Status)
				}
				if resp.ObjectID != "obj-1" {
					t.Errorf("expected objectID 'obj-1', got '%s'", resp.ObjectID)
				}
				if resp.StreamID != "stream-1" {
					t.Errorf("expected streamID 'stream-1', got '%s'", resp.StreamID)
				}
			},
		},
		{
			name:   "Backend failure",
			method: http.MethodPost,
			payload: homerun.Message{
				Title:   "test title",
				Message: "test message",
			},
			pitcher:        failPitcher,
			expectedStatus: http.StatusServiceUnavailable,
			validateResp: func(t *testing.T, resp models.PitchResponse) {
				if resp.Status != "error" {
					t.Errorf("expected status 'error', got '%s'", resp.Status)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.payload == nil {
				req, err = http.NewRequest(tt.method, "/pitch", nil)
			} else if payloadStr, ok := tt.payload.(string); ok {
				req, err = http.NewRequest(tt.method, "/pitch", bytes.NewBufferString(payloadStr))
			} else {
				payloadBytes, _ := json.Marshal(tt.payload)
				req, err = http.NewRequest(tt.method, "/pitch", bytes.NewBuffer(payloadBytes))
			}

			if err != nil {
				t.Fatal(err)
			}

			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := NewPitchHandler(tt.pitcher)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.validateResp != nil {
				var response models.PitchResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("could not unmarshal response: %v", err)
					return
				}
				tt.validateResp(t, response)
			}
		})
	}
}

func TestRespondWithError(t *testing.T) {
	rr := httptest.NewRecorder()
	respondWithError(rr, http.StatusBadRequest, "Invalid input")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response models.PitchResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("could not unmarshal response: %v", err)
	}

	if response.Status != "error" || response.Message != "Invalid input" {
		t.Errorf("unexpected response: %+v", response)
	}
}

func TestRespondWithJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	respondWithJSON(rr, http.StatusOK, models.PitchResponse{
		ObjectID: "test-id",
		Status:   "success",
		Message:  "ok",
	})

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", ct)
	}

	var response models.PitchResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("could not unmarshal response: %v", err)
	}

	if response.ObjectID != "test-id" {
		t.Errorf("expected objectID 'test-id', got '%s'", response.ObjectID)
	}
}
