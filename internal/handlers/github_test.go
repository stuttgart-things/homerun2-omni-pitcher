package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/models"
)

func computeSignature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestGitHubPitchHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		eventType      string
		payload        any
		pitcher        *recordingPitcher
		secret         string
		signature      string
		expectedStatus int
	}{
		{
			name:           "Method not allowed",
			method:         http.MethodGet,
			pitcher:        &recordingPitcher{},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Missing event header",
			method:         http.MethodPost,
			eventType:      "",
			payload:        models.GitHubWebhookPayload{},
			pitcher:        &recordingPitcher{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ping event",
			method:         http.MethodPost,
			eventType:      "ping",
			payload:        map[string]string{"zen": "test"},
			pitcher:        &recordingPitcher{},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Push event",
			method:    http.MethodPost,
			eventType: "push",
			payload: models.GitHubWebhookPayload{
				Ref:    "refs/heads/main",
				Before: "abc1234567890",
				After:  "def1234567890",
				Pusher: models.GitHubPusher{Name: "developer", Email: "dev@example.com"},
				Repository: models.GitHubRepository{
					FullName: "org/repo",
					HTMLURL:  "https://github.com/org/repo",
					Topics:   []string{"go", "microservice"},
				},
				Commits: []models.GitHubCommit{
					{ID: "def1234567890", Message: "feat: add feature\n\nDetailed description"},
				},
			},
			pitcher:        &recordingPitcher{},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Pull request opened",
			method:    http.MethodPost,
			eventType: "pull_request",
			payload: models.GitHubWebhookPayload{
				Action: "opened",
				Sender: models.GitHubUser{Login: "author", HTMLURL: "https://github.com/author"},
				Repository: models.GitHubRepository{
					FullName: "org/repo",
					HTMLURL:  "https://github.com/org/repo",
				},
				PullRequest: &models.GitHubPullRequest{
					Number:  42,
					Title:   "Add new feature",
					Body:    "This PR adds a new feature",
					HTMLURL: "https://github.com/org/repo/pull/42",
					User:    models.GitHubUser{Login: "author"},
				},
			},
			pitcher:        &recordingPitcher{},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Pull request merged",
			method:    http.MethodPost,
			eventType: "pull_request",
			payload: models.GitHubWebhookPayload{
				Action: "closed",
				Sender: models.GitHubUser{Login: "merger"},
				Repository: models.GitHubRepository{
					FullName: "org/repo",
					HTMLURL:  "https://github.com/org/repo",
				},
				PullRequest: &models.GitHubPullRequest{
					Number:  42,
					Title:   "Add new feature",
					Merged:  true,
					HTMLURL: "https://github.com/org/repo/pull/42",
					User:    models.GitHubUser{Login: "author"},
				},
			},
			pitcher:        &recordingPitcher{},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Issue opened",
			method:    http.MethodPost,
			eventType: "issues",
			payload: models.GitHubWebhookPayload{
				Action: "opened",
				Sender: models.GitHubUser{Login: "reporter"},
				Repository: models.GitHubRepository{
					FullName: "org/repo",
					HTMLURL:  "https://github.com/org/repo",
				},
				Issue: &models.GitHubIssue{
					Number:  10,
					Title:   "Bug report",
					Body:    "Something is broken",
					HTMLURL: "https://github.com/org/repo/issues/10",
					User:    models.GitHubUser{Login: "reporter"},
				},
			},
			pitcher:        &recordingPitcher{},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Release published",
			method:    http.MethodPost,
			eventType: "release",
			payload: models.GitHubWebhookPayload{
				Action: "published",
				Sender: models.GitHubUser{Login: "releaser"},
				Repository: models.GitHubRepository{
					FullName: "org/repo",
					HTMLURL:  "https://github.com/org/repo",
				},
				Release: &models.GitHubRelease{
					TagName: "v1.0.0",
					Name:    "Release 1.0.0",
					Body:    "First stable release",
					HTMLURL: "https://github.com/org/repo/releases/tag/v1.0.0",
					Author:  models.GitHubUser{Login: "releaser"},
				},
			},
			pitcher:        &recordingPitcher{},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Workflow run completed - failure",
			method:    http.MethodPost,
			eventType: "workflow_run",
			payload: models.GitHubWebhookPayload{
				Action: "completed",
				Sender: models.GitHubUser{Login: "ci-bot"},
				Repository: models.GitHubRepository{
					FullName: "org/repo",
					HTMLURL:  "https://github.com/org/repo",
				},
				WorkflowRun: &models.GitHubWorkflowRun{
					Name:       "CI",
					Status:     "completed",
					Conclusion: "failure",
					HTMLURL:    "https://github.com/org/repo/actions/runs/123",
					Actor:      models.GitHubUser{Login: "developer"},
					Event:      "push",
					HeadBranch: "main",
				},
			},
			pitcher:        &recordingPitcher{},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Unknown event type",
			method:    http.MethodPost,
			eventType: "star",
			payload: models.GitHubWebhookPayload{
				Action:     "created",
				Sender:     models.GitHubUser{Login: "fan"},
				Repository: models.GitHubRepository{FullName: "org/repo", HTMLURL: "https://github.com/org/repo"},
			},
			pitcher:        &recordingPitcher{},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.payload != nil {
				body, _ = json.Marshal(tt.payload)
			}

			req, err := http.NewRequest(tt.method, "/pitch/github", bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			if tt.eventType != "" {
				req.Header.Set("X-GitHub-Event", tt.eventType)
			}
			if tt.signature != "" {
				req.Header.Set("X-Hub-Signature-256", tt.signature)
			}

			rr := httptest.NewRecorder()
			handler := NewGitHubPitchHandler(tt.pitcher, tt.secret)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d (body: %s)", tt.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestGitHubSignatureValidation(t *testing.T) {
	secret := "mysecret"
	rp := &recordingPitcher{}

	payload := models.GitHubWebhookPayload{
		Action:     "opened",
		Sender:     models.GitHubUser{Login: "user"},
		Repository: models.GitHubRepository{FullName: "org/repo"},
		Issue: &models.GitHubIssue{
			Number:  1,
			Title:   "Test",
			Body:    "Test body",
			HTMLURL: "https://github.com/org/repo/issues/1",
			User:    models.GitHubUser{Login: "user"},
		},
	}
	body, _ := json.Marshal(payload)

	t.Run("valid signature", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/pitch/github", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GitHub-Event", "issues")
		req.Header.Set("X-Hub-Signature-256", computeSignature(body, secret))

		rr := httptest.NewRecorder()
		handler := NewGitHubPitchHandler(rp, secret)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/pitch/github", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GitHub-Event", "issues")
		req.Header.Set("X-Hub-Signature-256", "sha256=invalidsignature")

		rr := httptest.NewRecorder()
		handler := NewGitHubPitchHandler(rp, secret)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("missing signature when secret configured", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/pitch/github", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GitHub-Event", "issues")

		rr := httptest.NewRecorder()
		handler := NewGitHubPitchHandler(rp, secret)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("no secret configured skips validation", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/pitch/github", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GitHub-Event", "issues")

		rr := httptest.NewRecorder()
		handler := NewGitHubPitchHandler(rp, "")
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})
}

func TestGitHubEventMapping(t *testing.T) {
	t.Run("push event maps correctly", func(t *testing.T) {
		rp := &recordingPitcher{}
		payload := models.GitHubWebhookPayload{
			Ref:    "refs/heads/main",
			Before: "aaa1234567890",
			After:  "bbb1234567890",
			Pusher: models.GitHubPusher{Name: "dev"},
			Repository: models.GitHubRepository{
				FullName: "org/repo",
				HTMLURL:  "https://github.com/org/repo",
				Topics:   []string{"go"},
			},
			Commits: []models.GitHubCommit{
				{ID: "bbb1234567890", Message: "fix: bug"},
			},
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, "/pitch/github", bytes.NewBuffer(body))
		req.Header.Set("X-GitHub-Event", "push")
		rr := httptest.NewRecorder()
		NewGitHubPitchHandler(rp, "").ServeHTTP(rr, req)

		if len(rp.messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(rp.messages))
		}
		msg := rp.messages[0]
		if msg.Title != "Push to org/repo:main" {
			t.Errorf("unexpected title: %s", msg.Title)
		}
		if msg.Author != "dev" {
			t.Errorf("expected author 'dev', got '%s'", msg.Author)
		}
		if msg.System != "org/repo" {
			t.Errorf("expected system 'org/repo', got '%s'", msg.System)
		}
		if msg.Tags != "go" {
			t.Errorf("expected tags 'go', got '%s'", msg.Tags)
		}
	})

	t.Run("merged PR gets success severity", func(t *testing.T) {
		rp := &recordingPitcher{}
		payload := models.GitHubWebhookPayload{
			Action:     "closed",
			Sender:     models.GitHubUser{Login: "merger"},
			Repository: models.GitHubRepository{FullName: "org/repo"},
			PullRequest: &models.GitHubPullRequest{
				Number: 1, Title: "feat", Merged: true,
				HTMLURL: "https://github.com/org/repo/pull/1",
				User:    models.GitHubUser{Login: "author"},
			},
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, "/pitch/github", bytes.NewBuffer(body))
		req.Header.Set("X-GitHub-Event", "pull_request")
		rr := httptest.NewRecorder()
		NewGitHubPitchHandler(rp, "").ServeHTTP(rr, req)

		if rp.messages[0].Severity != "success" {
			t.Errorf("expected severity 'success', got '%s'", rp.messages[0].Severity)
		}
	})

	t.Run("workflow failure gets critical severity", func(t *testing.T) {
		rp := &recordingPitcher{}
		payload := models.GitHubWebhookPayload{
			Action:     "completed",
			Sender:     models.GitHubUser{Login: "bot"},
			Repository: models.GitHubRepository{FullName: "org/repo"},
			WorkflowRun: &models.GitHubWorkflowRun{
				Name: "CI", Conclusion: "failure",
				HTMLURL: "https://github.com/org/repo/actions/runs/1",
				Actor:   models.GitHubUser{Login: "dev"}, Event: "push", HeadBranch: "main",
			},
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, "/pitch/github", bytes.NewBuffer(body))
		req.Header.Set("X-GitHub-Event", "workflow_run")
		rr := httptest.NewRecorder()
		NewGitHubPitchHandler(rp, "").ServeHTTP(rr, req)

		if rp.messages[0].Severity != "critical" {
			t.Errorf("expected severity 'critical', got '%s'", rp.messages[0].Severity)
		}
	})

	t.Run("release maps artifacts to tag", func(t *testing.T) {
		rp := &recordingPitcher{}
		payload := models.GitHubWebhookPayload{
			Action:     "published",
			Sender:     models.GitHubUser{Login: "releaser"},
			Repository: models.GitHubRepository{FullName: "org/repo"},
			Release: &models.GitHubRelease{
				TagName: "v2.0.0", Name: "v2.0.0",
				HTMLURL: "https://github.com/org/repo/releases/tag/v2.0.0",
				Author:  models.GitHubUser{Login: "releaser"},
			},
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, "/pitch/github", bytes.NewBuffer(body))
		req.Header.Set("X-GitHub-Event", "release")
		rr := httptest.NewRecorder()
		NewGitHubPitchHandler(rp, "").ServeHTTP(rr, req)

		if rp.messages[0].Artifacts != "v2.0.0" {
			t.Errorf("expected artifacts 'v2.0.0', got '%s'", rp.messages[0].Artifacts)
		}
	})
}

func TestValidateGitHubSignature(t *testing.T) {
	secret := "test-secret"
	body := []byte(`{"test": true}`)

	t.Run("valid", func(t *testing.T) {
		sig := computeSignature(body, secret)
		if !validateGitHubSignature(body, sig, secret) {
			t.Error("expected valid signature")
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		sig := computeSignature(body, "wrong-secret")
		if validateGitHubSignature(body, sig, secret) {
			t.Error("expected invalid signature")
		}
	})

	t.Run("missing prefix", func(t *testing.T) {
		if validateGitHubSignature(body, "noprefixhex", secret) {
			t.Error("expected invalid signature")
		}
	})

	t.Run("invalid hex", func(t *testing.T) {
		if validateGitHubSignature(body, "sha256=notvalidhex!!!", secret) {
			t.Error("expected invalid signature")
		}
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("firstLine", func(t *testing.T) {
		if got := firstLine("first\nsecond"); got != "first" {
			t.Errorf("expected 'first', got '%s'", got)
		}
		if got := firstLine("no newline"); got != "no newline" {
			t.Errorf("expected 'no newline', got '%s'", got)
		}
	})

	t.Run("shortSHA", func(t *testing.T) {
		if got := shortSHA("abc1234567890"); got != "abc1234" {
			t.Errorf("expected 'abc1234', got '%s'", got)
		}
		if got := shortSHA("short"); got != "short" {
			t.Errorf("expected 'short', got '%s'", got)
		}
	})

	t.Run("truncate", func(t *testing.T) {
		if got := truncate("short", 100); got != "short" {
			t.Errorf("expected 'short', got '%s'", got)
		}
		long := "this is a very long string that should be truncated"
		got := truncate(long, 20)
		if len(got) != 20 {
			t.Errorf("expected length 20, got %d", len(got))
		}
		if got[len(got)-3:] != "..." {
			t.Error("expected truncated string to end with '...'")
		}
	})
}
