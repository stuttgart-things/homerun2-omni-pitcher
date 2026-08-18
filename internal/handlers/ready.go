package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// ReadinessCheck reports whether a backing dependency is usable. It must
// respect ctx so a hung backend fails the probe instead of hanging it.
type ReadinessCheck func(ctx context.Context) error

// Default bounds for the readiness probe. checkTimeout keeps a slow backend
// from holding the probe open; cacheTTL keeps probe traffic from amplifying
// onto the backend when several probes (or several replicas of a scraper)
// arrive at once.
const (
	DefaultReadyCheckTimeout = 2 * time.Second
	DefaultReadyCacheTTL     = 1 * time.Second
)

// NewReadyHandler returns a readiness handler.
//
// A nil check means "nothing to verify" and the handler always reports ready —
// that is the file-pitcher case, which has no external dependency.
//
// Deliberately separate from /health: /health answers "is this process alive"
// and backs the liveness probe, so it must NOT fail when Redis is down.
// Restarting the pod would not fix a backend outage and would turn it into
// crashloop noise. This endpoint answers "should this pod receive traffic",
// which is exactly what a backend outage should change.
func NewReadyHandler(check ReadinessCheck) http.HandlerFunc {
	return newReadyHandler(check, DefaultReadyCheckTimeout, DefaultReadyCacheTTL)
}

func newReadyHandler(check ReadinessCheck, timeout, ttl time.Duration) http.HandlerFunc {
	var (
		mu       sync.Mutex
		lastErr  error
		lastRun  time.Time
		inflight bool
	)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if check == nil {
			writeReady(w, http.StatusOK, "ready", "")
			return
		}

		mu.Lock()
		fresh := !lastRun.IsZero() && time.Since(lastRun) < ttl
		if fresh || inflight {
			err := lastErr
			mu.Unlock()
			respondReady(w, err)
			return
		}
		inflight = true
		mu.Unlock()

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		err := check(ctx)
		cancel()

		mu.Lock()
		lastErr, lastRun, inflight = err, time.Now(), false
		mu.Unlock()

		respondReady(w, err)
	}
}

func respondReady(w http.ResponseWriter, err error) {
	if err != nil {
		slog.Warn("readiness check failed", "error", err)
		writeReady(w, http.StatusServiceUnavailable, "not ready", err.Error())
		return
	}
	writeReady(w, http.StatusOK, "ready", "")
}

func writeReady(w http.ResponseWriter, code int, status, reason string) {
	body := map[string]string{"status": status}
	if reason != "" {
		body["reason"] = reason
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("error encoding readiness response", "error", err)
	}
}
