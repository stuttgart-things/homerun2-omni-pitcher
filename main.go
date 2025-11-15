package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	homerun "github.com/stuttgart-things/homerun-library"
)

type PitchRequest struct {
	Title           string `json:"title"`
	Message         string `json:"message"`
	Severity        string `json:"severity,omitempty"`
	Author          string `json:"author,omitempty"`
	Timestamp       string `json:"timestamp,omitempty"`
	System          string `json:"system,omitempty"`
	Tags            string `json:"tags,omitempty"`
	AssigneeAddress string `json:"assigneeaddress,omitempty"`
	AssigneeName    string `json:"assigneename,omitempty"`
	Artifacts       string `json:"artifacts,omitempty"`
	Url             string `json:"url,omitempty"`
}

type PitchResponse struct {
	ObjectID string `json:"objectId"`
	StreamID string `json:"streamId"`
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
}

func main() {
	port := homerun.GetEnv("PORT", "8080")

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/pitch", pitchHandler)

	log.Printf("Starting homerun2-omni-pitcher on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func pitchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PitchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate required fields
	if req.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}
	if req.Message == "" {
		respondWithError(w, http.StatusBadRequest, "Message is required")
		return
	}

	// Set defaults for optional fields
	if req.Severity == "" {
		req.Severity = "info"
	}
	if req.Author == "" {
		req.Author = "unknown"
	}
	if req.Timestamp == "" {
		req.Timestamp = time.Now().Format(time.RFC3339)
	}
	if req.System == "" {
		req.System = "homerun2-omni-pitcher"
	}

	// Get Redis connection details from environment variables
	redisAddr := homerun.GetEnv("REDIS_ADDR", "localhost")
	redisPort := homerun.GetEnv("REDIS_PORT", "6379")
	redisPassword := homerun.GetEnv("REDIS_PASSWORD", "")
	redisStream := homerun.GetEnv("REDIS_STREAM", "messages")

	// Create homerun Message
	msg := homerun.Message{
		Title:           req.Title,
		Message:         req.Message,
		Severity:        req.Severity,
		Author:          req.Author,
		Timestamp:       req.Timestamp,
		System:          req.System,
		Tags:            req.Tags,
		AssigneeAddress: req.AssigneeAddress,
		AssigneeName:    req.AssigneeName,
		Artifacts:       req.Artifacts,
		Url:             req.Url,
	}

	// Enqueue message in Redis Streams
	objectID, streamID := homerun.EnqueueMessageInRedisStreams(
		msg,
		map[string]string{
			"addr":     redisAddr,
			"port":     redisPort,
			"password": redisPassword,
			"stream":   redisStream,
		},
	)

	// Respond with success
	respondWithJSON(w, http.StatusOK, PitchResponse{
		ObjectID: objectID,
		StreamID: streamID,
		Status:   "success",
		Message:  "Message successfully enqueued",
	})

	log.Printf("Message pitched: objectID=%s, streamID=%s", objectID, streamID)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, PitchResponse{
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
