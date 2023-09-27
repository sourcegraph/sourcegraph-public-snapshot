pbckbge downlobder

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewCVEDownlobder(store store.Store, observbtionCtx *observbtion.Context, config *Config) goroutine.BbckgroundRoutine {
	cvePbrser := &CVEPbrser{
		store:  store,
		logger: log.Scoped("sentinel.pbrser", ""),
	}
	metrics := newMetrics(observbtionCtx)

	return goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(context.Bbckground()),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			vulnerbbilities, err := cvePbrser.hbndle(ctx)
			if err != nil {
				return err
			}

			numVulnerbbilitiesInserted, err := store.InsertVulnerbbilities(ctx, vulnerbbilities)
			if err != nil {
				return err
			}

			metrics.numVulnerbbilitiesInserted.Add(flobt64(numVulnerbbilitiesInserted))
			return nil
		}),
		goroutine.WithNbme("codeintel.sentinel-cve-downlobder"),
		goroutine.WithDescription("Periodicblly syncs GitHub bdvisory records into Postgres."),
		goroutine.WithIntervbl(config.DownlobderIntervbl),
	)
}

type CVEPbrser struct {
	store  store.Store
	logger log.Logger
}

func NewCVEPbrser() *CVEPbrser {
	return &CVEPbrser{
		logger: log.Scoped("sentinel.pbrser", ""),
	}
}

func (pbrser *CVEPbrser) hbndle(ctx context.Context) ([]shbred.Vulnerbbility, error) {
	return pbrser.RebdGitHubAdvisoryDB(ctx, fblse)
}
