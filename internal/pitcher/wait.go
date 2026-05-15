package pitcher

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// WaitForReady retries probe with exponential backoff (1s, 2s, 4s, 8s, capped
// at 16s) until ctx is canceled or probe returns nil. Each attempt runs under
// a child context with timeout perAttemptTimeout.
//
// On success after retries, it logs at info level with the attempt count.
// On each failed attempt before the budget runs out, it logs at warn level.
// On budget exhaustion it returns the last probe error wrapped with the
// attempt count, leaving the os.Exit decision to the caller.
//
// Intended use: smooth over short readiness races at startup (Cilium identity
// propagation in a fresh namespace, sidecar boot delay, etc.) without turning
// genuine misconfiguration into a silent hang.
func WaitForReady(ctx context.Context, probe func(context.Context) error, perAttemptTimeout time.Duration) error {
	const maxBackoff = 16 * time.Second
	backoff := time.Second

	var lastErr error
	for attempt := 1; ; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, perAttemptTimeout)
		err := probe(attemptCtx)
		cancel()
		if err == nil {
			if attempt > 1 {
				slog.Info("readiness probe succeeded after retries", "attempts", attempt)
			}
			return nil
		}
		lastErr = err

		if ctx.Err() != nil {
			return fmt.Errorf("after %d attempts: %w", attempt, lastErr)
		}

		slog.Warn("readiness probe failed, retrying",
			"attempt", attempt,
			"error", err,
			"next_sleep", backoff.String(),
		)

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return fmt.Errorf("after %d attempts: %w", attempt, lastErr)
		}

		if backoff < maxBackoff {
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}
