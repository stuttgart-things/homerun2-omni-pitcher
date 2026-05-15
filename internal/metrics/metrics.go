// Package metrics declares the Prometheus collectors used by omni-pitcher
// and exposes the recording API. The /metrics endpoint is wired in main.go
// via promhttp.Handler().
package metrics

import (
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Source labels for omni_pitcher_pitches_total / omni_pitcher_pitch_duration_seconds.
// Stable strings so dashboards + alerts can target them directly.
const (
	SourceRaw     = "raw"
	SourceGrafana = "grafana"
	SourceGitHub  = "github"
)

// Status labels for omni_pitcher_pitches_total.
const (
	StatusSuccess = "success"
	StatusError   = "error"
)

var (
	pitchesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "omni_pitcher_pitches_total",
			Help: "Total pitches received, partitioned by inbound source, normalized severity, and outcome.",
		},
		[]string{"source", "severity", "status"},
	)

	pitchDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "omni_pitcher_pitch_duration_seconds",
			Help:    "End-to-end /pitch* handler duration in seconds, partitioned by source.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"source"},
	)

	buildInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "omni_pitcher_build_info",
			Help: "Constant 1 labeled with build metadata so dashboards can correlate metric drift with deploys.",
		},
		[]string{"version", "commit"},
	)
)

// SetBuildInfo records the running binary's version + commit as a constant 1 gauge.
// Call once at startup.
func SetBuildInfo(version, commit string) {
	buildInfo.WithLabelValues(version, commit).Set(1)
}

// RecordPitch increments the pitch counter for the (source, severity, status)
// triple. Severity is lower-cased + defaulted to "unknown" for consistency
// with how dashboards aggregate.
func RecordPitch(source, severity, status string) {
	pitchesTotal.WithLabelValues(source, normalizeSeverity(severity), status).Inc()
}

// ObservePitchDuration records the elapsed time of a single pitch handler call.
func ObservePitchDuration(source string, start time.Time) {
	pitchDuration.WithLabelValues(source).Observe(time.Since(start).Seconds())
}

func normalizeSeverity(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "unknown"
	}
	return s
}
