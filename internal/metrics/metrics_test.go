package metrics

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRecordPitch_IncrementsCounter(t *testing.T) {
	before := testutil.ToFloat64(pitchesTotal.WithLabelValues(SourceRaw, "info", StatusSuccess))
	RecordPitch(SourceRaw, "info", StatusSuccess)
	after := testutil.ToFloat64(pitchesTotal.WithLabelValues(SourceRaw, "info", StatusSuccess))

	if after-before != 1 {
		t.Errorf("counter delta = %v, want 1", after-before)
	}
}

func TestRecordPitch_NormalizesSeverity(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "unknown"},
		{"  ", "unknown"},
		{"INFO", "info"},
		{"Warning", "warning"},
		{"critical", "critical"},
		{" Success ", "success"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeSeverity(tt.input)
			if got != tt.want {
				t.Errorf("normalizeSeverity(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRecordPitch_DistinctLabelCombinations(t *testing.T) {
	beforeSuccess := testutil.ToFloat64(pitchesTotal.WithLabelValues(SourceGrafana, "warning", StatusSuccess))
	beforeError := testutil.ToFloat64(pitchesTotal.WithLabelValues(SourceGrafana, "warning", StatusError))

	RecordPitch(SourceGrafana, "warning", StatusSuccess)
	RecordPitch(SourceGrafana, "warning", StatusError)
	RecordPitch(SourceGrafana, "warning", StatusSuccess)

	afterSuccess := testutil.ToFloat64(pitchesTotal.WithLabelValues(SourceGrafana, "warning", StatusSuccess))
	afterError := testutil.ToFloat64(pitchesTotal.WithLabelValues(SourceGrafana, "warning", StatusError))

	if afterSuccess-beforeSuccess != 2 {
		t.Errorf("success delta = %v, want 2", afterSuccess-beforeSuccess)
	}
	if afterError-beforeError != 1 {
		t.Errorf("error delta = %v, want 1", afterError-beforeError)
	}
}

func TestObservePitchDuration_RecordsHistogram(t *testing.T) {
	before := testutil.CollectAndCount(pitchDuration, "omni_pitcher_pitch_duration_seconds")
	ObservePitchDuration(SourceGitHub, time.Now().Add(-50*time.Millisecond))
	after := testutil.CollectAndCount(pitchDuration, "omni_pitcher_pitch_duration_seconds")

	if after == before {
		t.Errorf("histogram count did not change (before=%d after=%d)", before, after)
	}
}

func TestSetBuildInfo_ExposesVersionAndCommit(t *testing.T) {
	SetBuildInfo("1.2.3", "abc1234")

	got := testutil.ToFloat64(buildInfo.WithLabelValues("1.2.3", "abc1234"))
	if got != 1 {
		t.Errorf("build_info gauge = %v, want 1", got)
	}
}

func TestExposition_ContainsExpectedFamilies(t *testing.T) {
	// Drive each collector once so they appear in the exposition.
	RecordPitch(SourceRaw, "info", StatusSuccess)
	ObservePitchDuration(SourceRaw, time.Now().Add(-10*time.Millisecond))
	SetBuildInfo("test", "deadbeef")

	// Hit promhttp.Handler() the same way Prometheus scrapes do.
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	promhttp.Handler().ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	got := string(body)

	for _, want := range []string{
		"omni_pitcher_pitches_total",
		"omni_pitcher_pitch_duration_seconds",
		"omni_pitcher_build_info",
		"go_goroutines", // free runtime metric from default registry
	} {
		if !strings.Contains(got, want) {
			t.Errorf("exposition missing %q", want)
		}
	}
}
