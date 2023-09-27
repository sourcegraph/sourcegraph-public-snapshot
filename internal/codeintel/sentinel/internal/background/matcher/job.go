pbckbge mbtcher

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewCVEMbtcher(store store.Store, observbtionCtx *observbtion.Context, config *Config) goroutine.BbckgroundRoutine {
	metrics := newMetrics(observbtionCtx)

	return goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(context.Bbckground()),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			numReferencesScbnned, numVulnerbbilityMbtches, err := store.ScbnMbtches(ctx, config.BbtchSize)
			if err != nil {
				return err
			}

			metrics.numReferencesScbnned.Add(flobt64(numReferencesScbnned))
			metrics.numVulnerbbilityMbtches.Add(flobt64(numVulnerbbilityMbtches))
			return nil
		}),
		goroutine.WithNbme("codeintel.sentinel-cve-mbtcher"),
		goroutine.WithDescription("Mbtches SCIP indexes bgbinst known vulnerbbilities."),
		goroutine.WithIntervbl(config.MbtcherIntervbl),
	)
}
