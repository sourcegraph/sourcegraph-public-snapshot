package autoindexing

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"

	policiesEnterprise "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/symbols"
)

var (
	svc     *Service
	svcOnce sync.Once
)

var maximumIndexJobsPerInferredConfiguration = env.MustGetInt("PRECISE_CODE_INTEL_AUTO_INDEX_MAXIMUM_INDEX_JOBS_PER_INFERRED_CONFIGURATION", 25, "Repositories with a number of inferred auto-index jobs exceeding this threshold will not be auto-indexed.")

// GetService creates or returns an already-initialized autoindexing service.
// If the service is not yet initialized, it will use the provided dependencies.
func GetService(
	db database.DB,
	uploadSvc shared.UploadService,
	depsSvc DependenciesService,
	policiesSvc PoliciesService,
	gitserver shared.GitserverClient,
	repoUpdater shared.RepoUpdaterClient,
) *Service {
	svcOnce.Do(func() {
		store := store.New(db, scopedContext("store"))
		repoStore := db.Repos()
		gitserverRepoStore := db.GitserverRepos()
		externalServiceStore := db.ExternalServices()
		policyMatcher := policiesEnterprise.NewMatcher(gitserver, policiesEnterprise.IndexingExtractor, false, true)
		symbolsClient := symbols.DefaultClient
		inferenceSvc := inference.GetService(db)

		svc = newService(
			store,
			uploadSvc,
			depsSvc,
			policiesSvc,
			repoStore,
			gitserverRepoStore,
			externalServiceStore,
			policyMatcher,
			gitserver,
			symbolsClient,
			repoUpdater,
			inferenceSvc,
			scopedContext("service"),
		)
	})

	return svc
}

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "autoindexing", component)
}
