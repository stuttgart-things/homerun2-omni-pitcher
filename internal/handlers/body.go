package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	homerun "github.com/stuttgart-things/homerun-library/v3"
)

// DefaultMaxBodyBytes caps how much of a request body the pitch endpoints will
// read. Payloads here are alerts and webhook events, not uploads, so 1 MiB is
// generous — GitHub itself rejects webhook deliveries larger than 25 MB and a
// realistic event is orders of magnitude smaller.
const DefaultMaxBodyBytes int64 = 1 << 20

// maxBodyBytes is resolved once at package init from MAX_BODY_BYTES, matching
// the "config loaded once at startup" convention in CLAUDE.md.
var maxBodyBytes = loadMaxBodyBytes()

func loadMaxBodyBytes() int64 {
	raw := homerun.GetEnv("MAX_BODY_BYTES", "")
	if raw == "" {
		return DefaultMaxBodyBytes
	}

	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n <= 0 {
		slog.Warn("invalid MAX_BODY_BYTES, falling back to default",
			"value", raw, "default", DefaultMaxBodyBytes)
		return DefaultMaxBodyBytes
	}
	return n
}

// limitBody caps r.Body so a handler cannot be made to buffer an unbounded
// payload. It must be called before anything reads the body — in particular
// before io.ReadAll on the GitHub path, where the signature check would
// otherwise run only after the whole payload is already in memory.
func limitBody(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
}

// isBodyTooLarge reports whether err was caused by exceeding the body cap,
// so the caller can answer 413 instead of a misleading 400.
func isBodyTooLarge(err error) bool {
	var maxErr *http.MaxBytesError
	return errors.As(err, &maxErr)
}
