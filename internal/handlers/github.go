package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/models"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/pitcher"

	homerun "github.com/stuttgart-things/homerun-library/v2"
)

// NewGitHubPitchHandler creates a handler that accepts GitHub webhook payloads
// and converts them into homerun.Message for pitching.
// If webhookSecret is non-empty, the handler validates X-Hub-Signature-256.
func NewGitHubPitchHandler(p pitcher.Pitcher, webhookSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Failed to read request body")
			return
		}

		// Validate webhook signature if secret is configured
		if webhookSecret != "" {
			signature := r.Header.Get("X-Hub-Signature-256")
			if !validateGitHubSignature(body, signature, webhookSecret) {
				respondWithError(w, http.StatusUnauthorized, "Invalid webhook signature")
				return
			}
		}

		eventType := r.Header.Get("X-GitHub-Event")
		if eventType == "" {
			respondWithError(w, http.StatusBadRequest, "Missing X-GitHub-Event header")
			return
		}

		// Handle ping event (GitHub sends this when webhook is first configured)
		if eventType == "ping" {
			respondWithJSON(w, http.StatusOK, map[string]string{
				"status":  "success",
				"message": "pong",
			})
			return
		}

		var payload models.GitHubWebhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid GitHub webhook payload")
			return
		}

		msg := githubEventToMessage(eventType, payload)

		objectID, streamID, err := p.Pitch(msg)
		if err != nil {
			slog.Error("failed to pitch github event", "error", err, "event", eventType)
			respondWithError(w, http.StatusServiceUnavailable, "Failed to enqueue event")
			return
		}

		respondWithJSON(w, http.StatusOK, models.PitchResponse{
			ObjectID: objectID,
			StreamID: streamID,
			Status:   "success",
			Message:  fmt.Sprintf("GitHub %s event enqueued", eventType),
		})

		slog.Info("github event pitched", "objectID", objectID, "streamID", streamID, "event", eventType)
	}
}

// validateGitHubSignature checks the HMAC-SHA256 signature from GitHub.
func validateGitHubSignature(body []byte, signature, secret string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	sig, err := hex.DecodeString(strings.TrimPrefix(signature, "sha256="))
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := mac.Sum(nil)

	return hmac.Equal(sig, expected)
}

// githubEventToMessage maps a GitHub webhook event to a homerun.Message.
func githubEventToMessage(eventType string, payload models.GitHubWebhookPayload) homerun.Message {
	switch eventType {
	case "push":
		return mapPushEvent(payload)
	case "pull_request":
		return mapPullRequestEvent(payload)
	case "issues":
		return mapIssueEvent(payload)
	case "release":
		return mapReleaseEvent(payload)
	case "workflow_run":
		return mapWorkflowRunEvent(payload)
	default:
		return mapGenericEvent(eventType, payload)
	}
}

func mapPushEvent(p models.GitHubWebhookPayload) homerun.Message {
	branch := strings.TrimPrefix(p.Ref, "refs/heads/")
	title := fmt.Sprintf("Push to %s:%s", p.Repository.FullName, branch)

	var messages []string
	for _, c := range p.Commits {
		messages = append(messages, fmt.Sprintf("%.7s %s", c.ID, firstLine(c.Message)))
	}
	message := strings.Join(messages, "; ")
	if message == "" && p.HeadCommit != nil {
		message = firstLine(p.HeadCommit.Message)
	}
	if message == "" {
		message = fmt.Sprintf("Push to %s", branch)
	}

	return homerun.Message{
		Title:     title,
		Message:   message,
		Severity:  "info",
		Author:    p.Pusher.Name,
		Timestamp: time.Now().Format(time.RFC3339),
		System:    p.Repository.FullName,
		Tags:      joinTags(p.Repository.Topics),
		Url:       p.Repository.HTMLURL + "/compare/" + shortSHA(p.Before) + "..." + shortSHA(p.After),
	}
}

func mapPullRequestEvent(p models.GitHubWebhookPayload) homerun.Message {
	pr := p.PullRequest
	if pr == nil {
		return mapGenericEvent("pull_request", p)
	}

	title := fmt.Sprintf("PR #%d %s: %s", pr.Number, p.Action, pr.Title)
	message := pr.Body
	if message == "" {
		message = fmt.Sprintf("Pull request %s by %s", p.Action, pr.User.Login)
	}

	severity := "info"
	if p.Action == "closed" && pr.Merged {
		severity = "success"
	}

	return homerun.Message{
		Title:           title,
		Message:         truncate(message, 500),
		Severity:        severity,
		Author:          pr.User.Login,
		Timestamp:       time.Now().Format(time.RFC3339),
		System:          p.Repository.FullName,
		Tags:            joinTags(p.Repository.Topics),
		Url:             pr.HTMLURL,
		AssigneeName:    p.Sender.Login,
		AssigneeAddress: p.Sender.HTMLURL,
	}
}

func mapIssueEvent(p models.GitHubWebhookPayload) homerun.Message {
	issue := p.Issue
	if issue == nil {
		return mapGenericEvent("issues", p)
	}

	title := fmt.Sprintf("Issue #%d %s: %s", issue.Number, p.Action, issue.Title)
	message := issue.Body
	if message == "" {
		message = fmt.Sprintf("Issue %s by %s", p.Action, issue.User.Login)
	}

	return homerun.Message{
		Title:     title,
		Message:   truncate(message, 500),
		Severity:  "info",
		Author:    issue.User.Login,
		Timestamp: time.Now().Format(time.RFC3339),
		System:    p.Repository.FullName,
		Tags:      joinTags(p.Repository.Topics),
		Url:       issue.HTMLURL,
	}
}

func mapReleaseEvent(p models.GitHubWebhookPayload) homerun.Message {
	rel := p.Release
	if rel == nil {
		return mapGenericEvent("release", p)
	}

	title := fmt.Sprintf("Release %s: %s", p.Action, rel.TagName)
	name := rel.Name
	if name == "" {
		name = rel.TagName
	}

	message := rel.Body
	if message == "" {
		message = fmt.Sprintf("Release %s %s", name, p.Action)
	}

	return homerun.Message{
		Title:     title,
		Message:   truncate(message, 500),
		Severity:  "success",
		Author:    rel.Author.Login,
		Timestamp: time.Now().Format(time.RFC3339),
		System:    p.Repository.FullName,
		Tags:      joinTags(p.Repository.Topics),
		Url:       rel.HTMLURL,
		Artifacts: rel.TagName,
	}
}

func mapWorkflowRunEvent(p models.GitHubWebhookPayload) homerun.Message {
	wf := p.WorkflowRun
	if wf == nil {
		return mapGenericEvent("workflow_run", p)
	}

	title := fmt.Sprintf("Workflow %s: %s", wf.Name, p.Action)
	message := fmt.Sprintf("Workflow %s %s on %s (branch: %s)", wf.Name, wf.Conclusion, wf.Event, wf.HeadBranch)

	severity := "info"
	switch wf.Conclusion {
	case "success":
		severity = "success"
	case "failure":
		severity = "critical"
	case "cancelled", "skipped":
		severity = "warning"
	}

	return homerun.Message{
		Title:     title,
		Message:   message,
		Severity:  severity,
		Author:    wf.Actor.Login,
		Timestamp: time.Now().Format(time.RFC3339),
		System:    p.Repository.FullName,
		Tags:      joinTags(p.Repository.Topics),
		Url:       wf.HTMLURL,
	}
}

func mapGenericEvent(eventType string, p models.GitHubWebhookPayload) homerun.Message {
	title := fmt.Sprintf("GitHub %s event", eventType)
	action := p.Action
	if action == "" {
		action = "triggered"
	}
	message := fmt.Sprintf("%s %s on %s", eventType, action, p.Repository.FullName)

	return homerun.Message{
		Title:     title,
		Message:   message,
		Severity:  "info",
		Author:    p.Sender.Login,
		Timestamp: time.Now().Format(time.RFC3339),
		System:    p.Repository.FullName,
		Tags:      joinTags(p.Repository.Topics),
		Url:       p.Repository.HTMLURL,
	}
}

// Helper functions

func firstLine(s string) string {
	if line, _, found := strings.Cut(s, "\n"); found {
		return line
	}
	return s
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func joinTags(tags []string) string {
	return strings.Join(tags, ",")
}
