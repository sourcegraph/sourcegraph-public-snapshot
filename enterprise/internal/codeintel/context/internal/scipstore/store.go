package store

import (
	"context"

	logger "github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"

	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type ScipStore interface {
	GetSCIPDocumentsByFuzzySelector(ctx context.Context, selector, scipNameType string) (documents []*scip.Document, err error)
	GetSCIPDocumentsByPreciseSelector(ctx context.Context, uploadID int, schemeID, packageManagerID, packageManagerName, packageVersionID, descriptorID int) (documents []*scip.Document, err error)
}

type store struct {
	db         *basestore.Store
	logger     logger.Logger
	operations *operations
}

func New(observationCtx *observation.Context, db codeintelshared.CodeIntelDB) ScipStore {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		logger:     logger.Scoped("context.scip_store", ""),
		operations: newOperations(observationCtx),
	}
}
