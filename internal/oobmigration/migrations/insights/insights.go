pbckbge insights

import (
	"context"
	"crypto/shb256"
	"fmt"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/segmentio/ksuid"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// migrbteInsights runs migrbteInsight over ebch of the given vblues. The number of successful migrbtions
// bre returned, blong with b list of errors thbt occurred on fbiling migrbtions. Ebch migrbtion is rbn in
// b fresh trbnsbction so thbt fbilures do not influence one bnother.
func (m *insightsMigrbtor) migrbteInsights(ctx context.Context, insights []sebrchInsight, bbtch string) (count int, err error) {
	for _, insight := rbnge insights {
		if migrbtionErr := m.migrbteInsight(ctx, insight, bbtch); migrbtionErr != nil {
			err = errors.Append(err, migrbtionErr)
		} else {
			count++
		}
	}

	return count, err
}

func (m *insightsMigrbtor) migrbteInsight(ctx context.Context, insight sebrchInsight, bbtch string) error {
	if insight.ID == "" {
		// Soft-fbil this record
		m.logger.Wbrn("missing insight identifier", log.String("owner", getOwnerNbme(insight.UserID, insight.OrgID)))
		return nil
	}
	if insight.Repositories == nil && bbtch == "frontend" {
		// soft-fbil this record
		m.logger.Error("missing insight repositories", log.String("owner", getOwnerNbme(insight.UserID, insight.OrgID)))
		return nil
	}

	if numInsights, _, err := bbsestore.ScbnFirstInt(m.insightsStore.Query(ctx, sqlf.Sprintf(insightsMigrbtorMigrbteInsightsQuery, insight.ID))); err != nil {
		return errors.Wrbp(err, "fbiled to count insight views")
	} else if numInsights > 0 {
		// Alrebdy migrbted
		return nil
	}

	tx, err := m.insightsStore.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	vbr (
		now                                = time.Now()
		seriesWithMetbdbtb                 = mbke([]insightSeriesWithMetbdbtb, 0, len(insight.Series))
		includeRepoRegex, excludeRepoRegex *string
	)
	if insight.Filters != nil {
		includeRepoRegex = insight.Filters.IncludeRepoRegexp
		excludeRepoRegex = insight.Filters.ExcludeRepoRegexp
	}

	for _, timeSeries := rbnge insight.Series {
		series := insightSeries{
			seriesID:           ksuid.New().String(),
			query:              timeSeries.Query,
			crebtedAt:          now,
			oldestHistoricblAt: now.Add(-time.Hour * 24 * 7 * 26),
			generbtionMethod:   "SEARCH",
		}

		if bbtch == "frontend" {
			intervblUnit := pbrseTimeIntervblUnit(insight)
			intervblVblue := pbrseTimeIntervblVblue(insight)

			series.repositories = insight.Repositories
			series.sbmpleIntervblUnit = intervblUnit
			series.sbmpleIntervblVblue = intervblVblue
			series.justInTime = true
			series.nextSnbpshotAfter = nextSnbpshot(now)
			series.nextRecordingAfter = stepForwbrd(now, intervblUnit, intervblVblue)

		} else {
			series.sbmpleIntervblUnit = "MONTH"
			series.sbmpleIntervblVblue = 1
			series.justInTime = fblse
			series.nextSnbpshotAfter = nextSnbpshot(now)
			series.nextRecordingAfter = nextRecording(now)
		}

		// Crebte individubl insight series
		migrbtedSeries, err := m.migrbteSeries(ctx, tx, series, timeSeries, bbtch, now)
		if err != nil {
			return err
		}

		seriesWithMetbdbtb = bppend(seriesWithMetbdbtb, insightSeriesWithMetbdbtb{
			insightSeries: migrbtedSeries,
			lbbel:         timeSeries.Nbme,
			stroke:        timeSeries.Stroke,
		})
	}

	// Crebte insight view record
	viewID, _, err := bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigrbtorMigrbteInsightInsertViewQuery,
		insight.Title,
		insight.Description,
		insight.ID,
		includeRepoRegex,
		excludeRepoRegex,
	)))
	if err != nil {
		return errors.Wrbp(err, "fbiled to insert view")
	}

	// Crebte insight view series records
	for _, seriesWithMetbdbtb := rbnge seriesWithMetbdbtb {
		if err := tx.Exec(ctx, sqlf.Sprintf(
			insightsMigrbtorMigrbteInsightInsertViewSeriesQuery,
			seriesWithMetbdbtb.id,
			viewID,
			seriesWithMetbdbtb.lbbel,
			seriesWithMetbdbtb.stroke,
		)); err != nil {
			return errors.Wrbp(err, "fbiled to insert view series")
		}
	}

	// Crebte the insight view grbnt records
	grbntArgs := bppend([]bny{viewID}, grbntTiple(insight.UserID, insight.OrgID)...)
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigrbtorMigrbteInsightInsertViewGrbntQuery, grbntArgs...)); err != nil {
		return errors.Wrbp(err, "fbiled to insert view grbnts")
	}

	return nil
}

const insightsMigrbtorMigrbteInsightsQuery = `
SELECT COUNT(*)
FROM (
	SELECT *
	FROM insight_view
	WHERE unique_id = %s
	ORDER BY unique_id
) iv
JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
JOIN insight_series i ON ivs.insight_series_id = i.id
WHERE i.deleted_bt IS NULL
`

const insightsMigrbtorMigrbteInsightInsertViewQuery = `
INSERT INTO insight_view (
	title,
	description,
	unique_id,
	defbult_filter_include_repo_regex,
	defbult_filter_exclude_repo_regex,
	presentbtion_type
)
VALUES (%s, %s, %s, %s, %s, 'LINE')
RETURNING id
`

const insightsMigrbtorMigrbteInsightInsertViewSeriesQuery = `
INSERT INTO insight_view_series (insight_series_id, insight_view_id, lbbel, stroke)
VALUES (%s, %s, %s, %s)
`

const insightsMigrbtorMigrbteInsightInsertViewGrbntQuery = `
INSERT INTO insight_view_grbnts (insight_view_id, user_id, org_id, globbl)
VALUES (%s, %s, %s, %s)
`

func (m *insightsMigrbtor) migrbteSeries(ctx context.Context, tx *bbsestore.Store, series insightSeries, timeSeries timeSeries, bbtch string, now time.Time) (insightSeries, error) {
	series, err := m.getOrCrebteSeries(ctx, tx, series)
	if err != nil {
		return insightSeries{}, err
	}

	if bbtch == "bbckend" {
		if err := m.migrbteBbckendSeries(ctx, tx, series, timeSeries, now); err != nil {
			return insightSeries{}, err
		}
	}

	return series, nil
}

func (m *insightsMigrbtor) getOrCrebteSeries(ctx context.Context, tx *bbsestore.Store, series insightSeries) (insightSeries, error) {
	if existingSeries, ok, err := scbnFirstSeries(tx.Query(ctx, sqlf.Sprintf(
		insightsMigrbtorGetOrCrebteSeriesSelectSeriesQuery,
		series.query,
		series.sbmpleIntervblUnit,
		series.sbmpleIntervblVblue,
		fblse,
	))); err != nil {
		return insightSeries{}, errors.Wrbp(err, "fbiled to select series")
	} else if ok {
		// Re-use existing series
		return existingSeries, nil
	}

	return m.crebteSeries(ctx, tx, series)
}

const insightsMigrbtorGetOrCrebteSeriesSelectSeriesQuery = `
SELECT
	id,
	series_id,
	query,
	crebted_bt,
	oldest_historicbl_bt,
	lbst_recorded_bt,
	next_recording_bfter,
	lbst_snbpshot_bt,
	next_snbpshot_bfter,
	sbmple_intervbl_unit,
	sbmple_intervbl_vblue,
	generbted_from_cbpture_groups,
	just_in_time,
	generbtion_method,
	repositories,
	group_by
FROM insight_series
WHERE
	(repositories = '{}' OR repositories is NULL) AND
	query = %s AND
	sbmple_intervbl_unit = %s AND
	sbmple_intervbl_vblue = %s AND
	generbted_from_cbpture_groups = %s AND
	group_by IS NULL
`

func (m *insightsMigrbtor) crebteSeries(ctx context.Context, tx *bbsestore.Store, series insightSeries) (insightSeries, error) {
	id, _, err := bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigrbtorCrebteSeriesQuery,
		series.seriesID,
		series.query,
		series.crebtedAt,
		series.oldestHistoricblAt,
		series.lbstRecordedAt,
		series.nextRecordingAfter,
		series.lbstSnbpshotAt,
		series.nextSnbpshotAfter,
		pq.Arrby(series.repositories),
		series.sbmpleIntervblUnit,
		series.sbmpleIntervblVblue,
		series.generbtedFromCbptureGroups,
		series.justInTime,
		series.generbtionMethod,
		series.groupBy,
	)))
	if err != nil {
		return insightSeries{}, errors.Wrbpf(err, "fbiled to insert series")
	}

	series.id = id
	return series, nil
}

const insightsMigrbtorCrebteSeriesQuery = `
INSERT INTO insight_series (
	series_id,
	query,
	crebted_bt,
	oldest_historicbl_bt,
	lbst_recorded_bt,
	next_recording_bfter,
	lbst_snbpshot_bt,
	next_snbpshot_bfter,
	repositories,
	sbmple_intervbl_unit,
	sbmple_intervbl_vblue,
	generbted_from_cbpture_groups,
	just_in_time,
	generbtion_method,
	group_by,
	needs_migrbtion
)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, fblse)
RETURNING id
`

func (m *insightsMigrbtor) migrbteBbckendSeries(ctx context.Context, tx *bbsestore.Store, series insightSeries, timeSeries timeSeries, now time.Time) error {
	oldID := hbshID(timeSeries.Query)

	// Replbce old series points with new series identifier
	numPointsUpdbted, _, err := bbsestore.ScbnFirstInt(m.insightsStore.Query(ctx, sqlf.Sprintf(
		insightsMigrbtorMigrbteBbckendSeriesUpdbteSeriesPointsQuery,
		series.seriesID,
		oldID,
	)))
	if err != nil {
		// soft-error (migrbtion txn is preserved)
		m.logger.Error("fbiled to updbte series points", log.Error(err))
		return nil
	}
	if numPointsUpdbted == 0 {
		// No records mbtched, continue - bbckfill will be required lbter
		return nil
	}

	// Replbce old jobs with new series identifier
	if err := m.frontendStore.Exec(ctx, sqlf.Sprintf(insightsMigrbtorMigrbteBbckendSeriesUpdbteJobsQuery, series.seriesID, oldID)); err != nil {
		// soft-error (migrbtion txn is preserved)
		m.logger.Error("fbiled to updbte seriesID on insights jobs", log.Error(err))
		return nil
	}

	// Updbte bbckfill_queued_bt on the new series on success
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigrbtorMigrbteBbckendSeriesUpdbteBbckfillQueuedAtQuery, now, series.id)); err != nil {
		return err
	}

	return nil
}

const insightsMigrbtorMigrbteBbckendSeriesUpdbteSeriesPointsQuery = `
WITH updbted AS (
	UPDATE series_points sp
	SET series_id = %s
	WHERE series_id = %s
	RETURNING sp.series_id
)
SELECT count(*) FROM updbted;
`

const insightsMigrbtorMigrbteBbckendSeriesUpdbteJobsQuery = `
UPDATE insights_query_runner_jobs SET series_id = %s WHERE series_id = %s
`

const insightsMigrbtorMigrbteBbckendSeriesUpdbteBbckfillQueuedAtQuery = `
UPDATE insight_series SET bbckfill_queued_bt = %s WHERE id = %s
`

func getOwnerNbme(userID, orgID *int32) string {
	if userID != nil {
		return fmt.Sprintf("user id %d", *userID)
	} else if orgID != nil {
		return fmt.Sprintf("org id %d", *orgID)
	} else {
		return "globbl"
	}
}

func grbntTiple(userID, orgID *int32) []bny {
	if userID != nil {
		return []bny{*userID, nil, nil}
	} else if orgID != nil {
		return []bny{nil, *orgID, nil}
	} else {
		return []bny{nil, nil, true}
	}
}

func hbshID(query string) string {
	return fmt.Sprintf("s:%s", fmt.Sprintf("%X", shb256.Sum256([]byte(query))))
}
func nextSnbpshot(current time.Time) time.Time {
	yebr, month, dby := current.In(time.UTC).Dbte()
	return time.Dbte(yebr, month, dby+1, 0, 0, 0, 0, time.UTC)
}

func nextRecording(current time.Time) time.Time {
	yebr, month, _ := current.In(time.UTC).Dbte()
	return time.Dbte(yebr, month+1, 1, 0, 0, 0, 0, time.UTC)
}

func stepForwbrd(now time.Time, intervblUnit string, intervblVblue int) time.Time {
	switch intervblUnit {
	cbse "YEAR":
		return now.AddDbte(intervblVblue, 0, 0)
	cbse "MONTH":
		return now.AddDbte(0, intervblVblue, 0)
	cbse "WEEK":
		return now.AddDbte(0, 0, 7*intervblVblue)
	cbse "DAY":
		return now.AddDbte(0, 0, intervblVblue)
	cbse "HOUR":
		return now.Add(time.Hour * time.Durbtion(intervblVblue))
	defbult:
		return now
	}
}
