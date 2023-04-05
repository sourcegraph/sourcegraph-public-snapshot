package timeutil

import (
	"testing"
	"time"

	"gotest.tools/assert"
)

func TestDuration_FormatDuration(t *testing.T) {
	testCases := []struct {
		duration time.Duration
		want     string
	}{
		{duration: 24 * time.Hour, want: "1d"},
		{duration: 24*time.Hour + 13*time.Hour, want: "1d 13h"},
		{duration: 24*time.Hour + 13*time.Hour + 55*time.Minute, want: "1d 13h 55m"},
		{duration: 24*time.Hour + 13*time.Hour + 55*time.Minute + 7*time.Minute, want: "1d 14h 2m"},
		{duration: 24*time.Hour + 13*time.Hour + 55*time.Minute + 45*time.Second, want: "1d 13h 55m 45s"},
		{duration: 24*time.Hour + 13*time.Hour + 55*time.Minute + 45*time.Second + 45*time.Second, want: "1d 13h 56m 30s"},
		{duration: 45 * time.Second, want: "45s"},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.want, FormatDuration(tc.duration), tc.duration)
	}
}
