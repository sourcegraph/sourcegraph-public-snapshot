package autoindexing

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	codeinteltypes "github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type DependenciesService interface {
	UpsertDependencyRepos(ctx context.Context, deps []dependencies.Repo) ([]dependencies.Repo, error)
}

type PoliciesService interface {
	GetConfigurationPolicies(ctx context.Context, opts codeinteltypes.GetConfigurationPoliciesOptions) ([]codeinteltypes.ConfigurationPolicy, int, error)
}

type ReposStore interface {
	ListMinimalRepos(context.Context, database.ReposListOptions) ([]types.MinimalRepo, error)
}

type GitserverRepoStore interface {
	GetByNames(ctx context.Context, names ...api.RepoName) (map[api.RepoName]*types.GitserverRepo, error)
}

type ExternalServiceStore interface {
	Upsert(ctx context.Context, svcs ...*types.ExternalService) (err error)
	List(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error)
}

type AutoIndexingServiceForDepScheduling interface {
	QueueIndexesForPackage(ctx context.Context, pkg precise.Package) error
	InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (id int, err error)
}

type PolicyMatcher interface {
	CommitsDescribedByPolicyInternal(ctx context.Context, repositoryID int, policies []codeinteltypes.ConfigurationPolicy, now time.Time, filterCommits ...string) (map[string][]policies.PolicyMatch, error)
}

type AutoIndexingServiceForDepSchedulingShim struct {
	*Service
}

func (s *AutoIndexingServiceForDepSchedulingShim) QueueIndexesForPackage(ctx context.Context, pkg precise.Package) error {
	return s.Service.queueIndexesForPackage(ctx, pkg)
}

func (s *AutoIndexingServiceForDepSchedulingShim) InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (id int, err error) {
	return s.Service.insertDependencyIndexingJob(ctx, uploadID, externalServiceKind, syncTime)
}
