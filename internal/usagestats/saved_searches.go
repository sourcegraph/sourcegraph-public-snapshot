package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type savedSearches struct {
	TotalSavedSearches int32

	UniqueUserSavedSearchOwners, UniqueUsers int32 // 2nd is the old field name for backcompat

	UniqueOrgSavedSearchOwners int32
	UserOwnedSavedSearches     int32

	OrgOwnedSavedSearches, OrgSavedSearches int32 // 2nd is the old field name for backcompat

	SavedSearchesCreatedLastUTCDay int32
	SavedSearchesUpdatedLastUTCDay int32

	NotificationsSent    int32
	NotificationsClicked int32
	UniqueUserPageViews  int32
}

func GetSavedSearches(ctx context.Context, db database.DB) (*savedSearches, error) {
	const q = `
	SELECT
	(SELECT COUNT(*) FROM saved_searches) AS totalSavedSearches,
	(SELECT COUNT(DISTINCT user_id) FROM saved_searches) AS uniqueUserSavedSearchOwners,
	(SELECT COUNT(DISTINCT org_id) FROM saved_searches) AS uniqueOrgSavedSearchOwners,
	(SELECT COUNT(*) FROM saved_searches WHERE user_id IS NOT NULL) AS userOwnedSavedSearches,
	(SELECT COUNT(*) FROM saved_searches WHERE org_id IS NOT NULL) AS orgOwnedSavedSearches,
	(SELECT COUNT(*) FROM saved_searches WHERE created_at >= DATE_TRUNC('day', NOW() AT TIME ZONE 'UTC') - INTERVAL '1 day' AND created_at < DATE_TRUNC('day', NOW() AT TIME ZONE 'UTC')) AS savedSearchesCreatedLastUTCDay,
	(SELECT COUNT(*) FROM saved_searches WHERE updated_at >= DATE_TRUNC('day', NOW() AT TIME ZONE 'UTC') - INTERVAL '1 day' AND updated_at < DATE_TRUNC('day', NOW() AT TIME ZONE 'UTC')) AS savedSearchesUpdatedLastUTCDay,
	(SELECT COUNT(*) FROM event_logs WHERE event_logs.name = 'SavedSearchEmailNotificationSent') AS notificationsSent,
	(SELECT COUNT(*) FROM event_logs WHERE event_logs.name = 'SavedSearchEmailClicked') AS notificationsClicked,
	(SELECT COUNT(DISTINCT user_id) FROM event_logs WHERE event_logs.name = 'ViewSavedSearchListPage') AS uniqueUserPageViews
	`
	var ss savedSearches
	if err := db.QueryRowContext(ctx, q).Scan(
		&ss.TotalSavedSearches,
		&ss.UniqueUserSavedSearchOwners,
		&ss.UniqueOrgSavedSearchOwners,
		&ss.UserOwnedSavedSearches,
		&ss.OrgOwnedSavedSearches,
		&ss.SavedSearchesCreatedLastUTCDay,
		&ss.SavedSearchesUpdatedLastUTCDay,
		&ss.NotificationsSent,
		&ss.NotificationsClicked,
		&ss.UniqueUserPageViews,
	); err != nil {
		return nil, err
	}

	// Set old field names for backcompat
	ss.UniqueUsers = ss.UniqueUserSavedSearchOwners
	ss.OrgSavedSearches = ss.OrgOwnedSavedSearches

	return &ss, nil
}
