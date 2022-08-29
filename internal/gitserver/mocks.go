package gitserver

import (
	"io"
	"io/fs"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

// Mocks is used to mock behavior in tests. Tests must call ResetMocks() when finished to ensure its
// mocks are not (inadvertently) used by subsequent tests.
//
// (The emptyMocks is used by ResetMocks to zero out Mocks without needing to use a named type.)
//
// NOTE: These are temporary and are being copied over from the vcs/git package
// while we move most of that functionality onto our gitserver client. Once
// that's done, we should take advantage of the generated mock client in this
// package instead.
var Mocks, emptyMocks struct {
	ExecReader       func(args []string) (reader io.ReadCloser, err error)
	ReadDir          func(commit api.CommitID, name string, recurse bool) ([]fs.FileInfo, error)
	ResolveRevision  func(spec string, opt ResolveRevisionOptions) (api.CommitID, error)
	LsFiles          func(repo api.RepoName, commit api.CommitID) ([]string, error)
	MergeBase        func(repo api.RepoName, a, b api.CommitID) (api.CommitID, error)
	GetDefaultBranch func(repo api.RepoName) (refName string, commit api.CommitID, err error)
	ReadFile         func(commit api.CommitID, name string) ([]byte, error)
	NewFileReader    func(commit api.CommitID, name string) (io.ReadCloser, error)
	Stat             func(commit api.CommitID, name string) (fs.FileInfo, error)
	GetCommit        func(api.CommitID) (*gitdomain.Commit, error)
	Commits          func(repo api.RepoName, opt CommitsOptions) ([]*gitdomain.Commit, error)
}

// ResetMocks clears the mock functions set on Mocks (so that subsequent tests don't inadvertently
// use them).
func ResetMocks() {
	Mocks = emptyMocks
}
