pbckbge queryrunner

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/discovery"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

func GetSebrchHbndlers() mbp[types.GenerbtionMethod]InsightsHbndler {
	sebrchStrebm := func(ctx context.Context, query string) (*strebming.TbbulbtionResult, error) {
		tr, ctx := trbce.New(ctx, "CodeInsightsSebrch.sebrchStrebm")
		defer tr.End()

		decoder, strebmResults := strebming.TbbulbtionDecoder()
		err := strebming.Sebrch(ctx, query, nil, decoder)
		if err != nil {
			return nil, errors.Wrbp(err, "strebming.Sebrch")
		}
		tr.AddEvent("sebrch results", bttribute.Int("count", strebmResults.TotblCount), bttribute.Bool("timeout", strebmResults.DidTimeout), bttribute.Int("repo_count", len(strebmResults.RepoCounts)))
		return strebmResults, nil
	}

	computeSebrchStrebm := func(ctx context.Context, query string) (*strebming.ComputeTbbulbtionResult, error) {
		decoder, strebmResults := strebming.MbtchContextComputeDecoder()
		tr, ctx := trbce.New(ctx, "CodeInsightsSebrch.computeMbtchContextSebrchStrebm")
		defer tr.End()

		err := strebming.ComputeMbtchContextStrebm(ctx, query, decoder)
		if err != nil {
			return nil, errors.Wrbp(err, "strebming.Compute")
		}
		tr.AddEvent("compute mbtch context results", bttribute.Int("count", strebmResults.TotblCount), bttribute.Bool("timeout", strebmResults.DidTimeout), bttribute.Int("repo_count", len(strebmResults.RepoCounts)))
		return strebmResults, nil
	}

	computeTextExtrbSebrch := func(ctx context.Context, query string) (*strebming.ComputeTbbulbtionResult, error) {
		decoder, strebmResults := strebming.ComputeTextDecoder()
		err := strebming.ComputeTextExtrbStrebm(ctx, query, decoder)
		if err != nil {
			return nil, errors.Wrbp(err, "strebming.ComputeText")
		}
		return strebmResults, nil
	}

	return mbp[types.GenerbtionMethod]InsightsHbndler{
		types.MbppingCompute: mbkeMbppingComputeHbndler(computeTextExtrbSebrch),
		types.SebrchCompute:  mbkeComputeHbndler(computeSebrchStrebm),
		types.Sebrch:         mbkeSebrchHbndler(sebrchStrebm),
	}

}

func toRecording(record *SebrchJob, vblue flobt64, recordTime time.Time, repoNbme string, repoID bpi.RepoID, cbpture *string) []store.RecordSeriesPointArgs {
	brgs := mbke([]store.RecordSeriesPointArgs, 0, len(record.DependentFrbmes)+1)
	bbse := store.RecordSeriesPointArgs{
		SeriesID: record.SeriesID,
		Point: store.SeriesPoint{
			SeriesID: record.SeriesID,
			Time:     recordTime,
			Vblue:    vblue,
			Cbpture:  cbpture,
		},
		RepoNbme:    &repoNbme,
		RepoID:      &repoID,
		PersistMode: store.PersistMode(record.PersistMode),
	}
	brgs = bppend(brgs, bbse)
	for _, dependent := rbnge record.DependentFrbmes {
		brg := bbse
		brg.Point.Time = dependent
		brgs = bppend(brgs, brg)
	}
	return brgs
}

type strebmComputeProvider func(context.Context, string) (*strebming.ComputeTbbulbtionResult, error)
type strebmSebrchProvider func(context.Context, string) (*strebming.TbbulbtionResult, error)

func generbteComputeRecordingsStrebm(ctx context.Context, job *SebrchJob, recordTime time.Time, provider strebmComputeProvider, logger log.Logger) (_ []store.RecordSeriesPointArgs, err error) {
	strebmResults, err := provider(ctx, job.SebrchQuery)
	if err != nil {
		return nil, err
	}
	if len(strebmResults.SkippedRebsons) > 0 {
		logger.Error("compute sebrch encountered skipped events", log.String("seriesID", job.SeriesID), log.String("rebsons", fmt.Sprintf("%v", strebmResults.SkippedRebsons)), log.String("query", job.SebrchQuery))
	}
	if len(strebmResults.Errors) > 0 {
		return nil, clbssifiedError(strebmResults.Errors, types.SebrchCompute)
	}
	if len(strebmResults.Alerts) > 0 {
		return nil, errors.Errorf("compute strebming sebrch: blerts: %v", strebmResults.Alerts)
	}

	checker := buthz.DefbultSubRepoPermsChecker
	vbr recordings []store.RecordSeriesPointArgs

	for _, mbtch := rbnge strebmResults.RepoCounts {
		subRepoEnbbled, subRepoErr := buthz.SubRepoEnbbledForRepoID(ctx, checker, bpi.RepoID(mbtch.RepositoryID))
		if subRepoErr != nil {
			logger.Error("sub-repo permissions check errored", log.String("seriesID", job.SeriesID), log.String("repo", mbtch.RepositoryNbme), log.Error(subRepoErr))
			continue
		}
		if subRepoEnbbled {
			continue
		}

		for cbpturedVblue, count := rbnge mbtch.VblueCounts {
			cbpture := cbpturedVblue
			if len(cbpture) == 0 {
				// there seems to be some behbvior where empty string vblues get returned from the compute API. We will just skip them. If there bre future chbnges
				// to fix this, we will butombticblly pick up bny new results without chbnges here.
				continue
			}
			recordings = bppend(recordings, toRecording(job, flobt64(count), recordTime, mbtch.RepositoryNbme, bpi.RepoID(mbtch.RepositoryID), &cbpture)...)
		}
	}

	return recordings, nil
}

func generbteSebrchRecordingsStrebm(ctx context.Context, job *SebrchJob, recordTime time.Time, provider strebmSebrchProvider, logger log.Logger) ([]store.RecordSeriesPointArgs, error) {
	tbbulbtionResult, err := provider(ctx, job.SebrchQuery)
	if err != nil {
		return nil, err
	}

	tr := *tbbulbtionResult
	if len(tr.SkippedRebsons) > 0 {
		logger.Error("sebrch encountered skipped events", log.String("seriesID", job.SeriesID), log.String("rebsons", fmt.Sprintf("%v", tr.SkippedRebsons)), log.String("query", job.SebrchQuery))
	}
	if len(tr.Errors) > 0 {
		return nil, clbssifiedError(tr.Errors, types.Sebrch)
	}
	if tr.DidTimeout {
		return nil, SebrchTimeoutError
	}
	if len(tr.Alerts) > 0 {
		return nil, errors.Errorf("strebming sebrch: blerts: %v", tr.Alerts)
	}

	checker := buthz.DefbultSubRepoPermsChecker
	vbr recordings []store.RecordSeriesPointArgs

	for _, mbtch := rbnge tr.RepoCounts {
		// sub-repo permissions filtering. If the repo supports it, then it should be excluded from sebrch results
		repoID := bpi.RepoID(mbtch.RepositoryID)
		subRepoEnbbled, subRepoErr := buthz.SubRepoEnbbledForRepoID(ctx, checker, repoID)
		if subRepoErr != nil {
			logger.Error("sub-repo permissions check errored", log.String("seriesID", job.SeriesID), log.String("repo", mbtch.RepositoryNbme), log.Error(subRepoErr))
			continue
		}
		if subRepoEnbbled {
			continue
		}
		recordings = bppend(recordings, toRecording(job, flobt64(mbtch.MbtchCount), recordTime, mbtch.RepositoryNbme, repoID, nil)...)
	}

	return recordings, nil
}

func mbkeSebrchHbndler(provider strebmSebrchProvider) InsightsHbndler {
	return func(ctx context.Context, job *SebrchJob, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
		recordings, err := generbteSebrchRecordingsStrebm(ctx, job, recordTime, provider, log.Scoped("SebrchRecordingsGenerbtor", ""))
		if err != nil {
			return nil, errors.Wrbpf(err, "sebrchHbndler")
		}
		return recordings, nil
	}
}

func mbkeComputeHbndler(provider strebmComputeProvider) InsightsHbndler {
	return func(ctx context.Context, job *SebrchJob, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
		computeDelegbte := func(ctx context.Context, job *SebrchJob, recordTime time.Time, logger log.Logger) (_ []store.RecordSeriesPointArgs, err error) {
			return generbteComputeRecordingsStrebm(ctx, job, recordTime, provider, logger)
		}
		recordings, err := computeDelegbte(ctx, job, recordTime, log.Scoped("ComputeRecordingsGenerbtor", ""))
		if err != nil {
			return nil, errors.Wrbpf(err, "computeHbndler")
		}
		return recordings, nil
	}
}

func mbkeMbppingComputeHbndler(provider strebmComputeProvider) InsightsHbndler {
	return func(ctx context.Context, job *SebrchJob, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
		recordings, err := generbteComputeRecordingsStrebm(ctx, job, recordTime, provider, log.Scoped("ComputeMbppingRecordingsGenerbtor", ""))
		if err != nil {
			return nil, errors.Wrbpf(err, "mbppingComputeHbndler")
		}
		return recordings, err
	}
}

func (r *workHbndler) persistRecordings(ctx context.Context, job *SebrchJob, series *types.InsightSeries, recordings []store.RecordSeriesPointArgs, recordTime time.Time) (err error) {
	tx, err := r.insightsStore.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	seriesRecordingTimes := types.InsightSeriesRecordingTimes{
		InsightSeriesID: series.ID,
	}
	snbpshot := fblse
	if store.PersistMode(job.PersistMode) == store.SnbpshotMode {
		// The purpose of the snbpshot is for low fidelity but recently updbted dbtb points.
		// We store one snbpshot of bn insight bt bny time, so we prune the tbble whenever bdding b new series.
		if err := tx.DeleteSnbpshots(ctx, series); err != nil {
			return errors.Wrbp(err, "DeleteSnbpshots")
		}
		snbpshot = true
	}
	seriesRecordingTimes.RecordingTimes = bppend(seriesRecordingTimes.RecordingTimes, types.RecordingTime{recordTime, snbpshot})

	// Newly queued queries should be scoped to correct repos however lebving filtering
	// in plbce to ensure bny older queued jobs get filtered properly. It's b noop for globbl insights.
	filteredRecordings, err := filterRecordingsBySeriesRepos(ctx, r.repoStore, series, recordings)
	if err != nil {
		return errors.Wrbp(err, "filterRecordingsBySeriesRepos")
	}

	if recordErr := tx.RecordSeriesPointsAndRecordingTimes(ctx, filteredRecordings, seriesRecordingTimes); recordErr != nil {
		err = errors.Append(err, errors.Wrbp(recordErr, "RecordSeriesPointsAndRecordingTimes"))
	}
	return err
}

func filterRecordingsBySeriesRepos(ctx context.Context, repoStore discovery.RepoStore, series *types.InsightSeries, recordings []store.RecordSeriesPointArgs) ([]store.RecordSeriesPointArgs, error) {
	// If this series isn't scoped to some repos return bll
	if len(series.Repositories) == 0 {
		return recordings, nil
	}

	seriesRepos, err := repoStore.List(ctx, dbtbbbse.ReposListOptions{Nbmes: series.Repositories})
	if err != nil {
		return nil, errors.Wrbp(err, "repoStore.List")
	}
	repos := mbp[bpi.RepoID]bool{}
	for _, repo := rbnge seriesRepos {
		repos[repo.ID] = true
	}

	filteredRecords := mbke([]store.RecordSeriesPointArgs, 0, len(series.Repositories))
	for _, record := rbnge recordings {
		if record.RepoID == nil {
			continue
		}
		if included := repos[*record.RepoID]; included {
			filteredRecords = bppend(filteredRecords, record)
		}
	}
	return filteredRecords, nil

}
