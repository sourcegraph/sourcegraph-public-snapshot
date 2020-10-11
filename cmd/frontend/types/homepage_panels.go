package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

func GetHomepagePanels(ctx context.Context) (*types.GetHomepagePanels, error) {
	const q = `
	SELECT
  COUNT(*) FILTER (WHERE name = RecentFilesPanelFileClicked) FROM event_logs/COUNT(*) FILTER (WHERE name = RecentFilesPanelLoaded) FROM event_logs AS recentFilesClicked,
  COUNT(*) FILTER (WHERE name = RecentSearchesPanelSearchClicked) FROM event_logs/COUNT(*) FILTER (WHERE name = RecentSearchesPanelLoaded) FROM event_logs AS recentSearchClicked,
  COUNT(*) FILTER (WHERE name = RepositoriesPanelRepoFilterClicked) FROM event_logs/COUNT(*) FILTER (WHERE name = RepositoriesPanelLoaded) FROM event_logs AS recentRepositoriesClicked,
  COUNT(*) FILTER (WHERE name = SavedSearchesPanelSearchClicked) FROM event_logs/COUNT(*) FILTER (WHERE name = SavedSearchesPanelLoaded) FROM event_logs AS savedSearchesClicked,
  COUNT(*) FILTER (WHERE name = SavedSearchesPanelCreateButtonClicked) FROM event_logs/COUNT(*) FILTER (WHERE name = SavedSearchesPanelLoaded) FROM event_logs AS newSavedSearchesClicked,
  COUNT(*) FILTER (WHERE name = RecentSearchesPanelLoaded) FROM event_logs as totalShown,
  COUNT(DISTINCT user_id) FILTER (WHERE name = RecentFilesPanelFileClicked) FROM event_logs/COUNT(DISTINCT user_id) FILTER (WHERE name = RecentFilesPanelLoaded) FROM event_logs AS percentUsersFilesClicked,
  COUNT(DISTINCT user_id) FILTER (WHERE name = RecentSearchesPanelSearchClicked) FROM event_logs/COUNT(DISTINCT user_id) FILTER (WHERE name = RecentSearchesPanelLoaded) FROM event_logs AS percentUsersSearchClicked,
  COUNT(DISTINCT user_id) FILTER (WHERE name = RepositoriesPanelRepoFilterClicked) FROM event_logs/COUNT(DISTINCT user_id) FILTER (WHERE name = RepositoriesPanelLoaded) FROM event_logs AS percentUsersRepositoriesClicked,
  COUNT(DISTINCT user_id) FILTER (WHERE name = SavedSearchesPanelSearchClicked) FROM event_logs/COUNT(DISTINCT user_id) FILTER (WHERE name = SavedSearchesPanelLoaded) FROM event_logs AS percentUserssavedSearchesClicked,
  COUNT(DISTINCT user_id) FILTER (WHERE name = SavedSearchesPanelCreateButtonClicked) FROM event_logs/COUNT(DISTINCT user_id) FILTER (WHERE name = SavedSearchesPanelLoaded) FROM event_logs AS percentUsersnewSavedSearchesClicked,
  COUNT(DISTINCT user_id) FILTER (WHERE name = RecentSearchesPanelLoaded) FROM event_logs AS percentUsersShown,
FROM
  event_logs`
	var (
		totalSavedSearches   int
		uniqueUsers          int
		notificationsSent    int
		notificationsClicked int
		uniqueUserPageViews  int
		orgSavedSearches     int
	)
	if err := dbconn.Global.QueryRowContext(ctx, q).Scan(
		&totalSavedSearches,
		&uniqueUsers,
		&notificationsSent,
		&notificationsClicked,
		&uniqueUserPageViews,
		&orgSavedSearches,
	); err != nil {
		return nil, err
	}

	return &types.SavedSearches{
		TotalSavedSearches:   int32(totalSavedSearches),
		UniqueUsers:          int32(uniqueUsers),
		NotificationsSent:    int32(notificationsSent),
		NotificationsClicked: int32(notificationsClicked),
		UniqueUserPageViews:  int32(uniqueUserPageViews),
		OrgSavedSearches:     int32(orgSavedSearches),
	}, nil
}
