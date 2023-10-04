package gitserverfs

import (
	"path/filepath"
	"testing"

	"gotest.tools/assert"
)

func TestIgnorePath(t *testing.T) {
	reposDir := "/data/repos"

	for _, tc := range []struct {
		path         string
		shouldIgnore bool
	}{
		{path: filepath.Join(reposDir, TempDirName), shouldIgnore: true},
		{path: filepath.Join(reposDir, P4HomeName), shouldIgnore: true},
		// Double check handling of trailing space
		{path: filepath.Join(reposDir, P4HomeName+"   "), shouldIgnore: true},
		{path: filepath.Join(reposDir, "sourcegraph/sourcegraph"), shouldIgnore: false},
	} {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tc.shouldIgnore, IgnorePath(reposDir, tc.path))
		})
	}
}
