package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func doReady(t *testing.T, h http.HandlerFunc, method string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, "/ready", nil)
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func TestReadyHandler(t *testing.T) {
	tests := []struct {
		name           string
		check          ReadinessCheck
		expectedStatus int
		expectReason   bool
	}{
		{
			name:           "nil check is always ready",
			check:          nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "healthy backend is ready",
			check:          func(context.Context) error { return nil },
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unreachable backend is not ready",
			check:          func(context.Context) error { return fmt.Errorf("redis ping failed: connection refused") },
			expectedStatus: http.StatusServiceUnavailable,
			expectReason:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doReady(t, NewReadyHandler(tt.check), http.MethodGet)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.expectedStatus)
			}

			var body map[string]string
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if _, ok := body["reason"]; ok != tt.expectReason {
				t.Errorf("reason present = %v, want %v (body: %v)", ok, tt.expectReason, body)
			}
		})
	}
}

func TestReadyHandlerMethodNotAllowed(t *testing.T) {
	rec := doReady(t, NewReadyHandler(nil), http.MethodPost)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

// TestReadyHandlerCachesResult pins the amplification guard: a burst of probes
// must not turn into a burst of backend pings.
func TestReadyHandlerCachesResult(t *testing.T) {
	var calls int32
	h := newReadyHandler(func(context.Context) error {
		atomic.AddInt32(&calls, 1)
		return nil
	}, time.Second, time.Minute)

	for i := 0; i < 20; i++ {
		if rec := doReady(t, h, http.MethodGet); rec.Code != http.StatusOK {
			t.Fatalf("probe %d: status = %d", i, rec.Code)
		}
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("backend checked %d times across 20 probes, want 1", got)
	}
}

// TestReadyHandlerRechecksAfterTTL is the other half: the cache must expire, or
// readiness would never recover (or never fail) after the first result.
func TestReadyHandlerRechecksAfterTTL(t *testing.T) {
	var calls int32
	h := newReadyHandler(func(context.Context) error {
		atomic.AddInt32(&calls, 1)
		return nil
	}, time.Second, 10*time.Millisecond)

	doReady(t, h, http.MethodGet)
	time.Sleep(30 * time.Millisecond)
	doReady(t, h, http.MethodGet)

	if got := atomic.LoadInt32(&calls); got < 2 {
		t.Errorf("backend checked %d times, want >= 2 after TTL expiry", got)
	}
}

// TestReadyHandlerRecovers asserts a pod that lost its backend goes back to
// Ready once the backend returns, rather than latching failed.
func TestReadyHandlerRecovers(t *testing.T) {
	var down atomic.Bool
	down.Store(true)

	h := newReadyHandler(func(context.Context) error {
		if down.Load() {
			return fmt.Errorf("backend down")
		}
		return nil
	}, time.Second, time.Millisecond)

	if rec := doReady(t, h, http.MethodGet); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("while down: status = %d, want 503", rec.Code)
	}

	down.Store(false)
	time.Sleep(5 * time.Millisecond)

	if rec := doReady(t, h, http.MethodGet); rec.Code != http.StatusOK {
		t.Errorf("after recovery: status = %d, want 200", rec.Code)
	}
}

// TestReadyHandlerBoundsSlowCheck asserts a hung backend fails the probe
// instead of hanging it — the handler must pass a deadline to the check.
func TestReadyHandlerBoundsSlowCheck(t *testing.T) {
	h := newReadyHandler(func(ctx context.Context) error {
		deadline, ok := ctx.Deadline()
		if !ok {
			return fmt.Errorf("check received a context with no deadline")
		}
		if time.Until(deadline) > time.Second {
			return fmt.Errorf("deadline too generous: %v", time.Until(deadline))
		}
		<-ctx.Done()
		return ctx.Err()
	}, 20*time.Millisecond, time.Millisecond)

	done := make(chan *httptest.ResponseRecorder, 1)
	go func() { done <- doReady(t, h, http.MethodGet) }()

	select {
	case rec := <-done:
		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want 503 for a hung backend", rec.Code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("probe hung instead of timing out")
	}
}

// TestReadyHandlerConcurrent guards against races in the cache bookkeeping.
func TestReadyHandlerConcurrent(t *testing.T) {
	h := newReadyHandler(func(context.Context) error { return nil }, time.Second, time.Millisecond)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			doReady(t, h, http.MethodGet)
		}()
	}
	wg.Wait()
}
