package gitdomain

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
)

func TestIsAllowedGitCmd(t *testing.T) {
	allowed := [][]string{
		// Required for code monitors
		{"rev-parse", "HEAD"},
		{"rev-parse", "83838383"},
		{"rev-parse", "--glob=refs/heads/*"},
		{"rev-parse", "--glob=refs/heads/*", "--exclude=refs/heads/cc/*"},
	}

	logger := logtest.Scoped(t)
	for _, args := range allowed {
		t.Run("", func(t *testing.T) {
			if !IsAllowedGitCmd(logger, args) {
				t.Fatalf("expected args to be allowed: %q", args)
			}
		})
	}
}
