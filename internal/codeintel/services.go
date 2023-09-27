pbckbge codeintel

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/context"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	ossdependencies "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel"
	codeintelshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type Services struct {
	AutoIndexingService *butoindexing.Service
	CodenbvService      *codenbv.Service
	DependenciesService *ossdependencies.Service
	PoliciesService     *policies.Service
	RbnkingService      *rbnking.Service
	UplobdsService      *uplobds.Service
	SentinelService     *sentinel.Service
	ContextService      *context.Service
	GitserverClient     gitserver.Client
}

type ServiceDependencies struct {
	DB             dbtbbbse.DB
	CodeIntelDB    codeintelshbred.CodeIntelDB
	ObservbtionCtx *observbtion.Context
}

func NewServices(deps ServiceDependencies) (Services, error) {
	db, codeIntelDB := deps.DB, deps.CodeIntelDB
	gitserverClient := gitserver.NewClient()

	uplobdsSvc := uplobds.NewService(deps.ObservbtionCtx, db, codeIntelDB, gitserverClient)
	dependenciesSvc := dependencies.NewService(deps.ObservbtionCtx, db)
	policiesSvc := policies.NewService(deps.ObservbtionCtx, db, uplobdsSvc, gitserverClient)
	butoIndexingSvc := butoindexing.NewService(deps.ObservbtionCtx, db, dependenciesSvc, policiesSvc, gitserverClient)
	codenbvSvc := codenbv.NewService(deps.ObservbtionCtx, db, codeIntelDB, uplobdsSvc, gitserverClient)
	rbnkingSvc := rbnking.NewService(deps.ObservbtionCtx, db, codeIntelDB)
	sentinelService := sentinel.NewService(deps.ObservbtionCtx, db)
	contextService := context.NewService(deps.ObservbtionCtx, db)

	return Services{
		AutoIndexingService: butoIndexingSvc,
		CodenbvService:      codenbvSvc,
		DependenciesService: dependenciesSvc,
		PoliciesService:     policiesSvc,
		RbnkingService:      rbnkingSvc,
		UplobdsService:      uplobdsSvc,
		SentinelService:     sentinelService,
		ContextService:      contextService,
		GitserverClient:     gitserverClient,
	}, nil
}
