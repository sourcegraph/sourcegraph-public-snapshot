package background

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/background/exporter"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/background/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/background/mapper"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/background/reducer"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewSymbolExporter(observationCtx *observation.Context, store store.Store, lsifstore lsifstore.Store, config *exporter.Config) goroutine.BackgroundRoutine {
	return exporter.NewSymbolExporter(observationCtx, store, lsifstore, config)
}

func NewMapper(observationCtx *observation.Context, store store.Store, config *mapper.Config) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		mapper.NewMapper(observationCtx, store, config),
		mapper.NewSeedMapper(observationCtx, store, config),
	}
}

func NewReducer(observationCtx *observation.Context, store store.Store, config *reducer.Config) goroutine.BackgroundRoutine {
	return reducer.NewReducer(observationCtx, store, config)
}

func NewSymbolJanitor(observationCtx *observation.Context, store store.Store, config *janitor.Config) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		janitor.NewSymbolDefinitionsJanitor(observationCtx, store, config),
		janitor.NewSymbolReferencesJanitor(observationCtx, store, config),
		janitor.NewSymbolInitialPathsJanitor(observationCtx, store, config),
		janitor.NewAbandonedDefinitionsJanitor(observationCtx, store, config),
		janitor.NewAbandonedReferencesJanitor(observationCtx, store, config),
		janitor.NewAbandonedInitialCountsJanitor(observationCtx, store, config),
		janitor.NewRankCountsJanitor(observationCtx, store, config),
		janitor.NewRankJanitor(observationCtx, store, config),
	}
}
