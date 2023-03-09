package codeintel

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/search"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads"
	ossdependencies "github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Services struct {
	AutoIndexingService *autoindexing.Service
	CodenavService      *codenav.Service
	DependenciesService *ossdependencies.Service
	PoliciesService     *policies.Service
	RankingService      *ranking.Service
	UploadsService      *uploads.Service
	SentinelService     *sentinel.Service
	SearchService       *search.Service
}

type ServiceDependencies struct {
	DB              database.DB
	CodeIntelDB     codeintelshared.CodeIntelDB
	GitserverClient *gitserver.Client
	ObservationCtx  *observation.Context
}

func NewServices(deps ServiceDependencies) (Services, error) {
	db, codeIntelDB := deps.DB, deps.CodeIntelDB
	gitserverClient := gitserver.New(scopedContext("gitserver", deps.ObservationCtx), db)

	uploadsSvc := uploads.NewService(deps.ObservationCtx, db, codeIntelDB, gitserverClient)
	dependenciesSvc := dependencies.NewService(deps.ObservationCtx, db)
	policiesSvc := policies.NewService(deps.ObservationCtx, db, uploadsSvc, gitserverClient)
	autoIndexingSvc := autoindexing.NewService(deps.ObservationCtx, db, dependenciesSvc, policiesSvc, gitserverClient)
	codenavSvc := codenav.NewService(deps.ObservationCtx, db, codeIntelDB, uploadsSvc, gitserverClient)
	rankingSvc := ranking.NewService(deps.ObservationCtx, db, codeIntelDB)
	sentinelService := sentinel.NewService(deps.ObservationCtx, db)

	searchMaxIndexes := 500 // default value of PRECISE_CODE_INTEL_MAXIMUM_INDEXES_PER_MONIKER_SEARCH
	searchService, err := search.NewService(deps.ObservationCtx, gitserverClient, codenavSvc, searchMaxIndexes)
	if err != nil {
		return Services{}, errors.Wrap(err, "search.NewService")
	}

	return Services{
		AutoIndexingService: autoIndexingSvc,
		CodenavService:      codenavSvc,
		DependenciesService: dependenciesSvc,
		PoliciesService:     policiesSvc,
		RankingService:      rankingSvc,
		UploadsService:      uploadsSvc,
		SentinelService:     sentinelService,
		SearchService:       searchService,
	}, nil
}

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "worker", component, parent)
}
