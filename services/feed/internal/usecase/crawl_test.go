package usecase

import (
	"testing"
	"time"
)

func TestParseInterval(t *testing.T) {
	cases := map[string]time.Duration{
		"":     15 * time.Minute,
		"30m":  30 * time.Minute,
		"0":    0,
		"off":  0,
		"bad":  15 * time.Minute,
		"-1s":  15 * time.Minute,
	}
	for raw, want := range cases {
		if got := ParseInterval(raw); got != want {
			t.Fatalf("%q: got %s want %s", raw, got, want)
		}
	}
}
