package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	homerun "github.com/stuttgart-things/homerun-library"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/handlers"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/models"
)

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
			handler := http.HandlerFunc(handlers.NewPitchHandler(homerun.RedisConfig{}))
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
	handler := http.HandlerFunc(handlers.NewPitchHandler(homerun.RedisConfig{}))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}
