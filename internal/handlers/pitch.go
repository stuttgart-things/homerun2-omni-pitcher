package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/config"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/models"

	homerun "github.com/stuttgart-things/homerun-library"
)

// NewPitchHandler creates a PitchHandler with the given Redis config
func NewPitchHandler(redisConfig config.RedisConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		PitchHandlerWithConfig(w, r, redisConfig)
	}
}

func PitchHandler(w http.ResponseWriter, r *http.Request) {
	redisConfig := config.LoadRedisConfig()
	PitchHandlerWithConfig(w, r, redisConfig)
}

func PitchHandlerWithConfig(w http.ResponseWriter, r *http.Request, redisConfig config.RedisConfig) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg homerun.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate required fields
	if msg.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}
	if msg.Message == "" {
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

	// Enqueue message in Redis Streams
	objectID, streamID := homerun.EnqueueMessageInRedisStreams(
		msg,
		redisConfig.ToMap(),
	)

	// Check if enqueue failed (empty objectID indicates failure)
	if objectID == "" {
		log.Printf("Failed to enqueue message to Redis stream")
		respondWithError(w, http.StatusServiceUnavailable, "Failed to enqueue message to Redis")
		return
	}

	// Respond with success
	respondWithJSON(w, http.StatusOK, models.PitchResponse{
		ObjectID: objectID,
		StreamID: streamID,
		Status:   "success",
		Message:  "Message successfully enqueued",
	})

	log.Printf("Message pitched: objectID=%s, streamID=%s", objectID, streamID)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, models.PitchResponse{
		Status:  "error",
		Message: message,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
