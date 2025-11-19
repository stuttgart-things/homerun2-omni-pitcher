package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/models"
	homerun "github.com/stuttgart-things/homerun-library"
)

// Note: Tests that validate successful message enqueuing require a live Redis instance
// or mocking the homerun library. These tests focus on validation logic and error cases
// that can be tested without Redis connectivity.

func TestPitchHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		payload        interface{}
		expectedStatus int
		validateResp   func(t *testing.T, resp models.PitchResponse)
		skipRedis      bool // Skip tests that require Redis
	}{
		{
			name:           "Method not allowed - GET",
			method:         http.MethodGet,
			payload:        nil,
			expectedStatus: http.StatusMethodNotAllowed,
			validateResp:   nil,
			skipRedis:      true,
		},
		{
			name:           "Method not allowed - PUT",
			method:         http.MethodPut,
			payload:        nil,
			expectedStatus: http.StatusMethodNotAllowed,
			validateResp:   nil,
			skipRedis:      true,
		},
		{
			name:           "Method not allowed - DELETE",
			method:         http.MethodDelete,
			payload:        nil,
			expectedStatus: http.StatusMethodNotAllowed,
			validateResp:   nil,
			skipRedis:      true,
		},
		{
			name:           "Invalid JSON payload",
			method:         http.MethodPost,
			payload:        "invalid-json",
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, resp models.PitchResponse) {
				if resp.Status != "error" {
					t.Errorf("expected status 'error', got '%s'", resp.Status)
				}
				if resp.Message != "Invalid JSON payload" {
					t.Errorf("expected message 'Invalid JSON payload', got '%s'", resp.Message)
				}
			},
			skipRedis: true,
		},
		{
			name:   "Missing title",
			method: http.MethodPost,
			payload: homerun.Message{
				Message: "test message",
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, resp models.PitchResponse) {
				if resp.Status != "error" {
					t.Errorf("expected status 'error', got '%s'", resp.Status)
				}
				if resp.Message != "Title is required" {
					t.Errorf("expected message 'Title is required', got '%s'", resp.Message)
				}
			},
			skipRedis: true,
		},
		{
			name:   "Missing message",
			method: http.MethodPost,
			payload: homerun.Message{
				Title: "test title",
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, resp models.PitchResponse) {
				if resp.Status != "error" {
					t.Errorf("expected status 'error', got '%s'", resp.Status)
				}
				if resp.Message != "Message is required" {
					t.Errorf("expected message 'Message is required', got '%s'", resp.Message)
				}
			},
			skipRedis: true,
		},
		{
			name:   "Empty title",
			method: http.MethodPost,
			payload: homerun.Message{
				Title:   "",
				Message: "test message",
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, resp models.PitchResponse) {
				if resp.Status != "error" {
					t.Errorf("expected status 'error', got '%s'", resp.Status)
				}
				if resp.Message != "Title is required" {
					t.Errorf("expected message 'Title is required', got '%s'", resp.Message)
				}
			},
			skipRedis: true,
		},
		{
			name:   "Empty message",
			method: http.MethodPost,
			payload: homerun.Message{
				Title:   "test title",
				Message: "",
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, resp models.PitchResponse) {
				if resp.Status != "error" {
					t.Errorf("expected status 'error', got '%s'", resp.Status)
				}
				if resp.Message != "Message is required" {
					t.Errorf("expected message 'Message is required', got '%s'", resp.Message)
				}
			},
			skipRedis: true,
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
			handler := http.HandlerFunc(PitchHandler)
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

// TestPitchHandlerDefaultValues tests that default values are properly set
// Note: This test validates the logic but cannot verify Redis enqueuing without integration testing
func TestPitchHandlerDefaultValues(t *testing.T) {
	// This test documents expected behavior for default values
	// In practice, these would be validated in integration tests with actual Redis
	tests := []struct {
		name     string
		input    homerun.Message
		defaults map[string]string
	}{
		{
			name: "Sets default severity",
			input: homerun.Message{
				Title:   "test",
				Message: "test",
			},
			defaults: map[string]string{
				"severity": "info",
			},
		},
		{
			name: "Sets default author",
			input: homerun.Message{
				Title:   "test",
				Message: "test",
			},
			defaults: map[string]string{
				"author": "unknown",
			},
		},
		{
			name: "Sets default system",
			input: homerun.Message{
				Title:   "test",
				Message: "test",
			},
			defaults: map[string]string{
				"system": "homerun2-omni-pitcher",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Document expected defaults
			t.Logf("Expected defaults: %v", tt.defaults)
			// Actual validation would require mocking or integration testing
		})
	}
}

func TestRespondWithError(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		message  string
		expected models.PitchResponse
	}{
		{
			name:    "Bad request error",
			code:    http.StatusBadRequest,
			message: "Invalid input",
			expected: models.PitchResponse{
				Status:  "error",
				Message: "Invalid input",
			},
		},
		{
			name:    "Internal server error",
			code:    http.StatusInternalServerError,
			message: "Something went wrong",
			expected: models.PitchResponse{
				Status:  "error",
				Message: "Something went wrong",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			respondWithError(rr, tt.code, tt.message)

			if status := rr.Code; status != tt.code {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.code)
			}

			var response models.PitchResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Errorf("could not unmarshal response: %v", err)
				return
			}

			if response.Status != tt.expected.Status {
				t.Errorf("expected status '%s', got '%s'", tt.expected.Status, response.Status)
			}

			if response.Message != tt.expected.Message {
				t.Errorf("expected message '%s', got '%s'", tt.expected.Message, response.Message)
			}
		})
	}
}

func TestRespondWithJSON(t *testing.T) {
	tests := []struct {
		name           string
		code           int
		payload        interface{}
		expectedStatus int
	}{
		{
			name: "Success response",
			code: http.StatusOK,
			payload: models.PitchResponse{
				ObjectID: "test-object-id",
				StreamID: "test-stream-id",
				Status:   "success",
				Message:  "Test message",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Created response",
			code: http.StatusCreated,
			payload: models.PitchResponse{
				Status:  "success",
				Message: "Resource created",
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			respondWithJSON(rr, tt.code, tt.payload)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			contentType := rr.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
			}

			var response models.PitchResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Errorf("could not unmarshal response: %v", err)
			}
		})
	}
}
