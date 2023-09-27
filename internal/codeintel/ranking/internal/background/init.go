pbckbge bbckground

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/bbckground/coordinbtor"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/bbckground/exporter"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/bbckground/jbnitor"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/bbckground/mbpper"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/bbckground/reducer"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/lsifstore"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewSymbolExporter(observbtionCtx *observbtion.Context, store store.Store, lsifstore lsifstore.Store, config *exporter.Config) goroutine.BbckgroundRoutine {
	return exporter.NewSymbolExporter(observbtionCtx, store, lsifstore, config)
}

func NewCoordinbtor(observbtionCtx *observbtion.Context, store store.Store, config *coordinbtor.Config) goroutine.BbckgroundRoutine {
	return coordinbtor.NewCoordinbtor(observbtionCtx, store, config)
}

func NewMbpper(observbtionCtx *observbtion.Context, store store.Store, config *mbpper.Config) []goroutine.BbckgroundRoutine {
	return []goroutine.BbckgroundRoutine{
		mbpper.NewMbpper(observbtionCtx, store, config),
		mbpper.NewSeedMbpper(observbtionCtx, store, config),
	}
}

func NewReducer(observbtionCtx *observbtion.Context, store store.Store, config *reducer.Config) goroutine.BbckgroundRoutine {
	return reducer.NewReducer(observbtionCtx, store, config)
}

func NewSymbolJbnitor(observbtionCtx *observbtion.Context, store store.Store, config *jbnitor.Config) []goroutine.BbckgroundRoutine {
	return []goroutine.BbckgroundRoutine{
		jbnitor.NewExportedUplobdsJbnitor(observbtionCtx, store, config),
		jbnitor.NewDeletedUplobdsJbnitor(observbtionCtx, store, config),
		jbnitor.NewAbbndonedExportedUplobdsJbnitor(observbtionCtx, store, config),
		jbnitor.NewProcessedReferencesJbnitor(observbtionCtx, store, config),
		jbnitor.NewProcessedPbthsJbnitor(observbtionCtx, store, config),
		jbnitor.NewRbnkCountsJbnitor(observbtionCtx, store, config),
		jbnitor.NewRbnkJbnitor(observbtionCtx, store, config),
	}
}
