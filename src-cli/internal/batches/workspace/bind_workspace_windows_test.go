package workspace

import (
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func mustHavePerm(t *testing.T, path string, want os.FileMode) error {
	t.Helper()

	have := mustGetPerm(t, path)

	// Go maps Windows file attributes onto Unix permissions in a fairly trivial
	// way: readonly files will be 0444, normal files will be 0666, and
	// directories will have 0111 or-ed onto that value. The end. Source:
	// https://sourcegraph.com/github.com/golang/go@fd841f65368906923e287afab91857043036459d/-/blob/src/os/types_windows.go#L112-134
	if want&0222 != 0 {
		want = 0666
	} else {
		want = 0444
	}
	if isDir(t, path) {
		want = want | 0111
	}

	if have != want {
		return errors.Errorf("unexpected permissions: have=%o want=%o", have, want)
	}
	return nil
}
