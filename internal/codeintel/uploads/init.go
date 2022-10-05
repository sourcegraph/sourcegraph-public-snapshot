package uploads

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	policiesEnterprise "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

var (
	svc     *Service
	svcOnce sync.Once
)

type RepoUpdaterClient interface {
	EnqueueRepoUpdate(ctx context.Context, repo api.RepoName) (*protocol.RepoUpdateResponse, error)
}

// GetService creates or returns an already-initialized uploads service.
// If the service is not yet initialized, it will use the provided dependencies.
func GetService(
	db database.DB,
	codeIntelDB stores.CodeIntelDB,
	gsc GitserverClient,
) *Service {
	svcOnce.Do(func() {
		store := store.New(db, scopedContext("store"))
		repoStore := backend.NewRepos(scopedContext("repos").Logger, db, gitserver.NewClient(db))
		lsifStore := lsifstore.New(codeIntelDB, scopedContext("lsifstore"))
		policyMatcher := policiesEnterprise.NewMatcher(gsc, policiesEnterprise.RetentionExtractor, true, false)
		locker := locker.NewWith(db, "codeintel")
		repoUpdater := repoupdater.New(&observation.TestContext)

		svc = newService(
			store,
			repoStore,
			lsifStore,
			gsc,
			nil, // written in circular fashion
			nil, // written in circular fashion
			policyMatcher,
			locker,
			scopedContext("service"),
		)
		svc.policySvc = policies.GetService(db, svc, gsc)
		svc.autoIndexingSvc = autoindexing.GetService(db, svc, dependencies.GetService(db, gsc), svc.policySvc, gsc, repoUpdater)
	})

	return svc
}

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "uploads", component)
}
