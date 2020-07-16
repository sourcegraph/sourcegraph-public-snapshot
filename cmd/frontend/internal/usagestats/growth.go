// Package usagestats provides an interface to update and access information about
// individual and aggregate Sourcegraph users' activity levels.
package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

func GetGrowthStatistics(ctx context.Context) (*types.GrowthStatistics, error) {
	const q = `
	WITH
  latest_usage_by_user AS (
  SELECT
    user_id,
    DATE_TRUNC('month', timestamp) AS month_active
  FROM
    event_logs
  GROUP BY
    1,
    2 ),
  sub AS (
  SELECT
    DISTINCT users.id,
    CASE
      WHEN DATE_TRUNC('month', month_active) = DATE_TRUNC('month', now()) THEN TRUE
    ELSE
    FALSE
  END
    AS current_month,
    CASE
      WHEN DATE_TRUNC('month', month_active) = DATE_TRUNC('month', now()) - INTERVAL '1 month' THEN TRUE
    ELSE
    FALSE
  END
    AS previous_month,
    DATE_TRUNC('month', DATE(users.created_at)) AS created_month,
    DATE_TRUNC('month', DATE(users.deleted_at)) AS deleted_month
  FROM
    users
  LEFT JOIN
    latest_usage_by_user
  ON
    latest_usage_by_user.user_id = users.id)
SELECT
  COUNT(*) FILTER (
  WHERE
    sub.created_month = DATE_TRUNC('month', now())) AS created_users,
  COUNT(*) FILTER (
  WHERE
    sub.deleted_month = DATE_TRUNC('month', now())) AS deleted_users,
  COUNT(*) FILTER (
  WHERE
    current_month = TRUE
    AND previous_month = FALSE
    AND created_month < DATE_TRUNC('month', now())
    AND (deleted_month < DATE_TRUNC('month', now())
      OR deleted_month IS NULL)) AS resurrected_users,
  COUNT(*) FILTER (
  WHERE
    current_month = FALSE
    AND previous_month = TRUE
    AND created_month < DATE_TRUNC('month', now())
    AND (deleted_month < DATE_TRUNC('month', now())
      OR deleted_month IS NULL)) AS churned_users,
  COUNT(*) FILTER (
  WHERE
    current_month = TRUE
    AND previous_month = TRUE
    AND created_month < DATE_TRUNC('month', now())
    AND (deleted_month < DATE_TRUNC('month', now())
      OR deleted_month IS NULL)) AS retained_users
FROM
  sub
	`
	var (
		createdUsers     int
		deletedUsers     int
		resurrectedUsers int
		churnedUsers     int
		retainedUsers    int
	)
	if err := dbconn.Global.QueryRowContext(ctx, q).Scan(
		&createdUsers,
		&deletedUsers,
		&resurrectedUsers,
		&churnedUsers,
		&retainedUsers,
	); err != nil {
		return nil, err
	}

	return &types.GrowthStatistics{
		DeletedUsers:     int32(deletedUsers),
		CreatedUsers:     int32(createdUsers),
		ResurrectedUsers: int32(resurrectedUsers),
		ChurnedUsers:     int32(churnedUsers),
		RetainedUsers:    int32(retainedUsers),
	}, nil
}
