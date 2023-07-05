package store

import (
	logger "github.com/sourcegraph/log"

	codeintelshared "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type ScipStore interface {
	// TODO
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
