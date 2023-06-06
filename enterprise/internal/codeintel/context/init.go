package context

import (
	scipstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/context/internal/scipstore"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/gosyntect"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewService(
	observationCtx *observation.Context,
	codeIntelDB codeintelshared.CodeIntelDB,
	syntectClient *gosyntect.Client,
	codenavSvc CodeNavService,
) *Service {
	return newService(
		observationCtx,
		scipstore.New(scopedContext("store", observationCtx), codeIntelDB),
		syntectClient,
		codenavSvc,
	)
}

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "context", component, parent)
}
