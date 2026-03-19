package pitcher

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	homerun "github.com/stuttgart-things/homerun-library/v3"
)

func TestFilePitcher(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-pitched.log")

	fp := &FilePitcher{Path: path}

	msg := homerun.Message{
		Title:   "Test Title",
		Message: "Test Message",
	}

	objectID, streamID, err := fp.Pitch(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(objectID, "file-") {
		t.Errorf("expected objectID to start with 'file-', got '%s'", objectID)
	}
	if streamID != "file" {
		t.Errorf("expected streamID 'file', got '%s'", streamID)
	}

	// Verify file contents
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read pitch file: %v", err)
	}

	var entry map[string]json.RawMessage
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("failed to parse pitched entry: %v", err)
	}

	if _, ok := entry["message"]; !ok {
		t.Error("expected 'message' field in pitched entry")
	}
	if _, ok := entry["objectID"]; !ok {
		t.Error("expected 'objectID' field in pitched entry")
	}
}

func TestFilePitcherMultipleMessages(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-pitched.log")

	fp := &FilePitcher{Path: path}

	for i := 0; i < 3; i++ {
		_, _, err := fp.Pitch(homerun.Message{
			Title:   "Title",
			Message: "Message",
		})
		if err != nil {
			t.Fatalf("unexpected error on message %d: %v", i, err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read pitch file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}
