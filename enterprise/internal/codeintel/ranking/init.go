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
) *Service {
	return newService(
		scopedContext("service", observationCtx),
		store.New(scopedContext("store", observationCtx), db),
		lsifstore.New(scopedContext("lsifstore", observationCtx), codeIntelDB),
		conf.DefaultClient(),
	)
}

func NewSymbolExporter(observationCtx *observation.Context, rankingService *Service) goroutine.BackgroundRoutine {
	return background.NewSymbolExporter(
		observationCtx,
		rankingService.store,
		rankingService.lsifstore,
		ConfigInst.SymbolExporterInterval,
		ConfigInst.SymbolExporterReadBatchSize,
		ConfigInst.SymbolExporterWriteBatchSize,
	)
}

func NewSymbolJanitor(observationCtx *observation.Context, rankingService *Service) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewSymbolDefinitionsJanitor(
			observationCtx,
			rankingService.store,
			ConfigInst.SymbolExporterInterval,
		),
		background.NewSymbolReferencesJanitor(
			observationCtx,
			rankingService.store,
			ConfigInst.SymbolExporterInterval,
		),
		background.NewSymbolInitialPathsJanitor(
			observationCtx,
			rankingService.store,
			ConfigInst.SymbolExporterInterval,
		),
		background.NewRankCountsJanitor(
			observationCtx,
			rankingService.store,
			ConfigInst.SymbolExporterInterval,
		),
		background.NewRankJanitor(
			observationCtx,
			rankingService.store,
			ConfigInst.SymbolExporterInterval,
		),
	}
}

func NewMapper(observationCtx *observation.Context, rankingService *Service) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewMapper(
			observationCtx,
			rankingService.store,
			ConfigInst.SymbolExporterInterval,
			ConfigInst.MapperBatchSize,
		),
		background.NewSeedMapper(
			observationCtx,
			rankingService.store,
			ConfigInst.SymbolExporterInterval,
			ConfigInst.MapperBatchSize,
		),
	}
}

func NewReducer(observationCtx *observation.Context, rankingService *Service) goroutine.BackgroundRoutine {
	return background.NewReducer(
		observationCtx,
		rankingService.store,
		ConfigInst.SymbolExporterInterval,
		ConfigInst.ReducerBatchSize,
	)
}

func scopedContext(component string, observationCtx *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "ranking", component, observationCtx)
}
