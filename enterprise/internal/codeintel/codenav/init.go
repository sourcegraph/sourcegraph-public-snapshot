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
) *Service {
	store := store.New(db, scopedContext("store"))
	lsifStore := lsifstore.New(codeIntelDB, scopedContext("lsifstore"))

	return newService(
		store,
		lsifStore,
		uploadSvc,
		gitserver,
		scopedContext("service"),
	)
}

type serviceDependencies struct {
	db          database.DB
	codeIntelDB codeintelshared.CodeIntelDB
	uploadSvc   UploadService
	gitserver   GitserverClient
}

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "codenav", component)
}
