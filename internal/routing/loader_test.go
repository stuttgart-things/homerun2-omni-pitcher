package routing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemp(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "routes.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

func TestLoadValid(t *testing.T) {
	path := writeTemp(t, `
streams:
  - messages
  - github-events
default_stream: messages
routes:
  - match: { endpoint: /pitch/github }
    stream: github-events
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.DefaultStream != "messages" {
		t.Errorf("default_stream = %q, want %q", cfg.DefaultStream, "messages")
	}
	if len(cfg.Routes) != 1 || cfg.Routes[0].Stream != "github-events" {
		t.Errorf("unexpected routes: %#v", cfg.Routes)
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name:    "empty streams",
			cfg:     Config{DefaultStream: "x"},
			wantErr: "streams must be non-empty",
		},
		{
			name:    "duplicate stream",
			cfg:     Config{Streams: []string{"a", "a"}, DefaultStream: "a"},
			wantErr: "duplicate",
		},
		{
			name:    "default not in allowlist",
			cfg:     Config{Streams: []string{"a"}, DefaultStream: "b"},
			wantErr: "default_stream",
		},
		{
			name: "route stream not in allowlist",
			cfg: Config{
				Streams:       []string{"a"},
				DefaultStream: "a",
				Routes:        []Route{{Match: Match{Endpoint: "/x"}, Stream: "b"}},
			},
			wantErr: "stream \"b\" is not in streams allowlist",
		},
		{
			name: "route without matchers",
			cfg: Config{
				Streams:       []string{"a"},
				DefaultStream: "a",
				Routes:        []Route{{Stream: "a"}},
			},
			wantErr: "at least one matcher is required",
		},
		{
			name: "valid minimal",
			cfg:  Config{Streams: []string{"a"}, DefaultStream: "a"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	path := writeTemp(t, "streams: [oops\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected parse error")
	}
}
