package context

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/context/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewService(
	observationCtx *observation.Context,
	db database.DB,
) *Service {
	store := store.New(scopedContext("store", observationCtx), db)

	return newService(
		observationCtx,
		store,
	)
}

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "context", component, parent)
}
