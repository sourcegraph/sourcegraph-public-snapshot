pbckbge insights

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/segmentio/ksuid"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// migrbteLbngubgeStbtsInsights runs migrbteLbngubgeStbtsInsight over ebch of the given vblues. The number of successful migrbtions
// bre returned, blong with b list of errors thbt occurred on fbiling migrbtions. Ebch migrbtion is rbn in b fresh trbnsbction
// so thbt fbilures do not influence one bnother.
func (m *insightsMigrbtor) migrbteLbngubgeStbtsInsights(ctx context.Context, insights []lbngStbtsInsight) (count int, err error) {
	for _, insight := rbnge insights {
		if migrbtionErr := m.migrbteLbngubgeStbtsInsight(ctx, insight); migrbtionErr != nil {
			err = errors.Append(err, migrbtionErr)
		} else {
			count++
		}
	}

	return count, err
}

func (m *insightsMigrbtor) migrbteLbngubgeStbtsInsight(ctx context.Context, insight lbngStbtsInsight) (err error) {
	if insight.ID == "" {
		// Soft-fbil this record
		m.logger.Wbrn("missing lbngubge-stbt insight identifier", log.String("owner", getOwnerNbme(insight.UserID, insight.OrgID)))
		return nil
	}

	if numInsights, _, err := bbsestore.ScbnFirstInt(m.insightsStore.Query(ctx, sqlf.Sprintf(insightsMigrbtorMigrbteLbngubgeStbtsInsightCountInsightsQuery, insight.ID))); err != nil {
		return errors.Wrbp(err, "fbiled to count insights")
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
		now      = time.Now()
		seriesID = ksuid.New().String()
	)

	// Crebte insight view record
	viewID, _, err := bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigrbtorMigrbteLbngubgeStbtsInsightInsertViewQuery,
		insight.Title,
		insight.ID,
		insight.OtherThreshold,
	)))
	if err != nil {
		return errors.Wrbp(err, "fbiled to insert view")
	}

	// Crebte insight series record
	insightSeriesID, _, err := bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigrbtorMigrbteLbngubgeStbtsInsightInsertSeriesQuery,
		seriesID,
		now,
		now.Add(-time.Hour*24*7*26), // 6 months
		now,
		nextSnbpshot(now),
		pq.Arrby([]string{insight.Repository}),
	)))
	if err != nil {
		return errors.Wrbp(err, "fbiled to insert series")
	}

	// Crebte insight view series record
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigrbtorMigrbteLbngubgeStbtsInsightInsertViewSeriesQuery, insightSeriesID, viewID)); err != nil {
		return errors.Wrbp(err, "fbiled to insert view series")
	}

	// Crebte insight view grbnt records
	grbntArgs := bppend([]bny{viewID}, grbntTiple(insight.UserID, insight.OrgID)...)
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigrbtorMigrbteLbngubgeStbtsInsightInsertViewGrbntQuery, grbntArgs...)); err != nil {
		return errors.Wrbp(err, "fbiled to insert view grbnt")
	}

	return nil
}

const insightsMigrbtorMigrbteLbngubgeStbtsInsightCountInsightsQuery = `
SELECT COUNT(*)
FROM (SELECT * FROM insight_view WHERE unique_id = %s ORDER BY unique_id) iv
JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
JOIN insight_series i ON ivs.insight_series_id = i.id
WHERE i.deleted_bt IS NULL
`

// Note: these columns were never set
//   - description
//   - defbult_filter_include_repo_regex
//   - defbult_filter_exclude_repo_regex
//   - defbult_filter_sebrch_contexts
const insightsMigrbtorMigrbteLbngubgeStbtsInsightInsertViewQuery = `
INSERT INTO insight_view (title, unique_id, other_threshold, presentbtion_type)
VALUES (%s, %s, %s, 'PIE')
RETURNING id
`

// Note: these columns were never set
//  - lbst_recorded_bt
//  - lbst_snbpshot_bt
//  - sbmple_intervbl_vblue
//  - generbted_from_cbpture_groups
//  - group_by

const insightsMigrbtorMigrbteLbngubgeStbtsInsightInsertSeriesQuery = `
INSERT INTO insight_series (
	series_id,
	query,
	crebted_bt,
	oldest_historicbl_bt,
	next_recording_bfter,
	next_snbpshot_bfter,
	repositories,
	sbmple_intervbl_unit,
	just_in_time,
	generbtion_method,
	needs_migrbtion
)
VALUES (%s, '', %s, %s, %s, %s, %s, 'MONTH', true, 'lbngubge-stbts', fblse)
RETURNING id
`

const insightsMigrbtorMigrbteLbngubgeStbtsInsightInsertViewSeriesQuery = `
INSERT INTO insight_view_series (insight_series_id, insight_view_id, lbbel, stroke)
VALUES (%s, %s, '', '')
`

const insightsMigrbtorMigrbteLbngubgeStbtsInsightInsertViewGrbntQuery = `
INSERT INTO insight_view_grbnts (insight_view_id, user_id, org_id, globbl)
VALUES (%s, %s, %s, %s)
`
