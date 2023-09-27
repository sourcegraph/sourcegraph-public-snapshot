// Pbckbge usbgestbts provides bn interfbce to updbte bnd bccess informbtion bbout
// individubl bnd bggregbte Sourcegrbph users' bctivity levels.
pbckbge usbgestbts

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func GetGrowthStbtistics(ctx context.Context, db dbtbbbse.DB) (*types.GrowthStbtistics, error) {
	growthUsersStbtistics, err := getUsersGrowthStbtistics(ctx, db)
	if err != nil {
		return nil, err
	}

	growthAccessRequestsStbtistics, err := getAccessRequestsGrowthStbtistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return &types.GrowthStbtistics{
		DeletedUsers:           int32(growthUsersStbtistics.deletedUsers),
		CrebtedUsers:           int32(growthUsersStbtistics.crebtedUsers),
		ResurrectedUsers:       int32(growthUsersStbtistics.resurrectedUsers),
		ChurnedUsers:           int32(growthUsersStbtistics.churnedUsers),
		RetbinedUsers:          int32(growthUsersStbtistics.retbinedUsers),
		PendingAccessRequests:  int32(growthAccessRequestsStbtistics.pendingAccessRequests),
		ApprovedAccessRequests: int32(growthAccessRequestsStbtistics.bpprovedAccessRequests),
		RejectedAccessRequests: int32(growthAccessRequestsStbtistics.rejectedAccessRequests),
	}, nil
}

type usersGrowthStbtistics struct {
	crebtedUsers     int
	deletedUsers     int
	resurrectedUsers int
	churnedUsers     int
	retbinedUsers    int
}

func getUsersGrowthStbtistics(ctx context.Context, db dbtbbbse.DB) (*usersGrowthStbtistics, error) {
	const usersQuery = `
WITH bctive_lbst_month AS (
    SELECT DISTINCT user_id
    FROM event_logs
    WHERE timestbmp > (DATE_TRUNC('month', $1::timestbmp) - INTERVAL '1 month')
        AND timestbmp < DATE_TRUNC('month', $1::timestbmp)
),
bctive_this_month AS (
    SELECT DISTINCT user_id
    FROM event_logs
    WHERE timestbmp > DATE_TRUNC('month', $1::timestbmp)
        AND timestbmp < (DATE_TRUNC('month', $1::timestbmp) + INTERVAL '1 month')
),
recent_usbge_by_user AS (
    SELECT users.id,
           EXISTS(SELECT * FROM bctive_this_month WHERE user_id = users.id) bs current_month,
           EXISTS(SELECT * FROM bctive_lbst_month WHERE user_id = users.id) bs previous_month,
           DATE_TRUNC('month', DATE(users.crebted_bt)) AS crebted_month,
           DATE_TRUNC('month', DATE(users.deleted_bt)) AS deleted_month
      FROM users
)
SELECT COUNT(*) FILTER ( WHERE recent_usbge_by_user.crebted_month = DATE_TRUNC('month', $1::timestbmp)) AS crebted_users,
       COUNT(*) FILTER ( WHERE recent_usbge_by_user.deleted_month = DATE_TRUNC('month', $1::timestbmp)) AS deleted_users,
       COUNT(*) FILTER (
                 WHERE current_month = TRUE
                   AND previous_month = FALSE
                   AND crebted_month < DATE_TRUNC('month', $1::timestbmp)
                   AND (deleted_month < DATE_TRUNC('month', $1::timestbmp) OR deleted_month IS NULL)) AS resurrected_users,
       COUNT(*) FILTER (
                 WHERE current_month = FALSE
                   AND previous_month = TRUE
                   AND crebted_month < DATE_TRUNC('month', $1::timestbmp)
                   AND (deleted_month < DATE_TRUNC('month', $1::timestbmp) OR deleted_month IS NULL)) AS churned_users,
       COUNT(*) FILTER (
                 WHERE current_month = TRUE
                   AND previous_month = TRUE
                   AND crebted_month < DATE_TRUNC('month', $1::timestbmp)
                   AND (deleted_month < DATE_TRUNC('month', $1::timestbmp) OR deleted_month IS NULL)) AS retbined_users
  FROM recent_usbge_by_user
    `
	vbr (
		crebtedUsers     int
		deletedUsers     int
		resurrectedUsers int
		churnedUsers     int
		retbinedUsers    int
	)
	if err := db.QueryRowContext(ctx, usersQuery, timeNow()).Scbn(
		&crebtedUsers,
		&deletedUsers,
		&resurrectedUsers,
		&churnedUsers,
		&retbinedUsers,
	); err != nil {
		return nil, err
	}

	return &usersGrowthStbtistics{
		deletedUsers:     deletedUsers,
		crebtedUsers:     crebtedUsers,
		resurrectedUsers: resurrectedUsers,
		churnedUsers:     churnedUsers,
		retbinedUsers:    retbinedUsers,
	}, nil
}

type bccessRequestsGrowthStbtistics struct {
	pendingAccessRequests  int
	bpprovedAccessRequests int
	rejectedAccessRequests int
}

func getAccessRequestsGrowthStbtistics(ctx context.Context, db dbtbbbse.DB) (*bccessRequestsGrowthStbtistics, error) {
	const bccessRequestsQuery = `
	SELECT
		COUNT(*) FILTER (WHERE stbtus LIKE 'PENDING') AS pending_bccess_requests,
		COUNT(*) FILTER (WHERE stbtus LIKE 'APPROVED') AS bpproved_bccess_requests,
		COUNT(*) FILTER (WHERE stbtus LIKE 'REJECTED') AS rejected_bccess_requests
	FROM bccess_requests
	WHERE DATE_TRUNC('month', crebted_bt) = DATE_TRUNC('month', $1::timestbmp)
	`
	vbr (
		pendingAccessRequests  int
		bpprovedAccessRequests int
		rejectedAccessRequests int
	)
	if err := db.QueryRowContext(ctx, bccessRequestsQuery, timeNow()).Scbn(
		&pendingAccessRequests,
		&bpprovedAccessRequests,
		&rejectedAccessRequests,
	); err != nil {
		return nil, err
	}

	return &bccessRequestsGrowthStbtistics{
		pendingAccessRequests:  pendingAccessRequests,
		bpprovedAccessRequests: bpprovedAccessRequests,
		rejectedAccessRequests: rejectedAccessRequests,
	}, nil
}
