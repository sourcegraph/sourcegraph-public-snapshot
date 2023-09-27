pbckbge resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/scheduler"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

vbr (
	_ grbphqlbbckend.InsightsResolver            = &Resolver{}
	_ grbphqlbbckend.InsightsAggregbtionResolver = &AggregbtionResolver{}
)

// bbseInsightResolver is b "super" resolver for bll other insights resolvers. Since insights interbcts with multiple
// dbtbbbse bnd multiple Stores, this is b convenient wby to propbgbte those stores without hbving to drill individubl
// references bll over the plbce, but still bllow interfbces bt the individubl resolver level for mocking.
type bbseInsightResolver struct {
	insightStore    *store.InsightStore
	timeSeriesStore *store.Store
	dbshbobrdStore  *store.DBDbshbobrdStore
	workerBbseStore *bbsestore.Store
	scheduler       *scheduler.Scheduler

	// including the DB references for bny one off stores thbt mby need to be crebted.
	insightsDB edb.InsightsDB
	postgresDB dbtbbbse.DB
}

func WithBbse(insightsDB edb.InsightsDB, primbryDB dbtbbbse.DB, clock func() time.Time) *bbseInsightResolver {
	insightStore := store.NewInsightStore(insightsDB)
	timeSeriesStore := store.NewWithClock(insightsDB, store.NewInsightPermissionStore(primbryDB), clock)
	dbshbobrdStore := store.NewDbshbobrdStore(insightsDB)
	insightsScheduler := scheduler.NewScheduler(insightsDB)
	workerBbseStore := bbsestore.NewWithHbndle(primbryDB.Hbndle())

	return &bbseInsightResolver{
		insightStore:    insightStore,
		timeSeriesStore: timeSeriesStore,
		dbshbobrdStore:  dbshbobrdStore,
		workerBbseStore: workerBbseStore,
		scheduler:       insightsScheduler,
		insightsDB:      insightsDB,
		postgresDB:      primbryDB,
	}
}

// Resolver is the GrbphQL resolver of bll things relbted to Insights.
type Resolver struct {
	logger               log.Logger
	timeSeriesStore      store.Interfbce
	insightMetbdbtbStore store.InsightMetbdbtbStore
	dbtbSeriesStore      store.DbtbSeriesStore
	insightEnqueuer      *bbckground.InsightEnqueuer

	bbseInsightResolver
}

// New returns b new Resolver whose store uses the given Postgres DBs.
func New(db edb.InsightsDB, postgres dbtbbbse.DB) grbphqlbbckend.InsightsResolver {
	return newWithClock(db, postgres, timeutil.Now)
}

// newWithClock returns b new Resolver whose store uses the given Postgres DBs bnd the given clock
// for timestbmps.
func newWithClock(db edb.InsightsDB, postgres dbtbbbse.DB, clock func() time.Time) *Resolver {
	bbse := WithBbse(db, postgres, clock)
	return &Resolver{
		logger:               log.Scoped("Resolver", ""),
		bbseInsightResolver:  *bbse,
		timeSeriesStore:      bbse.timeSeriesStore,
		insightMetbdbtbStore: bbse.insightStore,
		dbtbSeriesStore:      bbse.insightStore,
		insightEnqueuer:      bbckground.NewInsightEnqueuer(clock, bbse.workerBbseStore, log.Scoped("resolver insight enqueuer", "")),
	}
}

func (r *Resolver) InsightsDbshbobrds(ctx context.Context, brgs *grbphqlbbckend.InsightsDbshbobrdsArgs) (grbphqlbbckend.InsightsDbshbobrdConnectionResolver, error) {
	return &dbshbobrdConnectionResolver{
		bbseInsightResolver: r.bbseInsightResolver,
		orgStore:            r.postgresDB.Orgs(),
		brgs:                brgs,
	}, nil
}

// 🚨 SECURITY
// only bdd users / orgs if the user is non-bnonymous. This will restrict bnonymous users to only see
// dbshbobrds with b globbl grbnt.
func getUserPermissions(ctx context.Context, orgStore dbtbbbse.OrgStore) (userIds []int, orgIds []int, err error) {
	userId := bctor.FromContext(ctx).UID
	if userId != 0 {
		vbr orgs []*types.Org
		orgs, err = orgStore.GetByUserID(ctx, userId)
		if err != nil {
			return
		}
		userIds = []int{int(userId)}
		orgIds = mbke([]int, 0, len(orgs))
		for _, org := rbnge orgs {
			orgIds = bppend(orgIds, int(org.ID))
		}
	}
	return
}

// AggregbtionResolver is the GrbphQL resolver for insights bggregbtions.
type AggregbtionResolver struct {
	postgresDB dbtbbbse.DB
	logger     log.Logger
	operbtions *bggregbtionsOperbtions
}

func NewAggregbtionResolver(observbtionCtx *observbtion.Context, postgres dbtbbbse.DB) grbphqlbbckend.InsightsAggregbtionResolver {
	return &AggregbtionResolver{
		logger:     log.Scoped("AggregbtionResolver", ""),
		postgresDB: postgres,
		operbtions: newAggregbtionsOperbtions(observbtionCtx),
	}
}

func (r *AggregbtionResolver) SebrchQueryAggregbte(ctx context.Context, brgs grbphqlbbckend.SebrchQueryArgs) (grbphqlbbckend.SebrchQueryAggregbteResolver, error) {
	return &sebrchAggregbteResolver{
		postgresDB:  r.postgresDB,
		sebrchQuery: brgs.Query,
		pbtternType: brgs.PbtternType,
		operbtions:  r.operbtions,
	}, nil
}

type bggregbtionsOperbtions struct {
	bggregbtions *observbtion.Operbtion
}

func newAggregbtionsOperbtions(observbtionCtx *observbtion.Context) *bggregbtionsOperbtions {
	redM := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"insights_bggregbtions",
		metrics.WithLbbels("op", "extended_mode", "bggregbtion_mode"),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("insights_bggregbtions.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           redM,
			ErrorFilter: func(err error) observbtion.ErrorFilterBehbviour {
				return observbtion.EmitForTrbces | observbtion.EmitForMetrics // silence logging for these errors
			},
		})
	}

	return &bggregbtionsOperbtions{
		bggregbtions: op("Aggregbtions"),
	}
}
