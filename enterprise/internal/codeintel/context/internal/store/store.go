package store

import (
	"context"

	logger "github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store interface {
	GetSCIPDocumentsByFuzzySelector(ctx context.Context, selector, scipNameType string) (documents []*scip.Document, err error)
	GetSCIPDocumentsByPreciseSelector(ctx context.Context, uploadID int, schemeID, packageManagerID, packageManagerName, packageVersionID, descriptorID int) (documents []*scip.Document, err error)
}

type store struct {
	db         *basestore.Store
	logger     logger.Logger
	operations *operations
}

func New(observationCtx *observation.Context, db database.DB) Store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		logger:     logger.Scoped("context.store", ""),
		operations: newOperations(observationCtx),
	}
}
