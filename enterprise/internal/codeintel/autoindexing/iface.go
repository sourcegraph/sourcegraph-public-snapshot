package autoindexing

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/background"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type DependenciesService = background.DependenciesService

type PoliciesService = background.PoliciesService

type ReposStore = background.ReposStore

type GitserverRepoStore = background.GitserverRepoStore

type ExternalServiceStore = background.ExternalServiceStore

type AutoIndexingServiceForDepScheduling interface {
	QueueIndexesForPackage(ctx context.Context, pkg precise.Package) error
	InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (id int, err error)
}

type PolicyMatcher = background.PolicyMatcher

type RepoUpdaterClient = background.RepoUpdaterClient

type GitserverClient = background.GitserverClient

type InferenceService interface {
	InferIndexJobs(ctx context.Context, repo api.RepoName, commit, overrideScript string) ([]config.IndexJob, error)
	InferIndexJobHints(ctx context.Context, repo api.RepoName, commit, overrideScript string) ([]config.IndexJobHint, error)
}

type UploadService = background.UploadService
