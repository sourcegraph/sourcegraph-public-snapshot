package codeintel

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database"
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

func NewServices(deps Databases) (Services, error) {
	db, codeIntelDB := deps.DB, deps.CodeIntelDB
	gitserverClient := gitserver.New(db, scopedContext("gitserver"))

	uploadsSvc := uploads.NewService(db, codeIntelDB, gitserverClient)
	dependenciesSvc := dependencies.NewService(db)
	policiesSvc := policies.NewService(db, uploadsSvc, gitserverClient)
	autoIndexingSvc := autoindexing.NewService(db, uploadsSvc, dependenciesSvc, policiesSvc, gitserverClient)
	codenavSvc := codenav.NewService(db, codeIntelDB, uploadsSvc, gitserverClient)
	rankingSvc := ranking.NewService(db, uploadsSvc, gitserverClient)

	return Services{
		AutoIndexingService: autoIndexingSvc,
		CodenavService:      codenavSvc,
		DependenciesService: dependenciesSvc,
		PoliciesService:     policiesSvc,
		RankingService:      rankingSvc,
		UploadsService:      uploadsSvc,
	}, nil
}

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "worker", component)
}
