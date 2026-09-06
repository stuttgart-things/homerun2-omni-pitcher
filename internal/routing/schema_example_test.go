package routing_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	homerun "github.com/stuttgart-things/homerun-library/v4"

	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/routing"
)

func message(system string) homerun.Message {
	return homerun.Message{Title: "t", Message: "m", System: system, Severity: "info"}
}

// The kustomize base documents how to enable routing by showing a routing file
// in the `routesConfig` comment. That example is the first thing an operator
// copies, so it should be held against the parser rather than left to drift:
// a snippet that no longer validates is worse than no snippet, because it
// fails at pod start with the process exiting.
const schemaK = "../../kcl/schema.k"

// example pulls the indented YAML out of the routesConfig doc comment.
func example(t *testing.T) string {
	t.Helper()

	source, err := os.ReadFile(filepath.Clean(schemaK))
	if err != nil {
		t.Fatalf("reading %s: %v", schemaK, err)
	}

	// The block between `#     routesConfig: """` and the closing `#     """`.
	re := regexp.MustCompile(`(?s)#\s+routesConfig: """\n(.*?)#\s+"""`)
	match := re.FindSubmatch(source)
	if match == nil {
		t.Fatalf("no routesConfig example found in %s — if it was removed, remove this test with it", schemaK)
	}

	var lines []string
	for _, line := range strings.Split(string(match[1]), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Strip the comment marker and the four spaces of doc indentation.
		lines = append(lines, strings.TrimPrefix(strings.TrimPrefix(trimmed, "#"), "    "))
	}
	return strings.Join(lines, "\n") + "\n"
}

func TestTheRoutingExampleInTheBaseIsValid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "routes.yaml")
	if err := os.WriteFile(path, []byte(example(t)), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := routing.Load(path)
	if err != nil {
		t.Fatalf("the example in kcl/schema.k does not load: %v", err)
	}
	if cfg.DefaultStream == "" {
		t.Fatal("the example has no default_stream, so unmatched messages would have nowhere to go")
	}
	if len(cfg.Routes) == 0 {
		t.Fatal("the example declares no routes, which demonstrates nothing")
	}
}

// And that it does what the comment says it does, rather than merely parsing.
func TestTheRoutingExampleRoutesWhatItClaims(t *testing.T) {
	path := filepath.Join(t.TempDir(), "routes.yaml")
	if err := os.WriteFile(path, []byte(example(t)), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := routing.Load(path)
	if err != nil {
		t.Fatalf("loading the example: %v", err)
	}
	router := routing.New(cfg)

	if got := router.Resolve("/pitch", message("tabletennis")); got != "tabletennis" {
		t.Errorf("a tabletennis message went to %q, not its own stream", got)
	}
	if got := router.Resolve("/pitch", message("github")); got != cfg.DefaultStream {
		t.Errorf("a github message went to %q, not the default stream %q", got, cfg.DefaultStream)
	}
}
