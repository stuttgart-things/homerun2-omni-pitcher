package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/routing"
	homerun "github.com/stuttgart-things/homerun-library/v3"
)

// capturingPitcher records the streamOverride passed to Pitch().
type capturingPitcher struct {
	gotOverride string
}

func (c *capturingPitcher) Pitch(_ homerun.Message, streamOverride ...string) (string, string, error) {
	if len(streamOverride) > 0 {
		c.gotOverride = streamOverride[0]
	}
	streamID := "default"
	if c.gotOverride != "" {
		streamID = c.gotOverride
	}
	return "obj", streamID, nil
}

func newRouter(t *testing.T) *routing.Router {
	t.Helper()
	cfg := &routing.Config{
		Streams:       []string{"messages", "github-events", "releases"},
		DefaultStream: "messages",
		Routes: []routing.Route{
			{Match: routing.Match{Endpoint: "/pitch/github"}, Stream: "github-events"},
			{Match: routing.Match{TagContains: "release"}, Stream: "releases"},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("router cfg invalid: %v", err)
	}
	return routing.New(cfg)
}

func TestPitchHandlerRoutesByTagMatcher(t *testing.T) {
	cp := &capturingPitcher{}
	r := newRouter(t)

	body, _ := json.Marshal(homerun.Message{Title: "v2.0", Message: "ship", Tags: "release,prod"})
	req := httptest.NewRequest(http.MethodPost, "/pitch", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	NewPitchHandler(cp, r).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	if cp.gotOverride != "releases" {
		t.Errorf("override = %q, want %q", cp.gotOverride, "releases")
	}
}

func TestPitchHandlerFallsThroughToDefault(t *testing.T) {
	cp := &capturingPitcher{}
	r := newRouter(t)

	body, _ := json.Marshal(homerun.Message{Title: "hello", Message: "world"})
	req := httptest.NewRequest(http.MethodPost, "/pitch", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	NewPitchHandler(cp, r).ServeHTTP(rr, req)

	if cp.gotOverride != "messages" {
		t.Errorf("override = %q, want default %q", cp.gotOverride, "messages")
	}
}

func TestPitchHandlerNilRouterSendsNoOverride(t *testing.T) {
	cp := &capturingPitcher{}

	body, _ := json.Marshal(homerun.Message{Title: "hello", Message: "world"})
	req := httptest.NewRequest(http.MethodPost, "/pitch", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	NewPitchHandler(cp, nil).ServeHTTP(rr, req)

	if cp.gotOverride != "" {
		t.Errorf("override = %q, want empty (legacy single-stream path)", cp.gotOverride)
	}
}
