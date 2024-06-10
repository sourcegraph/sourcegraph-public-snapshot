package codenav

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore"
	codeintelshared "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	searchClient "github.com/sourcegraph/sourcegraph/internal/search/client"
)

func NewService(
	observationCtx *observation.Context,
	db database.DB,
	codeIntelDB codeintelshared.CodeIntelDB,
	uploadSvc UploadService,
	gitserver gitserver.Client,
) *Service {
	lsifStore := lsifstore.New(scopedContext("lsifstore", observationCtx), codeIntelDB)
	logger := log.Scoped("codenav")
	searcher := searchClient.New(logger, db, gitserver)

	return newService(
		observationCtx,
		db.Repos(),
		lsifStore,
		uploadSvc,
		gitserver,
		searcher,
		logger,
	)
}

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "codenav", component, parent)
}
