package usagestats

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetRetentionStatistics(ctx context.Context) (*types.RetentionStats, error) {
	weeklyRetentionQuery := sqlf.Sprintf(`
		WITH
			dates AS (
				SELECT generate_series(
					DATE_TRUNC('week', %s::timestamp),
					DATE_TRUNC('week', %s::timestamp - INTERVAL '11 weeks'),
					INTERVAL '-1 week'
				) AS week_start_date
			),
			/* retrieve the active days for each user, their signup cohort and the number of weeks the event comes after their signup date. Captured last 4 weeks */
			cohorts AS (
				SELECT
					DATE_TRUNC('week', created_at) AS cohort_date,
					COUNT(*) AS cohort_size
				FROM users
				WHERE created_at >= DATE_TRUNC('week', %s::timestamp) - INTERVAL '11 weeks'
				GROUP BY cohort_date
			),
			sub AS (
				SELECT
					event_logs.user_id AS user_id,
					DATE_TRUNC('week', users.created_at) AS cohort_date,
					FLOOR(DATE_PART('day', event_logs.timestamp::TIMESTAMP - users.created_at::TIMESTAMP)/7) AS weeks_after_signup
				FROM event_logs
				JOIN users ON (
					users.id = event_logs.user_id AND
					users.created_at >= DATE_TRUNC('week', %s::timestamp) - INTERVAL '11 weeks'
				)
				GROUP BY user_id, cohort_date, weeks_after_signup
			) /* calculate retention percentages for weeks 0-3 for each of the last 4 weekly cohorts */

		SELECT
			dates.week_start_date::DATE AS week,
			cohorts.cohort_size,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_after_signup = 0 AND cohorts.cohort_date = dates.week_start_date)::DECIMAL/cohorts.cohort_size,3) AS week_0,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_after_signup = 1 AND cohorts.cohort_date = dates.week_start_date)::DECIMAL/cohorts.cohort_size,3) AS week_1,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_after_signup = 2 AND cohorts.cohort_date = dates.week_start_date)::DECIMAL/cohorts.cohort_size,3) AS week_2,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_after_signup = 3 AND cohorts.cohort_date = dates.week_start_date)::DECIMAL/cohorts.cohort_size,3) AS week_3,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_after_signup = 4 AND cohorts.cohort_date = dates.week_start_date)::DECIMAL/cohorts.cohort_size,3) AS week_4,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_after_signup = 5 AND cohorts.cohort_date = dates.week_start_date)::DECIMAL/cohorts.cohort_size,3) AS week_5,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_after_signup = 6 AND cohorts.cohort_date = dates.week_start_date)::DECIMAL/cohorts.cohort_size,3) AS week_6,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_after_signup = 7 AND cohorts.cohort_date = dates.week_start_date)::DECIMAL/cohorts.cohort_size,3) AS week_7,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_after_signup = 8 AND cohorts.cohort_date = dates.week_start_date)::DECIMAL/cohorts.cohort_size,3) AS week_8,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_after_signup = 9 AND cohorts.cohort_date = dates.week_start_date)::DECIMAL/cohorts.cohort_size,3) AS week_9,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_after_signup = 10 AND cohorts.cohort_date = dates.week_start_date)::DECIMAL/cohorts.cohort_size,3) AS week_10,
			ROUND(COUNT(DISTINCT user_id) FILTER (WHERE weeks_after_signup = 11 AND cohorts.cohort_date = dates.week_start_date)::DECIMAL/cohorts.cohort_size,3) AS week_11
		FROM dates
		LEFT JOIN cohorts ON dates.week_start_date = cohorts.cohort_date
		LEFT JOIN sub     ON dates.week_start_date = sub.cohort_date
		GROUP BY week, cohorts.cohort_size
		ORDER BY week DESC;
		`, timeNow(), timeNow(), timeNow(), timeNow())

	rows, err := dbconn.Global.QueryContext(ctx, weeklyRetentionQuery.Query(sqlf.PostgresBindVar), weeklyRetentionQuery.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	weeklyRetentionCohorts := []*types.WeeklyRetentionStats{}
	for rows.Next() {
		var w types.WeeklyRetentionStats

		err := rows.Scan(
			&w.WeekStart,
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
		weeklyRetentionCohorts = append(weeklyRetentionCohorts, &w)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	var r types.RetentionStats
	r.Weekly = weeklyRetentionCohorts
	return &r, nil
}
