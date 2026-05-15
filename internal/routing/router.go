package routing

import (
	"strings"

	homerun "github.com/stuttgart-things/homerun-library/v3"
)

// Router resolves an effective Redis stream for a given endpoint + message.
// A nil Router is valid and always returns "" (meaning "no override").
type Router struct {
	cfg *Config
}

// New wraps a validated Config in a Router.
func New(cfg *Config) *Router {
	if cfg == nil {
		return nil
	}
	return &Router{cfg: cfg}
}

// Resolve walks the rule list and returns the matching stream. If no rule
// matches, it falls back to DefaultStream. If the Router is nil (no config
// loaded), it returns "" so the caller keeps its legacy single-stream path.
func (r *Router) Resolve(endpoint string, msg homerun.Message) string {
	if r == nil || r.cfg == nil {
		return ""
	}
	for _, rule := range r.cfg.Routes {
		if matches(rule.Match, endpoint, msg) {
			return rule.Stream
		}
	}
	return r.cfg.DefaultStream
}

// Streams returns the configured allowlist (for startup logging).
func (r *Router) Streams() []string {
	if r == nil || r.cfg == nil {
		return nil
	}
	return r.cfg.Streams
}

// DefaultStream returns the configured fallback stream.
func (r *Router) DefaultStream() string {
	if r == nil || r.cfg == nil {
		return ""
	}
	return r.cfg.DefaultStream
}

// Routes returns the configured rule list.
func (r *Router) Routes() []Route {
	if r == nil || r.cfg == nil {
		return nil
	}
	return r.cfg.Routes
}

func matches(m Match, endpoint string, msg homerun.Message) bool {
	if m.Endpoint != "" && !strings.Contains(endpoint, m.Endpoint) {
		return false
	}
	if m.System != "" && !strings.Contains(msg.System, m.System) {
		return false
	}
	if m.Author != "" && !strings.Contains(msg.Author, m.Author) {
		return false
	}
	if m.TagContains != "" && !strings.Contains(msg.Tags, m.TagContains) {
		return false
	}
	if len(m.TitleContains) > 0 {
		hit := false
		for _, needle := range m.TitleContains {
			if needle == "" {
				continue
			}
			if strings.Contains(msg.Title, needle) {
				hit = true
				break
			}
		}
		if !hit {
			return false
		}
	}
	return true
}
