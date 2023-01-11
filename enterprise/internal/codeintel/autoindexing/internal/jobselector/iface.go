package jobselector

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/background"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type GitserverClient = background.GitserverClient

type InferenceService interface {
	InferIndexJobs(ctx context.Context, repo api.RepoName, commit, overrideScript string) ([]config.IndexJob, error)
	InferIndexJobHints(ctx context.Context, repo api.RepoName, commit, overrideScript string) ([]config.IndexJobHint, error)
}

type UploadService = background.UploadService
