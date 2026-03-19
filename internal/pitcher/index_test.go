package pitcher

import (
	"context"
	"testing"

	homerun "github.com/stuttgart-things/homerun-library/v3"
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
