package jscontext

import (
	"runtime"
	"testing"
)

func TestIsBot(t *testing.T) {
	tests := map[string]bool{
		"my bot":     true,
		"my Bot foo": true,
		"Chrome":     false,
	}
	for userAgent, want := range tests {
		got := isBot(userAgent)
		if got != want {
			t.Errorf("%q: want %v, got %v", userAgent, got, want)
		}
	}
}

func Test_likelyDockerOnMac(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.SkipNow()
	}

	tests := []struct {
		name string
		want bool
	}{
		{"base",
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := likelyDockerOnMac(); got != tt.want {
				t.Errorf("likelyDockerOnMac() = %v, want %v", got, tt.want)
			}
		})
	}
}
