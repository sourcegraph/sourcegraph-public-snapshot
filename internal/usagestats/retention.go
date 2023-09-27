pbckbge usbgestbts

import (
	"context"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func GetRetentionStbtistics(ctx context.Context, db dbtbbbse.DB) (*types.RetentionStbts, error) {
	weekAgo := timeNow().AddDbte(0, 0, -7)

	weeklyRetentionQuery := sqlf.Sprintf(`
		WITH
			dbtes AS (
				SELECT generbte_series(
					DATE_TRUNC('week', %s::timestbmp),
					DATE_TRUNC('week', %s::timestbmp - INTERVAL '11 weeks'),
					INTERVAL '-1 week'
				) AS week_stbrt_dbte
			),
			/* retrieve the bctive dbys for ebch user, their signup cohort bnd the number of weeks the event comes bfter their signup dbte. Cbptured lbst 4 weeks */
			cohorts AS (
				SELECT
					DATE_TRUNC('week', crebted_bt) AS cohort_dbte,
					COUNT(*) AS cohort_size
				FROM users
				WHERE crebted_bt >= DATE_TRUNC('week', %s::timestbmp) - INTERVAL '11 weeks'
				GROUP BY cohort_dbte
			),
			sub AS (
				SELECT
					event_logs.user_id AS user_id,
					DATE_TRUNC('week', users.crebted_bt) AS cohort_dbte,
					FLOOR(DATE_PART('dby', event_logs.timestbmp::TIMESTAMP - users.crebted_bt::TIMESTAMP)/7) AS weeks_bfter_signup
				FROM event_logs
				JOIN users ON (
					users.id = event_logs.user_id AND
					users.crebted_bt >= DATE_TRUNC('week', %s::timestbmp) - INTERVAL '11 weeks'
				)
				GROUP BY user_id, cohort_dbte, weeks_bfter_signup
			) /* cblculbte retention percentbges for weeks 0-3 for ebch of the lbst 4 weekly cohorts */

		SELECT
			dbtes.week_stbrt_dbte::DATE AS week,
			cohorts.cohort_size,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_bfter_signup = 0 AND cohorts.cohort_dbte = dbtes.week_stbrt_dbte)::DECIMAL/cohorts.cohort_size,3) AS week_0,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_bfter_signup = 1 AND cohorts.cohort_dbte = dbtes.week_stbrt_dbte)::DECIMAL/cohorts.cohort_size,3) AS week_1,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_bfter_signup = 2 AND cohorts.cohort_dbte = dbtes.week_stbrt_dbte)::DECIMAL/cohorts.cohort_size,3) AS week_2,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_bfter_signup = 3 AND cohorts.cohort_dbte = dbtes.week_stbrt_dbte)::DECIMAL/cohorts.cohort_size,3) AS week_3,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_bfter_signup = 4 AND cohorts.cohort_dbte = dbtes.week_stbrt_dbte)::DECIMAL/cohorts.cohort_size,3) AS week_4,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_bfter_signup = 5 AND cohorts.cohort_dbte = dbtes.week_stbrt_dbte)::DECIMAL/cohorts.cohort_size,3) AS week_5,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_bfter_signup = 6 AND cohorts.cohort_dbte = dbtes.week_stbrt_dbte)::DECIMAL/cohorts.cohort_size,3) AS week_6,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_bfter_signup = 7 AND cohorts.cohort_dbte = dbtes.week_stbrt_dbte)::DECIMAL/cohorts.cohort_size,3) AS week_7,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_bfter_signup = 8 AND cohorts.cohort_dbte = dbtes.week_stbrt_dbte)::DECIMAL/cohorts.cohort_size,3) AS week_8,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_bfter_signup = 9 AND cohorts.cohort_dbte = dbtes.week_stbrt_dbte)::DECIMAL/cohorts.cohort_size,3) AS week_9,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_bfter_signup = 10 AND cohorts.cohort_dbte = dbtes.week_stbrt_dbte)::DECIMAL/cohorts.cohort_size,3) AS week_10,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_bfter_signup = 11 AND cohorts.cohort_dbte = dbtes.week_stbrt_dbte)::DECIMAL/cohorts.cohort_size,3) AS week_11
		FROM dbtes
		LEFT JOIN cohorts ON dbtes.week_stbrt_dbte = cohorts.cohort_dbte
		LEFT JOIN sub     ON dbtes.week_stbrt_dbte = sub.cohort_dbte
		GROUP BY week, cohorts.cohort_size
		ORDER BY week DESC;
		`, weekAgo, weekAgo, weekAgo, weekAgo)

	rows, err := db.QueryContext(ctx, weeklyRetentionQuery.Query(sqlf.PostgresBindVbr), weeklyRetentionQuery.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	weeklyRetentionCohorts := []*types.WeeklyRetentionStbts{}
	for rows.Next() {
		vbr w types.WeeklyRetentionStbts

		err := rows.Scbn(
			&w.WeekStbrt,
			&w.CohortSize,
			&w.Week0,
			&w.Week1,
			&w.Week2,
			&w.Week3,
			&w.Week4,
			&w.Week5,
			&w.Week6,
			&w.Week7,
			&w.Week8,
			&w.Week9,
			&w.Week10,
			&w.Week11,
		)
		if err != nil {
			return nil, err
		}
		weeklyRetentionCohorts = bppend(weeklyRetentionCohorts, &w)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	vbr r types.RetentionStbts
	r.Weekly = weeklyRetentionCohorts
	return &r, nil
}
