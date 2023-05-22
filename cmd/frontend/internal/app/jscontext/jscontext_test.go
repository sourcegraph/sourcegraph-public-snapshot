package jscontext

import (
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
