package pitcher

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/redis/go-redis/v9"
	homerun "github.com/stuttgart-things/homerun-library/v4"
)

// EnsureIndex checks whether the RediSearch index exists and creates it if missing.
// This should be called once at startup before any messages are pitched.
//
// An existing index keeps the schema it was created with. If it predates the
// numeric timestamp, EnsureIndex logs how to recreate it instead of dropping it:
// homerun2-scout creates the same index, and dropping it under a running
// service is an operator's decision.
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
	info, err := client.Do(ctx, "FT.INFO", p.Config.Index).Result()
	if err == nil {
		slog.Info("redisearch index already exists", "index", p.Config.Index)
		if !hasNumericAttribute(info, homerun.RediSearchTimestampField) {
			slog.Warn("redisearch index has no NUMERIC "+homerun.RediSearchTimestampField+": time-range queries over it find nothing",
				"index", p.Config.Index,
				"fix", fmt.Sprintf("FT.DROPINDEX %s (without DD, the documents stay), then restart omni-pitcher to recreate it", p.Config.Index))
		}
		return nil
	}

	if !isUnknownIndexError(err) {
		return fmt.Errorf("failed to check redisearch index: %w", err)
	}

	slog.Info("redisearch index not found, creating", "index", p.Config.Index)

	if err := client.Do(ctx, indexCreateArgs(p.Config.Index)...).Err(); err != nil {
		return fmt.Errorf("failed to create redisearch index: %w", err)
	}

	slog.Info("redisearch index created", "index", p.Config.Index)
	return nil
}

// indexCreateArgs is the FT.CREATE command for index. homerun2-scout creates
// the same index (internal/aggregator/index.go); the two definitions must stay
// identical, since whichever service starts first creates it.
//
// The message fields use TEXT (not TAG) for FT.AGGREGATE GROUPBY compatibility:
// TAG fields on JSON indexes return only the total count without grouped rows.
// The event time is NUMERIC, from the Unix-seconds field homerun-library's
// Enqueue writes next to the RFC3339 timestamp, so time windows can be queried.
func indexCreateArgs(index string) []any {
	return []any{
		"FT.CREATE", index,
		"ON", "JSON",
		"SCHEMA",
		"$.severity", "AS", "severity", "TEXT",
		"$.system", "AS", "system", "TEXT",
		"$.timestamp", "AS", "timestamp", "TEXT",
		"$.title", "AS", "title", "TEXT",
		"$.message", "AS", "message", "TEXT",
		"$.author", "AS", "author", "TEXT",
		"$.tags", "AS", "tags", "TEXT",
		"$." + homerun.RediSearchTimestampField, "AS", homerun.RediSearchTimestampField, "NUMERIC", "SORTABLE",
	}
}

// hasNumericAttribute reports whether an FT.INFO reply declares attribute as
// NUMERIC. It reads both reply shapes go-redis returns: RESP3 (the default) a
// map whose attributes are maps, RESP2 a flat list whose attributes are flat
// lists.
func hasNumericAttribute(info any, attribute string) bool {
	for _, a := range infoAttributes(info) {
		fields := map[string]any{}
		switch v := a.(type) {
		case map[any]any:
			for k, val := range v {
				if ks, ok := k.(string); ok {
					fields[ks] = val
				}
			}
		case []any:
			for i := 0; i+1 < len(v); i += 2 {
				if ks, ok := v[i].(string); ok {
					fields[ks] = v[i+1]
				}
			}
		}
		if fields["attribute"] == attribute {
			typ, _ := fields["type"].(string)
			return strings.EqualFold(typ, "NUMERIC")
		}
	}
	return false
}

// infoAttributes returns the attributes list of an FT.INFO reply.
func infoAttributes(info any) []any {
	switch v := info.(type) {
	case map[any]any:
		attrs, _ := v["attributes"].([]any)
		return attrs
	case []any:
		for i := 0; i+1 < len(v); i += 2 {
			if v[i] == "attributes" {
				attrs, _ := v[i+1].([]any)
				return attrs
			}
		}
	}
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
