package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTokenAuthMiddleware(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	tests := []struct {
		name           string
		authHeader     string
		envToken       string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Missing authorization header",
			authHeader:     "",
			envToken:       "valid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Missing authorization header",
		},
		{
			name:           "Invalid format - no Bearer prefix",
			authHeader:     "Basic some-token",
			envToken:       "valid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid authorization header format",
		},
		{
			name:           "Invalid format - no space",
			authHeader:     "Bearertoken",
			envToken:       "valid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid authorization header format",
		},
		{
			name:           "Server auth not configured",
			authHeader:     "Bearer some-token",
			envToken:       "",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Server authentication not configured",
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer wrong-token",
			envToken:       "valid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid token",
		},
		{
			name:           "Valid token",
			authHeader:     "Bearer valid-token",
			envToken:       "valid-token",
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AUTH_TOKEN", tt.envToken)

			req, err := http.NewRequest("GET", "/test", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler := TokenAuthMiddleware(dummyHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedBody != "" {
				contentType := rr.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
				}
			}
		})
	}
}
