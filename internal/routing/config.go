// Package routing implements config-driven stream routing for omni-pitcher.
//
// An operator declares an allowlist of valid Redis Streams plus an ordered
// list of rules that pick a stream from that list based on lightweight message
// matchers. Routing is loaded once at startup from a YAML file pointed at by
// ROUTES_CONFIG; misconfiguration fails fast.
package routing

// Config is the on-disk shape of the routing file.
type Config struct {
	Streams       []string `yaml:"streams"`
	DefaultStream string   `yaml:"default_stream"`
	Routes        []Route  `yaml:"routes"`
}

// Route is one ordered rule. All matchers within a rule must match (AND).
// Rules are evaluated top-to-bottom; first match wins.
type Route struct {
	Match  Match  `yaml:"match"`
	Stream string `yaml:"stream"`
}

// Match is the set of substring matchers a single Route can declare.
// All non-zero matchers must match for the Route to fire.
type Match struct {
	Endpoint       string   `yaml:"endpoint,omitempty"`
	System         string   `yaml:"system,omitempty"`
	Author         string   `yaml:"author,omitempty"`
	TagContains    string   `yaml:"tag_contains,omitempty"`
	TitleContains  []string `yaml:"title_contains,omitempty"`
}

// HasAny reports whether the Match has at least one matcher set.
func (m Match) HasAny() bool {
	return m.Endpoint != "" || m.System != "" || m.Author != "" || m.TagContains != "" || len(m.TitleContains) > 0
}
