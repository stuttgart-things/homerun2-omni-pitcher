package pitcher

import (
	"context"
	"errors"
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
