package sqlite

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type GitserverClient interface {
	// GitDiff returns the paths that have changed between two commits.
	GitDiff(context.Context, api.RepoName, api.CommitID, api.CommitID) (*Changes, error)
}

// Changes are added, deleted, and modified paths.
type Changes struct {
	Added    []string
	Modified []string
	Deleted  []string
}
