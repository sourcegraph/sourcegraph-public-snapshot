package git

import (
	"io"
	"io/fs"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git/gitapi"
)

// Mocks is used to mock behavior in tests. Tests must call ResetMocks() when finished to ensure its
// mocks are not (inadvertently) used by subsequent tests.
//
// (The emptyMocks is used by ResetMocks to zero out Mocks without needing to use a named type.)
var Mocks, emptyMocks struct {
	GetCommit        func(api.CommitID) (*gitapi.Commit, error)
	ExecSafe         func(params []string) (stdout, stderr []byte, exitCode int, err error)
	ExecReader       func(args []string) (reader io.ReadCloser, err error)
	RawLogDiffSearch func(opt RawLogDiffSearchOptions) ([]*LogCommitSearchResult, bool, error)
	NewFileReader    func(commit api.CommitID, name string) (io.ReadCloser, error)
	ReadFile         func(commit api.CommitID, name string) ([]byte, error)
	ReadDir          func(commit api.CommitID, name string, recurse bool) ([]fs.FileInfo, error)
	LsFiles          func(repo api.RepoName, commit api.CommitID) ([]string, error)
	ResolveRevision  func(spec string, opt ResolveRevisionOptions) (api.CommitID, error)
	Stat             func(commit api.CommitID, name string) (fs.FileInfo, error)
	Commits          func(repo api.RepoName, opt CommitsOptions) ([]*gitapi.Commit, error)
	MergeBase        func(repo api.RepoName, a, b api.CommitID) (api.CommitID, error)
	GetDefaultBranch func(repo api.RepoName) (refName string, commit api.CommitID, err error)
}

// ResetMocks clears the mock functions set on Mocks (so that subsequent tests don't inadvertently
// use them).
func ResetMocks() {
	Mocks = emptyMocks
}
