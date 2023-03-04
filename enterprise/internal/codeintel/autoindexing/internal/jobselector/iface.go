package jobselector

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type GitserverClient interface {
	FileExists(ctx context.Context, repositoryID int, commit, file string) (bool, error)
	RawContents(ctx context.Context, repositoryID int, commit, file string) ([]byte, error)
}

type InferenceService interface {
	InferIndexJobs(ctx context.Context, repo api.RepoName, commit, overrideScript string) ([]config.IndexJob, error)
	InferIndexJobHints(ctx context.Context, repo api.RepoName, commit, overrideScript string) ([]config.IndexJobHint, error)
}
