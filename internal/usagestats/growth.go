// Package usagestats provides an interface to update and access information about
// individual and aggregate Sourcegraph users' activity levels.
package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetGrowthStatistics(ctx context.Context, db database.DB) (*types.GrowthStatistics, error) {
	growthUsersStatistics, err := getUsersGrowthStatistics(ctx, db)
	if err != nil {
		return nil, err
	}

	growthAccessRequestsStatistics, err := getAccessRequestsGrowthStatistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return &types.GrowthStatistics{
		DeletedUsers:           int32(growthUsersStatistics.deletedUsers),
		CreatedUsers:           int32(growthUsersStatistics.createdUsers),
		ResurrectedUsers:       int32(growthUsersStatistics.resurrectedUsers),
		ChurnedUsers:           int32(growthUsersStatistics.churnedUsers),
		RetainedUsers:          int32(growthUsersStatistics.retainedUsers),
		PendingAccessRequests:  int32(growthAccessRequestsStatistics.pendingAccessRequests),
		ApprovedAccessRequests: int32(growthAccessRequestsStatistics.approvedAccessRequests),
		RejectedAccessRequests: int32(growthAccessRequestsStatistics.rejectedAccessRequests),
	}, nil
}

type usersGrowthStatistics struct {
	createdUsers     int
	deletedUsers     int
	resurrectedUsers int
	churnedUsers     int
	retainedUsers    int
}

func getUsersGrowthStatistics(ctx context.Context, db database.DB) (*usersGrowthStatistics, error) {
	const usersQuery = `
WITH active_last_month AS (
    SELECT DISTINCT user_id
    FROM event_logs
    WHERE timestamp > (DATE_TRUNC('month', $1::timestamp) - INTERVAL '1 month')
        AND timestamp < DATE_TRUNC('month', $1::timestamp)
),
active_this_month AS (
    SELECT DISTINCT user_id
    FROM event_logs
    WHERE timestamp > DATE_TRUNC('month', $1::timestamp)
        AND timestamp < (DATE_TRUNC('month', $1::timestamp) + INTERVAL '1 month')
),
recent_usage_by_user AS (
    SELECT users.id,
           EXISTS(SELECT * FROM active_this_month WHERE user_id = users.id) as current_month,
           EXISTS(SELECT * FROM active_last_month WHERE user_id = users.id) as previous_month,
           DATE_TRUNC('month', DATE(users.created_at)) AS created_month,
           DATE_TRUNC('month', DATE(users.deleted_at)) AS deleted_month
      FROM users
)
SELECT COUNT(*) FILTER ( WHERE recent_usage_by_user.created_month = DATE_TRUNC('month', $1::timestamp)) AS created_users,
       COUNT(*) FILTER ( WHERE recent_usage_by_user.deleted_month = DATE_TRUNC('month', $1::timestamp)) AS deleted_users,
       COUNT(*) FILTER (
                 WHERE current_month = TRUE
                   AND previous_month = FALSE
                   AND created_month < DATE_TRUNC('month', $1::timestamp)
                   AND (deleted_month < DATE_TRUNC('month', $1::timestamp) OR deleted_month IS NULL)) AS resurrected_users,
       COUNT(*) FILTER (
                 WHERE current_month = FALSE
                   AND previous_month = TRUE
                   AND created_month < DATE_TRUNC('month', $1::timestamp)
                   AND (deleted_month < DATE_TRUNC('month', $1::timestamp) OR deleted_month IS NULL)) AS churned_users,
       COUNT(*) FILTER (
                 WHERE current_month = TRUE
                   AND previous_month = TRUE
                   AND created_month < DATE_TRUNC('month', $1::timestamp)
                   AND (deleted_month < DATE_TRUNC('month', $1::timestamp) OR deleted_month IS NULL)) AS retained_users
  FROM recent_usage_by_user
    `
	var (
		createdUsers     int
		deletedUsers     int
		resurrectedUsers int
		churnedUsers     int
		retainedUsers    int
	)
	if err := db.QueryRowContext(ctx, usersQuery, timeNow()).Scan(
		&createdUsers,
		&deletedUsers,
		&resurrectedUsers,
		&churnedUsers,
		&retainedUsers,
	); err != nil {
		return nil, err
	}

	return &usersGrowthStatistics{
		deletedUsers:     deletedUsers,
		createdUsers:     createdUsers,
		resurrectedUsers: resurrectedUsers,
		churnedUsers:     churnedUsers,
		retainedUsers:    retainedUsers,
	}, nil
}

type accessRequestsGrowthStatistics struct {
	pendingAccessRequests  int
	approvedAccessRequests int
	rejectedAccessRequests int
}

func getAccessRequestsGrowthStatistics(ctx context.Context, db database.DB) (*accessRequestsGrowthStatistics, error) {
	const accessRequestsQuery = `
	SELECT
		COUNT(*) FILTER (WHERE status LIKE 'PENDING') AS pending_access_requests,
		COUNT(*) FILTER (WHERE status LIKE 'APPROVED') AS approved_access_requests,
		COUNT(*) FILTER (WHERE status LIKE 'REJECTED') AS rejected_access_requests
	FROM access_requests
	WHERE DATE_TRUNC('month', created_at) = DATE_TRUNC('month', $1::timestamp)
	`
	var (
		pendingAccessRequests  int
		approvedAccessRequests int
		rejectedAccessRequests int
	)
	if err := db.QueryRowContext(ctx, accessRequestsQuery, timeNow()).Scan(
		&pendingAccessRequests,
		&approvedAccessRequests,
		&rejectedAccessRequests,
	); err != nil {
		return nil, err
	}

	return &accessRequestsGrowthStatistics{
		pendingAccessRequests:  pendingAccessRequests,
		approvedAccessRequests: approvedAccessRequests,
		rejectedAccessRequests: rejectedAccessRequests,
	}, nil
}
