package codenav

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/internal/lsifstore"
	codenavstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/internal/store"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewService(
	observationCtx *observation.Context,
	db database.DB,
	codeIntelDB codeintelshared.CodeIntelDB,
	uploadSvc UploadService,
	gitserver gitserver.Client,
) *Service {
	store := codenavstore.New(scopedContext("store", observationCtx), db)
	lsifStore := lsifstore.New(scopedContext("lsifstore", observationCtx), codeIntelDB)

	return newService(
		observationCtx,
		store,
		db.Repos(),
		lsifStore,
		uploadSvc,
		gitserver,
	)
}

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "codenav", component, parent)
}
