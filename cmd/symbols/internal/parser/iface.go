package parser

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type GitserverClient interface {
	// FetchTar returns an io.ReadCloser to a tar archive of a repository at the specified Git
	// remote URL and commit ID. If the error implements "BadRequest() bool", it will be used to
	// determine if the error is a bad request (eg invalid repo).
	FetchTar(context.Context, api.RepoName, api.CommitID, []string) (io.ReadCloser, error)

	// GitDiff returns the paths that have changed between two commits.
	GitDiff(context.Context, api.RepoName, api.CommitID, api.CommitID) (*Changes, error)
}

// Changes are added, deleted, and modified paths.
type Changes struct {
	Added    []string
	Modified []string
	Deleted  []string
}
