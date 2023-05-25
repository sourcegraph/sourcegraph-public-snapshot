package context

import (
	scipstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/context/internal/scipstore"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewService(
	observationCtx *observation.Context,
	codeIntelDB codeintelshared.CodeIntelDB,
) *Service {
	scipStore := scipstore.New(scopedContext("store", observationCtx), codeIntelDB)

	return newService(
		observationCtx,
		scipStore,
	)
}

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "context", component, parent)
}
