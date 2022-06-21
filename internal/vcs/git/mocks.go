package git

import (
	"io"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// Mocks is used to mock behavior in tests. Tests must call ResetMocks() when finished to ensure its
// mocks are not (inadvertently) used by subsequent tests.
//
// (The emptyMocks is used by ResetMocks to zero out Mocks without needing to use a named type.)
var Mocks, emptyMocks struct {
	ExecReader func(args []string) (reader io.ReadCloser, err error)
	MergeBase  func(repo api.RepoName, a, b api.CommitID) (api.CommitID, error)
}

// ResetMocks clears the mock functions set on Mocks (so that subsequent tests don't inadvertently
// use them).
func ResetMocks() {
	Mocks = emptyMocks
}
