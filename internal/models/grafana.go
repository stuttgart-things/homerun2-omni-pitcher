package models

// GrafanaWebhookPayload represents the JSON body sent by Grafana webhook contact points.
type GrafanaWebhookPayload struct {
	Receiver          string              `json:"receiver"`
	Status            string              `json:"status"`
	OrgID             int                 `json:"orgId"`
	Alerts            []GrafanaAlert      `json:"alerts"`
	GroupLabels       map[string]string   `json:"groupLabels"`
	CommonLabels      map[string]string   `json:"commonLabels"`
	CommonAnnotations map[string]string   `json:"commonAnnotations"`
	ExternalURL       string              `json:"externalURL"`
	Version           string              `json:"version"`
	GroupKey          string              `json:"groupKey"`
	TruncatedAlerts   int                 `json:"truncatedAlerts"`
	State             string              `json:"state"`
	Title             string              `json:"title"`
	Message           string              `json:"message"`
}

// GrafanaAlert represents a single alert within a Grafana webhook payload.
type GrafanaAlert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	Values       map[string]any    `json:"values"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
	SilenceURL   string            `json:"silenceURL"`
	DashboardURL string            `json:"dashboardURL"`
	PanelURL     string            `json:"panelURL"`
	ImageURL     string            `json:"imageURL"`
}
