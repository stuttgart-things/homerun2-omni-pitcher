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

	if !strings.Contains(err.Error(), "no such index") {
		return fmt.Errorf("failed to check redisearch index: %w", err)
	}

	slog.Info("redisearch index not found, creating", "index", p.Config.Index)

	args := []any{
		"FT.CREATE", p.Config.Index,
		"ON", "JSON",
		"SCHEMA",
		"$.severity", "AS", "severity", "TAG",
		"$.system", "AS", "system", "TAG",
		"$.timestamp", "AS", "timestamp", "TEXT",
		"$.title", "AS", "title", "TEXT",
		"$.message", "AS", "message", "TEXT",
		"$.author", "AS", "author", "TAG",
		"$.tags", "AS", "tags", "TAG",
	}

	if err := client.Do(ctx, args...).Err(); err != nil {
		return fmt.Errorf("failed to create redisearch index: %w", err)
	}

	slog.Info("redisearch index created", "index", p.Config.Index)
	return nil
}
