package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/metrics"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/models"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/pitcher"

	homerun "github.com/stuttgart-things/homerun-library/v4"
	"github.com/stuttgart-things/homerun-library/v4/routing"
)

// NewPitchHandler creates a pitch handler with the given Pitcher backend.
// If routes is non-nil, the resolved stream is passed as a per-request override.
func NewPitchHandler(p pitcher.Pitcher, routes *routing.StreamRoutes) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		limitBody(w, r)

		var msg homerun.Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			metrics.RecordPitch(metrics.SourceRaw, "", metrics.StatusError)
			metrics.ObservePitchDuration(metrics.SourceRaw, start)
			if isBodyTooLarge(err) {
				respondWithError(w, http.StatusRequestEntityTooLarge, "Request body too large")
				return
			}
			respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		// Required fields and defaults for the optional ones, shared with
		// homerun2-config-viewer's dry run through homerun-library routing.
		pitch, err := routing.PreparePitch(msg, time.Now())
		if err != nil {
			metrics.RecordPitch(metrics.SourceRaw, msg.Severity, metrics.StatusError)
			metrics.ObservePitchDuration(metrics.SourceRaw, start)
			respondWithError(w, http.StatusBadRequest, pitchErrorMessage(err))
			return
		}
		msg = pitch.Message

		stream, _ := routes.Resolve(r.URL.Path, msg)
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

// pitchErrorMessage is the response text for a message routing.PreparePitch
// rejects.
func pitchErrorMessage(err error) string {
	switch {
	case errors.Is(err, routing.ErrPitchTitleRequired):
		return "Title is required"
	case errors.Is(err, routing.ErrPitchMessageRequired):
		return "Message is required"
	default:
		return "Invalid message"
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
