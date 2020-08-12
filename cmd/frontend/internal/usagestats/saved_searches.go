// Package usagestats provides an interface to update and access information about
// individual and aggregate Sourcegraph users' activity levels.
package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

func GetSavedSearches(ctx context.Context) (*types.SavedSearches, error) {
	const q = `
	SELECT
	  (SELECT COUNT(*) FROM saved_searches) AS totalSavedSearches,
	  (SELECT COUNT(DISTINCT user_id) FROM saved_searches) AS uniqueUsers,
	  COUNT(*) FILTER (WHERE event_logs.name = 'SavedSearchEmailNotificationSent') AS notificationsSent,
	  COUNT(*) FILTER (WHERE event_logs.name = 'SavedSearchEmailClicked') AS notificationsClicked,
	  COUNT(DISTINCT user_id) FILTER (WHERE event_logs.name = 'ViewSavedSearchListPage') AS uniqueUserPageViews
	FROM
	  event_logs
	`
	var (
		totalSavedSearches     	int
		uniqueUsers     	    int
		notificationsSent 	 	int
		notificationsClicked 	int
		uniqueUserPageViews    	int
	)
	if err := dbconn.Global.QueryRowContext(ctx, q).Scan(
		&totalSavedSearches,
		&uniqueUsers,
		&notificationsSent,
		&notificationsClicked,
		&uniqueUserPageViews,
	); err != nil {
		return nil, err
	}

	return &types.SavedSearches{
		totalSavedSearches:     int32(totalSavedSearches),
		uniqueUsers:     	    int32(uniqueUsers),
		notificationsSent: 	    int32(notificationsSent),
		notificationsClicked:   int32(notificationsClicked),
		uniqueUserPageViews:    int32(uniqueUserPageViews),
	}, nil
}