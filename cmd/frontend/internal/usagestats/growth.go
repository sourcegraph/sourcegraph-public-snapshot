// Package usagestats provides an interface to update and access information about
// individual and aggregate Sourcegraph users' activity levels.
package usagestats

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func GetDeletedUsers(ctx context.Context) (*types.MonthlyGrowthStatistics, error) {
	const q = `
	WITH
	sub AS (
	SELECT
	  DISTINCT users.id,
	  CASE
		WHEN DATE_TRUNC('month', timestamp) = DATE_TRUNC('month', now()) THEN TRUE
	  ELSE
	  FALSE
	END
	  AS current_month,
	  CASE
		WHEN DATE_TRUNC('month', timestamp) = DATE_TRUNC('month', now()) - INTERVAL '1 month' THEN TRUE
	  ELSE
	  FALSE
	END
	  AS previous_month,
	  DATE_TRUNC('month', DATE(users.created_at)) AS created_month,
	  DATE_TRUNC('month', DATE(users.deleted_at)) AS deleted_month
	FROM
	  users
	LEFT JOIN
	  event_logs
	ON
	  event_logs.user_id = users.id)
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
		OR deleted_month IS NULL)) AS churned_users
  FROM
	sub	
	`
	var (
		createdUsers int
		deletedUsers int
		resurrectedUsers int
		churnedUsers int
		retainedUsers int
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

	return &types.MonthlyGrowthStatistics{
		DeletedUsers:      	int32(deletedUsers),
		CreatedUsers:       int32(createdUsers),
		ResurrectedUsers: 	int32(resurrectedUsers),
		ChurnedUsers:       int32(churnedUsers),
		RetainedUsers:		int32(retainedUsers),
	}, nil
}