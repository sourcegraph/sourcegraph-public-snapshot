package ranking

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/background"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/background/coordinator"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/background/exporter"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/background/janitor"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/background/mapper"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/background/reducer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/store"
	codeintelshared "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
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

var (
	ExporterConfigInst    = &exporter.Config{}
	CoordinatorConfigInst = &coordinator.Config{}
	MapperConfigInst      = &mapper.Config{}
	ReducerConfigInst     = &reducer.Config{}
	JanitorConfigInst     = &janitor.Config{}
)

func NewSymbolExporter(observationCtx *observation.Context, rankingService *Service) goroutine.BackgroundRoutine {
	return background.NewSymbolExporter(
		scopedContext("exporter", observationCtx),
		rankingService.store,
		rankingService.lsifstore,
		ExporterConfigInst,
	)
}

func NewCoordinator(observationCtx *observation.Context, rankingService *Service) goroutine.BackgroundRoutine {
	return background.NewCoordinator(
		scopedContext("coordinator", observationCtx),
		rankingService.store,
		CoordinatorConfigInst,
	)
}

func NewMapper(observationCtx *observation.Context, rankingService *Service) []goroutine.BackgroundRoutine {
	return background.NewMapper(
		scopedContext("mapper", observationCtx),
		rankingService.store,
		MapperConfigInst,
	)
}

func NewReducer(observationCtx *observation.Context, rankingService *Service) goroutine.BackgroundRoutine {
	return background.NewReducer(
		scopedContext("reducer", observationCtx),
		rankingService.store,
		ReducerConfigInst,
	)
}

func NewSymbolJanitor(observationCtx *observation.Context, rankingService *Service) []goroutine.BackgroundRoutine {
	return background.NewSymbolJanitor(
		scopedContext("janitor", observationCtx),
		rankingService.store,
		JanitorConfigInst,
	)
}

func scopedContext(component string, observationCtx *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "ranking", component, observationCtx)
}
