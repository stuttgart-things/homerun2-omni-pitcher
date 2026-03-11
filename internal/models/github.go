package models

// GitHubWebhookPayload contains the common fields shared across all GitHub webhook events.
type GitHubWebhookPayload struct {
	Action     string           `json:"action"`
	Sender     GitHubUser       `json:"sender"`
	Repository GitHubRepository `json:"repository"`

	// Event-specific fields (only one will be populated per event type)
	PullRequest *GitHubPullRequest `json:"pull_request,omitempty"`
	Issue       *GitHubIssue       `json:"issue,omitempty"`
	Release     *GitHubRelease     `json:"release,omitempty"`
	WorkflowRun *GitHubWorkflowRun `json:"workflow_run,omitempty"`

	// Push-specific fields (top-level in push events)
	Ref        string         `json:"ref"`
	Before     string         `json:"before"`
	After      string         `json:"after"`
	Commits    []GitHubCommit `json:"commits"`
	Pusher     GitHubPusher   `json:"pusher"`
	HeadCommit *GitHubCommit  `json:"head_commit,omitempty"`
}

type GitHubUser struct {
	Login     string `json:"login"`
	HTMLURL   string `json:"html_url"`
	AvatarURL string `json:"avatar_url"`
}

type GitHubRepository struct {
	FullName string   `json:"full_name"`
	HTMLURL  string   `json:"html_url"`
	Topics   []string `json:"topics"`
}

type GitHubPullRequest struct {
	Number  int        `json:"number"`
	Title   string     `json:"title"`
	Body    string     `json:"body"`
	State   string     `json:"state"`
	HTMLURL string     `json:"html_url"`
	User    GitHubUser `json:"user"`
	Merged  bool       `json:"merged"`
}

type GitHubIssue struct {
	Number  int        `json:"number"`
	Title   string     `json:"title"`
	Body    string     `json:"body"`
	State   string     `json:"state"`
	HTMLURL string     `json:"html_url"`
	User    GitHubUser `json:"user"`
}

type GitHubRelease struct {
	TagName string     `json:"tag_name"`
	Name    string     `json:"name"`
	Body    string     `json:"body"`
	HTMLURL string     `json:"html_url"`
	Author  GitHubUser `json:"author"`
}

type GitHubWorkflowRun struct {
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	Conclusion string     `json:"conclusion"`
	HTMLURL    string     `json:"html_url"`
	Actor      GitHubUser `json:"actor"`
	Event      string     `json:"event"`
	HeadBranch string     `json:"head_branch"`
}

type GitHubCommit struct {
	ID      string       `json:"id"`
	Message string       `json:"message"`
	Author  GitHubAuthor `json:"author"`
	URL     string       `json:"url"`
}

type GitHubAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GitHubPusher struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
