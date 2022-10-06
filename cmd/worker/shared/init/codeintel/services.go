package codeintel

import (
	"github.com/sourcegraph/sourcegraph/cmd/worker/memo"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Services struct {
	AutoIndexingService *autoindexing.Service
	CodenavService      *codenav.Service
	DependenciesService *dependencies.Service
	PoliciesService     *policies.Service
	UploadsService      *uploads.Service
}

// InitServices initializes and returns code intelligence services.
func InitServices() (*Services, error) {
	return initServicesMemo.Init()
}

var initServicesMemo = memo.NewMemoizedConstructor(func() (*Services, error) {
	db, err := workerdb.InitDBWithLogger(scopedContext("db").Logger)
	if err != nil {
		return nil, err
	}

	codeIntelDB, err := InitDBWithLogger(scopedContext("codeintel-db").Logger)
	if err != nil {
		return nil, err
	}

	gitserverClient := gitserver.New(db, scopedContext("gitserver"))
	repoUpdaterClient := repoupdater.New(scopedContext("repo-updater"))

	uploadsSvc := uploads.GetService(db, codeIntelDB, gitserverClient)
	codenavSvc := codenav.GetService(db, codeIntelDB, uploadsSvc, gitserverClient)
	dependenciesSvc := dependencies.GetService(db, gitserverClient)
	policiesSvc := policies.GetService(db, uploadsSvc, gitserverClient)
	autoIndexingSvc := autoindexing.GetService(db, uploadsSvc, dependenciesSvc, policiesSvc, gitserverClient, repoUpdaterClient)

	return &Services{
		AutoIndexingService: autoIndexingSvc,
		CodenavService:      codenavSvc,
		DependenciesService: dependenciesSvc,
		PoliciesService:     policiesSvc,
		UploadsService:      uploadsSvc,
	}, nil
})

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "worker", component)
}
