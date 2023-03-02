package ranking

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/store"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewService(
	observationCtx *observation.Context,
	db database.DB,
	codeIntelDB codeintelshared.CodeIntelDB,
	gitserverClient GitserverClient,
) *Service {
	return newService(
		scopedContext("service", observationCtx),
		store.New(scopedContext("store", observationCtx), db),
		lsifstore.New(scopedContext("lsifstore", observationCtx), codeIntelDB),
		gitserverClient,
		conf.DefaultClient(),
	)
}

func NewSymbolExporter(observationCtx *observation.Context, rankingService *Service) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewSymbolExporter(
			observationCtx,
			rankingService,
			ConfigInst.SymbolExporterNumRoutines,
			ConfigInst.SymbolExporterInterval,
			ConfigInst.SymbolExporterWriteBatchSize,
			ConfigInst.DocumentReferenceCountsEnabled,
		),
	}
}

func NewMapper(observationCtx *observation.Context, rankingService *Service) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewMapper(
			observationCtx,
			rankingService,
			ConfigInst.SymbolExporterInterval,
			ConfigInst.DocumentReferenceCountsEnabled,
		),
	}
}

func NewReducer(observationCtx *observation.Context, rankingService *Service) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewReducer(
			observationCtx,
			rankingService,
			ConfigInst.SymbolExporterInterval,
			ConfigInst.DocumentReferenceCountsEnabled,
		),
	}
}

func scopedContext(component string, observationCtx *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "ranking", component, observationCtx)
}
