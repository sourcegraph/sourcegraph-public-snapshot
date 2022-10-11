package uploads

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	policiesEnterprise "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetService creates or returns an already-initialized uploads service.
// If the service is not yet initialized, it will use the provided dependencies.
func GetService(
	db database.DB,
	codeIntelDB stores.CodeIntelDB,
	gsc GitserverClient,
) *Service {
	svc, _ := initServiceMemo.Init(serviceDependencies{
		db,
		codeIntelDB,
		gsc,
	})

	return svc
}

type serviceDependencies struct {
	db          database.DB
	codeIntelDB stores.CodeIntelDB
	gsc         GitserverClient
}

var initServiceMemo = memo.NewMemoizedConstructorWithArg(func(deps serviceDependencies) (*Service, error) {
	store := store.New(deps.db, scopedContext("store"))
	repoStore := backend.NewRepos(scopedContext("repos").Logger, deps.db, gitserver.NewClient(deps.db))
	lsifStore := lsifstore.New(deps.codeIntelDB, scopedContext("lsifstore"))
	policyMatcher := policiesEnterprise.NewMatcher(deps.gsc, policiesEnterprise.RetentionExtractor, true, false)
	locker := locker.NewWith(deps.db, "codeintel")

	svc := newService(
		store,
		repoStore,
		lsifStore,
		deps.gsc,
		nil, // written in circular fashion
		nil, // written in circular fashion
		policyMatcher,
		locker,
		scopedContext("service"),
	)

	dependenciesSvc := dependencies.GetService(deps.db, deps.gsc)
	policySvc := policies.GetService(deps.db, svc, deps.gsc)
	autoIndexingSvc := autoindexing.GetService(deps.db, svc, dependenciesSvc, policySvc, deps.gsc)

	svc.policySvc = policySvc
	svc.autoIndexingSvc = autoIndexingSvc
	return svc, nil
})

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "uploads", component)
}
