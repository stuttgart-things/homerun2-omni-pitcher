package pitcher

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/redis/go-redis/v9"
)

// EnsureIndex checks whether the RediSearch index exists and creates it if missing.
// This should be called once at startup before any messages are pitched.
func (p *RedisPitcher) EnsureIndex(ctx context.Context) error {
	if p.Config.Index == "" {
		slog.Debug("redisearch index not configured, skipping ensure")
		return nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     p.Config.Addr + ":" + p.Config.Port,
		Password: p.Config.Password,
	})
	defer func() { _ = client.Close() }()

	// Check if index already exists
	_, err := client.Do(ctx, "FT.INFO", p.Config.Index).Result()
	if err == nil {
		slog.Info("redisearch index already exists", "index", p.Config.Index)
		return nil
	}

	if !isUnknownIndexError(err) {
		return fmt.Errorf("failed to check redisearch index: %w", err)
	}

	slog.Info("redisearch index not found, creating", "index", p.Config.Index)

	// All fields use TEXT (not TAG) for FT.AGGREGATE GROUPBY compatibility.
	// TAG fields on JSON indexes return only the total count without grouped rows.
	args := []any{
		"FT.CREATE", p.Config.Index,
		"ON", "JSON",
		"SCHEMA",
		"$.severity", "AS", "severity", "TEXT",
		"$.system", "AS", "system", "TEXT",
		"$.timestamp", "AS", "timestamp", "TEXT",
		"$.title", "AS", "title", "TEXT",
		"$.message", "AS", "message", "TEXT",
		"$.author", "AS", "author", "TEXT",
		"$.tags", "AS", "tags", "TEXT",
	}

	if err := client.Do(ctx, args...).Err(); err != nil {
		return fmt.Errorf("failed to create redisearch index: %w", err)
	}

	slog.Info("redisearch index created", "index", p.Config.Index)
	return nil
}

// isUnknownIndexError reports whether err is RediSearch saying the index does
// not exist yet, which is the one case EnsureIndex must treat as "create it".
//
// The wording differs by version: RediSearch 2.x (redis-stack 7.2, what we run)
// answers "Unknown index name", older builds "no such index". Matching only the
// latter meant EnsureIndex reported a hard error instead of creating the index -
// and since main.go only logs a warning, the service came up with no search
// index at all.
func isUnknownIndexError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unknown index name") || strings.Contains(msg, "no such index")
}
