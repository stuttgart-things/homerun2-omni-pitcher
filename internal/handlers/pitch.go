package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/metrics"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/models"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/pitcher"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/routing"

	homerun "github.com/stuttgart-things/homerun-library/v3"
)

// NewPitchHandler creates a pitch handler with the given Pitcher backend.
// If router is non-nil, the resolved stream is passed as a per-request override.
func NewPitchHandler(p pitcher.Pitcher, router *routing.Router) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var msg homerun.Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			metrics.RecordPitch(metrics.SourceRaw, "", metrics.StatusError)
			metrics.ObservePitchDuration(metrics.SourceRaw, start)
			respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		// Validate required fields
		if msg.Title == "" {
			metrics.RecordPitch(metrics.SourceRaw, msg.Severity, metrics.StatusError)
			metrics.ObservePitchDuration(metrics.SourceRaw, start)
			respondWithError(w, http.StatusBadRequest, "Title is required")
			return
		}
		if msg.Message == "" {
			metrics.RecordPitch(metrics.SourceRaw, msg.Severity, metrics.StatusError)
			metrics.ObservePitchDuration(metrics.SourceRaw, start)
			respondWithError(w, http.StatusBadRequest, "Message is required")
			return
		}

		// Set defaults for optional fields
		if msg.Severity == "" {
			msg.Severity = "info"
		}
		if msg.Author == "" {
			msg.Author = "unknown"
		}
		if msg.Timestamp == "" {
			msg.Timestamp = time.Now().Format(time.RFC3339)
		}
		if msg.System == "" {
			msg.System = "homerun2-omni-pitcher"
		}

		stream := router.Resolve(r.URL.Path, msg)
		objectID, streamID, err := p.Pitch(msg, stream)
		if err != nil {
			metrics.RecordPitch(metrics.SourceRaw, msg.Severity, metrics.StatusError)
			metrics.ObservePitchDuration(metrics.SourceRaw, start)
			slog.Error("failed to pitch message", "error", err)
			respondWithError(w, http.StatusServiceUnavailable, "Failed to enqueue message")
			return
		}

		metrics.RecordPitch(metrics.SourceRaw, msg.Severity, metrics.StatusSuccess)
		metrics.ObservePitchDuration(metrics.SourceRaw, start)

		respondWithJSON(w, http.StatusOK, models.PitchResponse{
			ObjectID: objectID,
			StreamID: streamID,
			Status:   "success",
			Message:  "Message successfully enqueued",
		})

		slog.Info("message pitched", "objectID", objectID, "streamID", streamID)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, models.PitchResponse{
		Status:  "error",
		Message: message,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("error encoding response", "error", err)
	}
}
