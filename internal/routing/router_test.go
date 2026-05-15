package routing

import (
	"testing"

	homerun "github.com/stuttgart-things/homerun-library/v3"
)

func TestNilRouterReturnsEmpty(t *testing.T) {
	var r *Router
	if got := r.Resolve("/pitch", homerun.Message{}); got != "" {
		t.Errorf("nil Router.Resolve = %q, want empty", got)
	}
}

func TestResolveFirstMatchWins(t *testing.T) {
	cfg := &Config{
		Streams:       []string{"messages", "github-events", "grafana-alerts"},
		DefaultStream: "messages",
		Routes: []Route{
			{Match: Match{Endpoint: "/pitch/github"}, Stream: "github-events"},
			{Match: Match{System: "grafana"}, Stream: "grafana-alerts"},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("setup: %v", err)
	}
	r := New(cfg)

	if got := r.Resolve("/pitch/github", homerun.Message{System: "grafana"}); got != "github-events" {
		t.Errorf("first-match-wins broken: got %q, want %q", got, "github-events")
	}
	if got := r.Resolve("/pitch/grafana", homerun.Message{System: "grafana"}); got != "grafana-alerts" {
		t.Errorf("second rule fail: got %q", got)
	}
	if got := r.Resolve("/pitch", homerun.Message{}); got != "messages" {
		t.Errorf("default fallback fail: got %q", got)
	}
}

func TestResolveANDWithinRule(t *testing.T) {
	cfg := &Config{
		Streams:       []string{"a", "b"},
		DefaultStream: "a",
		Routes: []Route{
			{Match: Match{Endpoint: "/pitch", System: "grafana"}, Stream: "b"},
		},
	}
	r := New(cfg)

	if got := r.Resolve("/pitch", homerun.Message{System: "grafana"}); got != "b" {
		t.Errorf("both matchers should match: got %q", got)
	}
	if got := r.Resolve("/pitch", homerun.Message{System: "github"}); got != "a" {
		t.Errorf("one matcher mismatched should fall through: got %q", got)
	}
}

func TestMatchers(t *testing.T) {
	cases := []struct {
		name  string
		match Match
		ep    string
		msg   homerun.Message
		hit   bool
	}{
		{"endpoint substring", Match{Endpoint: "/pitch/github"}, "/pitch/github", homerun.Message{}, true},
		{"endpoint substring partial", Match{Endpoint: "github"}, "/pitch/github", homerun.Message{}, true},
		{"endpoint no match", Match{Endpoint: "github"}, "/pitch/grafana", homerun.Message{}, false},
		{"system substring", Match{System: "graf"}, "/x", homerun.Message{System: "grafana"}, true},
		{"author substring", Match{Author: "depend"}, "/x", homerun.Message{Author: "dependabot[bot]"}, true},
		{"tag_contains", Match{TagContains: "release"}, "/x", homerun.Message{Tags: "ci,release,prod"}, true},
		{"tag_contains miss", Match{TagContains: "release"}, "/x", homerun.Message{Tags: "ci,prod"}, false},
		{"title_contains any", Match{TitleContains: []string{"error", "failure"}}, "/x", homerun.Message{Title: "build failure on main"}, true},
		{"title_contains none", Match{TitleContains: []string{"error", "failure"}}, "/x", homerun.Message{Title: "release 1.0"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := matches(tc.match, tc.ep, tc.msg); got != tc.hit {
				t.Errorf("matches() = %v, want %v", got, tc.hit)
			}
		})
	}
}
