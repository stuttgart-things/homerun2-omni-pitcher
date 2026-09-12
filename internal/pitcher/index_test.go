package pitcher

import (
	"context"
	"errors"
	"strings"
	"testing"

	homerun "github.com/stuttgart-things/homerun-library/v4"
)

func TestEnsureIndexSkipsWhenEmpty(t *testing.T) {
	rp := &RedisPitcher{
		Config: homerun.RedisConfig{
			Index: "",
		},
	}

	// Should return nil immediately when index is not configured
	err := rp.EnsureIndex(context.Background())
	if err != nil {
		t.Errorf("EnsureIndex() with empty index should return nil, got: %v", err)
	}
}

// RediSearch 2.x (redis-stack 7.2, what we run) says "Unknown index name";
// older builds said "no such index". Matching only the latter made EnsureIndex
// report a hard error instead of creating the index, and since main.go only
// logs a warning, the service came up with no search index at all.
func TestIsUnknownIndexError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"redisearch 2.x", errors.New("Unknown index name"), true},
		{"older wording", errors.New("no such index"), true},
		{"lower case", errors.New("unknown index name"), true},
		{"unrelated failure", errors.New("connection refused"), false},
		{"wrong type", errors.New("WRONGTYPE Operation against a key"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isUnknownIndexError(tc.err); got != tc.want {
				t.Errorf("isUnknownIndexError(%q) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// The event time is NUMERIC so time windows can be queried; the message fields
// stay TEXT for FT.AGGREGATE GROUPBY.
func TestIndexCreateArgs(t *testing.T) {
	args := indexCreateArgs("messages")
	var parts []string
	for _, a := range args {
		parts = append(parts, a.(string))
	}
	joined := strings.Join(parts, " ")

	for _, want := range []string{
		"FT.CREATE messages ON JSON SCHEMA",
		"$.severity AS severity TEXT", "$.system AS system TEXT", "$.timestamp AS timestamp TEXT",
		"$.title AS title TEXT", "$.message AS message TEXT", "$.author AS author TEXT", "$.tags AS tags TEXT",
		"$." + homerun.RediSearchTimestampField + " AS " + homerun.RediSearchTimestampField + " NUMERIC SORTABLE",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("FT.CREATE args lack %q:\n%s", want, joined)
		}
	}
}

// Reply shapes captured from redis-stack 7.2 through go-redis v9.
func TestHasNumericAttribute(t *testing.T) {
	resp3 := map[any]any{"index_name": "messages", "attributes": []any{
		map[any]any{"WEIGHT": 1, "attribute": "timestamp", "flags": []any{}, "identifier": "$.timestamp", "type": "TEXT"},
		map[any]any{"attribute": "timestamp_unix", "flags": []any{"SORTABLE", "UNF"}, "identifier": "$.timestamp_unix", "type": "NUMERIC"},
	}}
	resp2 := []any{"index_name", "messages", "attributes", []any{
		[]any{"identifier", "$.timestamp", "attribute", "timestamp", "type", "TEXT", "WEIGHT", "1"},
		[]any{"identifier", "$.timestamp_unix", "attribute", "timestamp_unix", "type", "NUMERIC", "SORTABLE", "UNF"},
	}}
	legacy := map[any]any{"attributes": []any{
		map[any]any{"attribute": "timestamp", "identifier": "$.timestamp", "type": "TEXT"},
	}}

	cases := []struct {
		name      string
		info      any
		attribute string
		want      bool
	}{
		{"resp3 numeric", resp3, "timestamp_unix", true},
		{"resp2 numeric", resp2, "timestamp_unix", true},
		{"resp3 text attribute", resp3, "timestamp", false},
		{"index created before the numeric timestamp", legacy, "timestamp_unix", false},
		{"no attributes", map[any]any{}, "timestamp_unix", false},
		{"unexpected reply", "OK", "timestamp_unix", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasNumericAttribute(tc.info, tc.attribute); got != tc.want {
				t.Errorf("hasNumericAttribute = %v, want %v", got, tc.want)
			}
		})
	}
}
