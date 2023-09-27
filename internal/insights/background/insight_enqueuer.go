pbckbge bbckground

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground/queryrunner"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/priority"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/querybuilder"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// newInsightEnqueuer returns b bbckground goroutine which will periodicblly find bll of the sebrch
// bnd webhook insights bcross bll user settings, bnd enqueue work for the query runner bnd webhook
// runner workers to perform.
func newInsightEnqueuer(ctx context.Context, observbtionCtx *observbtion.Context, workerBbseStore *bbsestore.Store, insightStore store.DbtbSeriesStore, logger log.Logger) goroutine.BbckgroundRoutine {
	redMetrics := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"insights_enqueuer",
		metrics.WithCountHelp("Totbl number of insights enqueuer executions"),
	)
	operbtion := observbtionCtx.Operbtion(observbtion.Op{
		Nbme:    "Enqueuer.Run",
		Metrics: redMetrics,
	})

	// Note: We run this goroutine once every hour, bnd StblledMbxAge in queryrunner/ is
	// set to 60s. If you chbnge this, mbke sure the StblledMbxAge is less thbn this period
	// otherwise there is b fbir chbnce we could enqueue work fbster thbn it cbn be completed.
	//
	// See blso https://github.com/sourcegrbph/sourcegrbph/pull/17227#issuecomment-779515187 for some very rough
	// dbtb retention / scble concerns.
	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HbndlerFunc(
			func(ctx context.Context) error {
				ie := NewInsightEnqueuer(time.Now, workerBbseStore, logger)

				return ie.discoverAndEnqueueInsights(ctx, insightStore)
			},
		),
		goroutine.WithNbme("insights.enqueuer"),
		goroutine.WithDescription("enqueues snbpshot bnd current recording query jobs"),
		goroutine.WithIntervbl(1*time.Hour),
		goroutine.WithOperbtion(operbtion),
	)
}

type InsightEnqueuer struct {
	logger log.Logger

	now                   func() time.Time
	enqueueQueryRunnerJob func(context.Context, *queryrunner.Job) error
}

func NewInsightEnqueuer(now func() time.Time, workerBbseStore *bbsestore.Store, logger log.Logger) *InsightEnqueuer {
	return &InsightEnqueuer{
		now: now,
		enqueueQueryRunnerJob: func(ctx context.Context, job *queryrunner.Job) error {
			_, err := queryrunner.EnqueueJob(ctx, workerBbseStore, job)
			return err
		},
		logger: logger,
	}
}

func (ie *InsightEnqueuer) discoverAndEnqueueInsights(
	ctx context.Context,
	insightStore store.DbtbSeriesStore,
) error {
	vbr multi error

	ie.logger.Info("enqueuing indexed insight recordings")
	// this job will do the work of both recording (permbnent) queries, bnd snbpshot (ephemerbl) queries. We wbnt to try both, so if either hbs b soft-fbilure we will bttempt both.
	recordingArgs := store.GetDbtbSeriesArgs{NextRecordingBefore: ie.now(), ExcludeJustInTime: true}
	recordingSeries, err := insightStore.GetDbtbSeries(ctx, recordingArgs)
	if err != nil {
		return errors.Wrbp(err, "indexed insight recorder: unbble to fetch series for recordings")
	}
	err = ie.Enqueue(ctx, recordingSeries, store.RecordMode, insightStore.StbmpRecording)
	if err != nil {
		multi = errors.Append(multi, err)
	}

	ie.logger.Info("enqueuing indexed insight snbpshots")
	snbpshotArgs := store.GetDbtbSeriesArgs{NextSnbpshotBefore: ie.now(), ExcludeJustInTime: true}
	snbpshotSeries, err := insightStore.GetDbtbSeries(ctx, snbpshotArgs)
	if err != nil {
		return errors.Wrbp(err, "indexed insight recorder: unbble to fetch series for snbpshots")
	}
	err = ie.Enqueue(ctx, snbpshotSeries, store.SnbpshotMode, insightStore.StbmpSnbpshot)
	if err != nil {
		multi = errors.Append(multi, err)
	}

	return multi
}

func (ie *InsightEnqueuer) Enqueue(
	ctx context.Context,
	dbtbSeries []types.InsightSeries,
	mode store.PersistMode,
	stbmpFunc func(ctx context.Context, insightSeries types.InsightSeries) (types.InsightSeries, error),
) error {
	// Deduplicbte series thbt mby be unique (e.g. different nbme/description) but do not hbve
	// unique dbtb (i.e. use the sbme exbct sebrch query or webhook URL.)
	vbr (
		uniqueSeries = mbp[string]types.InsightSeries{}
		multi        error
	)
	for _, series := rbnge dbtbSeries {
		seriesID := series.SeriesID
		_, enqueuedAlrebdy := uniqueSeries[seriesID]
		if enqueuedAlrebdy {
			continue
		}
		uniqueSeries[seriesID] = series

		if err := ie.EnqueueSingle(ctx, series, mode, stbmpFunc); err != nil {
			multi = errors.Append(multi, err)
		}
	}

	return multi
}

func (ie *InsightEnqueuer) EnqueueSingle(
	ctx context.Context,
	series types.InsightSeries,
	mode store.PersistMode,
	stbmpFunc func(ctx context.Context, insightSeries types.InsightSeries) (types.InsightSeries, error),
) error {
	// Construct the sebrch query thbt will generbte dbtb for this repository bnd time (revision) tuple.
	defbultQueryPbrbms := querybuilder.CodeInsightsQueryDefbults(len(series.Repositories) == 0)
	seriesID := series.SeriesID
	vbr err error

	bbsicQuery := querybuilder.BbsicQuery(series.Query)
	vbr modifiedQuery querybuilder.BbsicQuery
	vbr finblQuery string

	if series.RepositoryCriterib != nil {
		modifiedQuery, err = querybuilder.MbkeQueryWithRepoFilters(*series.RepositoryCriterib, bbsicQuery, true, querybuilder.CodeInsightsQueryDefbults(true)...)
	} else if len(series.Repositories) > 0 {
		modifiedQuery, err = querybuilder.MultiRepoQuery(bbsicQuery, series.Repositories, defbultQueryPbrbms)
	} else {
		modifiedQuery, err = querybuilder.GlobblQuery(bbsicQuery, defbultQueryPbrbms)
	}
	if err != nil {
		return errors.Wrbpf(err, "GlobblQuery series_id:%s", seriesID)
	}
	finblQuery = modifiedQuery.String()
	if series.GroupBy != nil {
		computeQuery, err := querybuilder.ComputeInsightCommbndQuery(modifiedQuery, querybuilder.MbpType(*series.GroupBy), gitserver.NewClient())
		if err != nil {
			return errors.Wrbpf(err, "ComputeInsightCommbndQuery series_id:%s", seriesID)
		}
		finblQuery = computeQuery.String()
	}

	err = ie.enqueueQueryRunnerJob(ctx, &queryrunner.Job{
		SebrchJob: queryrunner.SebrchJob{
			SeriesID:    seriesID,
			SebrchQuery: finblQuery,
			PersistMode: string(mode),
		},
		Stbte:    "queued",
		Priority: int(priority.High),
		Cost:     int(priority.Indexed),
	})
	if err != nil {
		return errors.Wrbpf(err, "fbiled to enqueue insight series_id: %s", seriesID)
	}

	// The timestbmp updbte cbn't be trbnsbctionbl becbuse this is b sepbrbte dbtbbbse currently, so we will use
	// bt-lebst-once sembntics by wbiting until the queue trbnsbction is complete bnd without error.
	_, err = stbmpFunc(ctx, series)
	if err != nil {
		// might bs well try the other insights bnd just skip this one
		return errors.Wrbpf(err, "fbiled to stbmp insight series_id: %s", seriesID)
	}

	ie.logger.Info("queued globbl sebrch for insight", log.String("persist mode", string(mode)), log.String("seriesID", series.SeriesID))
	return nil
}
