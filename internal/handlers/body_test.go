package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// withMaxBodyBytes temporarily shrinks the cap so tests do not have to build
// megabyte payloads, and restores it afterwards.
func withMaxBodyBytes(t *testing.T, n int64) {
	t.Helper()
	orig := maxBodyBytes
	maxBodyBytes = n
	t.Cleanup(func() { maxBodyBytes = orig })
}

func TestLoadMaxBodyBytes(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want int64
	}{
		{name: "unset falls back to default", env: "", want: DefaultMaxBodyBytes},
		{name: "valid override is honoured", env: "2048", want: 2048},
		{name: "non-numeric falls back", env: "banana", want: DefaultMaxBodyBytes},
		{name: "zero falls back", env: "0", want: DefaultMaxBodyBytes},
		{name: "negative falls back", env: "-1", want: DefaultMaxBodyBytes},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != "" {
				t.Setenv("MAX_BODY_BYTES", tt.env)
			}
			if got := loadMaxBodyBytes(); got != tt.want {
				t.Errorf("loadMaxBodyBytes() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestPitchHandlersRejectOversizedBody asserts every pitch endpoint answers 413
// rather than buffering an unbounded payload.
func TestPitchHandlersRejectOversizedBody(t *testing.T) {
	withMaxBodyBytes(t, 64)

	oversized := bytes.Repeat([]byte("a"), 4096)
	p := &mockPitcher{objectID: "obj-1", streamID: "stream-1"}

	tests := []struct {
		name    string
		path    string
		handler http.HandlerFunc
		headers map[string]string
	}{
		{
			name:    "raw pitch",
			path:    "/pitch",
			handler: NewPitchHandler(p, nil),
		},
		{
			name:    "grafana pitch",
			path:    "/pitch/grafana",
			handler: NewGrafanaPitchHandler(p, nil),
		},
		{
			name:    "github pitch",
			path:    "/pitch/github",
			handler: NewGitHubPitchHandler(p, "", nil),
			headers: map[string]string{"X-GitHub-Event": "push"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"title":"t","message":"%s"}`, oversized)
			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(body))
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()

			tt.handler(rec, req)

			if rec.Code != http.StatusRequestEntityTooLarge {
				t.Errorf("status = %d, want %d (body: %s)",
					rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
			}
		})
	}
}

// TestGitHubHandlerCapsBodyBeforeSignatureCheck pins the ordering: an
// unauthenticated caller must not be able to make us buffer an oversized
// payload just to have the signature rejected afterwards.
func TestGitHubHandlerCapsBodyBeforeSignatureCheck(t *testing.T) {
	withMaxBodyBytes(t, 64)

	handler := NewGitHubPitchHandler(&mockPitcher{}, "s3cret", nil)

	body := fmt.Sprintf(`{"action":"opened","x":"%s"}`, strings.Repeat("a", 4096))
	req := httptest.NewRequest(http.MethodPost, "/pitch/github", strings.NewReader(body))
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", "sha256=deadbeef") // deliberately wrong
	rec := httptest.NewRecorder()

	handler(rec, req)

	// 413 (not 401) proves the cap tripped while reading, before the signature
	// check ever saw a fully buffered body.
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want %d — body cap must trip before signature validation",
			rec.Code, http.StatusRequestEntityTooLarge)
	}
}

// TestPitchHandlerAcceptsBodyUnderLimit guards against the cap being so eager
// that legitimate payloads break.
func TestPitchHandlerAcceptsBodyUnderLimit(t *testing.T) {
	withMaxBodyBytes(t, 4096)

	payload, err := json.Marshal(map[string]string{"title": "t", "message": "m"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/pitch", bytes.NewReader(payload))
	rec := httptest.NewRecorder()

	NewPitchHandler(&mockPitcher{objectID: "obj-1", streamID: "stream-1"}, nil)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
}

// TestGitHubHandlerValidSignatureUnderLimit confirms the happy path still works
// with the cap in place.
func TestGitHubHandlerValidSignatureUnderLimit(t *testing.T) {
	withMaxBodyBytes(t, 4096)

	const secret = "s3cret"
	body := []byte(`{"action":"opened"}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/pitch/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", sig)
	rec := httptest.NewRecorder()

	NewGitHubPitchHandler(&mockPitcher{}, secret, nil)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
}
