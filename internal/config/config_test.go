package config

import (
	"testing"
	"time"
)

func TestParseRedisStartupTimeout(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    time.Duration
		wantErr bool
	}{
		{"unset uses default", "", DefaultRedisStartupTimeout, false},
		{"whitespace uses default", "  ", DefaultRedisStartupTimeout, false},
		{"seconds", "90s", 90 * time.Second, false},
		{"minutes", "5m", 5 * time.Minute, false},
		{"bare number is not a duration", "120", 0, true},
		{"garbage", "soon", 0, true},
		{"zero", "0s", 0, true},
		{"negative", "-10s", 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseRedisStartupTimeout(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseRedisStartupTimeout(%q) error = %v, wantErr %v", tc.in, err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("ParseRedisStartupTimeout(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
