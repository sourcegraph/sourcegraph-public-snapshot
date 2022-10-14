package codeintel

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking"
	codeintelshared "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Services struct {
	AutoIndexingService *autoindexing.Service
	CodenavService      *codenav.Service
	DependenciesService *dependencies.Service
	PoliciesService     *policies.Service
	RankingService      *ranking.Service
	UploadsService      *uploads.Service
}

type Databases struct {
	DB          database.DB
	CodeIntelDB codeintelshared.CodeIntelDB
}

// GetServices creates or returns an already-initialized codeintel service collection.
// If the service collection is not yet initialized, a new one will be constructed using
// the given database handles.
func GetServices(dbs Databases) (Services, error) {
	return initServicesMemo.Init(dbs)
}

var initServicesMemo = memo.NewMemoizedConstructorWithArg(func(dbs Databases) (Services, error) {
	db, codeIntelDB := dbs.DB, dbs.CodeIntelDB
	gitserverClient := gitserver.New(db, scopedContext("gitserver"))

	uploadsSvc := uploads.GetService(db, codeIntelDB, gitserverClient)
	dependenciesSvc := dependencies.GetService(db, gitserverClient)
	policiesSvc := policies.GetService(db, uploadsSvc, gitserverClient)
	autoIndexingSvc := autoindexing.GetService(db, uploadsSvc, dependenciesSvc, policiesSvc, gitserverClient)
	codenavSvc := codenav.GetService(db, codeIntelDB, uploadsSvc, gitserverClient)
	rankingSvc := ranking.GetService(db, uploadsSvc, gitserverClient)

	return Services{
		AutoIndexingService: autoIndexingSvc,
		CodenavService:      codenavSvc,
		DependenciesService: dependenciesSvc,
		PoliciesService:     policiesSvc,
		RankingService:      rankingSvc,
		UploadsService:      uploadsSvc,
	}, nil
})

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "worker", component)
}
