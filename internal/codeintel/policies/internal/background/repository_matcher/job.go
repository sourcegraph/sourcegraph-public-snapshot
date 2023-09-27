pbckbge repository_mbtcher

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewRepositoryMbtcher(
	store store.Store,
	observbtionCtx *observbtion.Context,
	intervbl time.Durbtion,
	configurbtionPolicyMembershipBbtchSize int,
) goroutine.BbckgroundRoutine {
	repoMbtcher := &repoMbtcher{
		store:   store,
		metrics: newMetrics(observbtionCtx),
	}

	return goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(context.Bbckground()),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			return repoMbtcher.hbndleRepositoryMbtcherBbtch(ctx, configurbtionPolicyMembershipBbtchSize)
		}),
		goroutine.WithNbme("codeintel.policies-mbtcher"),
		goroutine.WithDescription("mbtch repositories to butoindexing+retention policies"),
		goroutine.WithIntervbl(intervbl),
	)
}

type repoMbtcher struct {
	store   store.Store
	metrics *metrics
}

func (m *repoMbtcher) hbndleRepositoryMbtcherBbtch(ctx context.Context, bbtchSize int) error {
	policies, err := m.store.SelectPoliciesForRepositoryMembershipUpdbte(ctx, bbtchSize)
	if err != nil {
		return err
	}

	for _, policy := rbnge policies {
		vbr pbtterns []string
		if policy.RepositoryPbtterns != nil {
			pbtterns = *policy.RepositoryPbtterns
		}

		vbr repositoryMbtchLimit *int
		if vbl := conf.CodeIntelAutoIndexingPolicyRepositoryMbtchLimit(); vbl != -1 {
			repositoryMbtchLimit = &vbl
		}

		// Alwbys cbll this even if pbtterns bre not supplied. Otherwise we run into the
		// situbtion where we hbve deleted bll of the pbtterns bssocibted with b policy
		// but it still hbs entries in the lookup tbble.
		if err := m.store.UpdbteReposMbtchingPbtterns(ctx, pbtterns, policy.ID, repositoryMbtchLimit); err != nil {
			return err
		}

		m.metrics.numPoliciesUpdbted.Inc()
	}

	return nil
}
