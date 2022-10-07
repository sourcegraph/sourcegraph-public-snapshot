package codenav

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/internal/store"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewService(
	db database.DB,
	codeIntelDB codeintelshared.CodeIntelDB,
	uploadSvc UploadService,
	gitserver GitserverClient,
	observationContext *observation.Context,
) *Service {
	store := store.New(db, scopedContext("store", observationContext))
	lsifStore := lsifstore.New(codeIntelDB, scopedContext("lsifstore", observationContext))

	return newService(
		store,
		lsifStore,
		uploadSvc,
		gitserver,
		observationContext,
	)
}

type serviceDependencies struct {
	db                 database.DB
	codeIntelDB        codeintelshared.CodeIntelDB
	uploadSvc          UploadService
	gitserver          GitserverClient
	observationContext *observation.Context
}

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "codenav", component, parent)
}
