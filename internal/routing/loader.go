package routing

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load parses the YAML routing config at path and validates it.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read routes config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse routes config %s: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid routes config %s: %w", path, err)
	}

	return &cfg, nil
}

// Validate enforces the allowlist + rule invariants documented in #105.
func (c *Config) Validate() error {
	if len(c.Streams) == 0 {
		return fmt.Errorf("streams must be non-empty")
	}

	seen := make(map[string]struct{}, len(c.Streams))
	for _, s := range c.Streams {
		if s == "" {
			return fmt.Errorf("streams contains an empty entry")
		}
		if _, dup := seen[s]; dup {
			return fmt.Errorf("streams contains duplicate %q", s)
		}
		seen[s] = struct{}{}
	}

	if c.DefaultStream == "" {
		return fmt.Errorf("default_stream is required")
	}
	if _, ok := seen[c.DefaultStream]; !ok {
		return fmt.Errorf("default_stream %q is not in streams allowlist", c.DefaultStream)
	}

	for i, r := range c.Routes {
		if !r.Match.HasAny() {
			return fmt.Errorf("routes[%d]: at least one matcher is required", i)
		}
		if r.Stream == "" {
			return fmt.Errorf("routes[%d]: stream is required", i)
		}
		if _, ok := seen[r.Stream]; !ok {
			return fmt.Errorf("routes[%d]: stream %q is not in streams allowlist", i, r.Stream)
		}
	}

	return nil
}
