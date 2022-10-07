package codeintel

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads"
	ossdependencies "github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Services struct {
	AutoIndexingService *autoindexing.Service
	CodenavService      *codenav.Service
	DependenciesService *ossdependencies.Service
	PoliciesService     *policies.Service
	RankingService      *ranking.Service
	UploadsService      *uploads.Service
}

type ServiceDependencies struct {
	DB                 database.DB
	CodeIntelDB        codeintelshared.CodeIntelDB
	GitserverClient    *gitserver.Client
	ObservationContext *observation.Context
}

func NewServices(deps ServiceDependencies) (Services, error) {
	db, codeIntelDB := deps.DB, deps.CodeIntelDB
	gitserverClient := gitserver.New(db, scopedContext("gitserver", deps.ObservationContext))

	uploadsSvc := uploads.NewService(db, codeIntelDB, gitserverClient, deps.ObservationContext)
	dependenciesSvc := dependencies.NewService(db, deps.ObservationContext)
	policiesSvc := policies.NewService(db, uploadsSvc, gitserverClient, deps.ObservationContext)
	autoIndexingSvc := autoindexing.NewService(db, uploadsSvc, dependenciesSvc, policiesSvc, gitserverClient, deps.ObservationContext)
	codenavSvc := codenav.NewService(db, codeIntelDB, uploadsSvc, gitserverClient, deps.ObservationContext)
	rankingSvc := ranking.NewService(db, uploadsSvc, gitserverClient, deps.ObservationContext)

	return Services{
		AutoIndexingService: autoIndexingSvc,
		CodenavService:      codenavSvc,
		DependenciesService: dependenciesSvc,
		PoliciesService:     policiesSvc,
		RankingService:      rankingSvc,
		UploadsService:      uploadsSvc,
	}, nil
}

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "worker", component, parent)
}
