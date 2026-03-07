package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	homerun "github.com/stuttgart-things/homerun-library"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/handlers"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/models"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/pitcher"
)

type mockPitcher struct {
	err error
}

func (m *mockPitcher) Pitch(_ homerun.Message) (string, string, error) {
	if m.err != nil {
		return "", "", m.err
	}
	return "mock-obj", "mock-stream", nil
}

var _ pitcher.Pitcher = (*mockPitcher)(nil)

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.NewHealthHandler(handlers.BuildInfo{Version: "test", Commit: "abc123", Date: "2026-01-01"}))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("could not unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("expected status to be 'healthy', got '%s'", response["status"])
	}

	if response["version"] != "test" {
		t.Errorf("expected version to be 'test', got '%s'", response["version"])
	}

	if response["commit"] != "abc123" {
		t.Errorf("expected commit to be 'abc123', got '%s'", response["commit"])
	}
}

func TestHealthHandlerMethodNotAllowed(t *testing.T) {
	req, err := http.NewRequest("POST", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.NewHealthHandler(handlers.BuildInfo{Version: "test", Commit: "abc123", Date: "2026-01-01"}))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestPitchHandlerValidation(t *testing.T) {
	p := &mockPitcher{}

	tests := []struct {
		name           string
		payload        string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Empty payload",
			payload:        "{}",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Title is required",
		},
		{
			name:           "Missing message",
			payload:        `{"title":"test"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Message is required",
		},
		{
			name:           "Invalid JSON",
			payload:        `{invalid}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/pitch", bytes.NewBufferString(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handlers.NewPitchHandler(p))
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			var response models.PitchResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Errorf("could not unmarshal response: %v", err)
			}

			if response.Status != "error" {
				t.Errorf("expected status to be 'error', got '%s'", response.Status)
			}

			if response.Message != tt.expectedError {
				t.Errorf("expected error message '%s', got '%s'", tt.expectedError, response.Message)
			}
		})
	}
}

func TestPitchHandlerMethodNotAllowed(t *testing.T) {
	req, err := http.NewRequest("GET", "/pitch", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.NewPitchHandler(&mockPitcher{}))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestPitchHandlerSuccess(t *testing.T) {
	req, err := http.NewRequest("POST", "/pitch", bytes.NewBufferString(`{"title":"test","message":"hello"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := handlers.NewPitchHandler(&mockPitcher{})
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var response models.PitchResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("could not unmarshal response: %v", err)
	}

	if response.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", response.Status)
	}
	if response.ObjectID != "mock-obj" {
		t.Errorf("expected objectID 'mock-obj', got '%s'", response.ObjectID)
	}
}

func TestPitchHandlerBackendError(t *testing.T) {
	req, err := http.NewRequest("POST", "/pitch", bytes.NewBufferString(`{"title":"test","message":"hello"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := handlers.NewPitchHandler(&mockPitcher{err: fmt.Errorf("connection refused")})
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", rr.Code)
	}
}
