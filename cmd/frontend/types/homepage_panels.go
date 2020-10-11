package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

func GetHomepagePanels(ctx context.Context) (*types.GetHomepagePanels, error) {
	const q = `
	SELECT
  COUNT(*) FILTER (WHERE name = RecentFilesPanelFileClicked) FROM event_logs/COUNT(*) FILTER (WHERE name = RecentFilesPanelLoaded) FROM event_logs AS recentFilesClickedPercentage,
  COUNT(*) FILTER (WHERE name = RecentSearchesPanelSearchClicked) FROM event_logs/COUNT(*) FILTER (WHERE name = RecentSearchesPanelLoaded) FROM event_logs AS recentSearchClickedPercentage,
  COUNT(*) FILTER (WHERE name = RepositoriesPanelRepoFilterClicked) FROM event_logs/COUNT(*) FILTER (WHERE name = RepositoriesPanelLoaded) FROM event_logs AS recentRepositoriesClickedPercentage,
  COUNT(*) FILTER (WHERE name = SavedSearchesPanelSearchClicked) FROM event_logs/COUNT(*) FILTER (WHERE name = SavedSearchesPanelLoaded) FROM event_logs AS savedSearchesClickedPercentage,
  COUNT(*) FILTER (WHERE name = SavedSearchesPanelCreateButtonClicked) FROM event_logs/COUNT(*) FILTER (WHERE name = SavedSearchesPanelLoaded) FROM event_logs AS newSavedSearchesClickedPercentage,
  COUNT(*) FILTER (WHERE name = RecentSearchesPanelLoaded) FROM event_logs as totalShownPercentage,
  COUNT(DISTINCT user_id) FILTER (WHERE name = RecentFilesPanelFileClicked) FROM event_logs/COUNT(DISTINCT user_id) FILTER (WHERE name = RecentFilesPanelLoaded) FROM event_logs AS UsersFilesClickedPercentage,
  COUNT(DISTINCT user_id) FILTER (WHERE name = RecentSearchesPanelSearchClicked) FROM event_logs/COUNT(DISTINCT user_id) FILTER (WHERE name = RecentSearchesPanelLoaded) FROM event_logs AS UsersSearchClickedPercentage,
  COUNT(DISTINCT user_id) FILTER (WHERE name = RepositoriesPanelRepoFilterClicked) FROM event_logs/COUNT(DISTINCT user_id) FILTER (WHERE name = RepositoriesPanelLoaded) FROM event_logs AS UsersRepositoriesClickedPercentage,
  COUNT(DISTINCT user_id) FILTER (WHERE name = SavedSearchesPanelSearchClicked) FROM event_logs/COUNT(DISTINCT user_id) FILTER (WHERE name = SavedSearchesPanelLoaded) FROM event_logs AS UsersSavedSearchesClickedPercentage,
  COUNT(DISTINCT user_id) FILTER (WHERE name = SavedSearchesPanelCreateButtonClicked) FROM event_logs/COUNT(DISTINCT user_id) FILTER (WHERE name = SavedSearchesPanelLoaded) FROM event_logs AS UsersNewSavedSearchesClickedPercentage,
  COUNT(DISTINCT user_id) FILTER (WHERE name = RecentSearchesPanelLoaded) FROM event_logs AS percentUsersShown,
FROM
  event_logs`
	var (
		recentFilesClicked                        float32
		recentSearchClickedPercentage             float32
		recentRepositoriesClickedPercentage       float32
		savedSearchesClickedPercentage            float32
		newSavedSearchesClickedPercentage         float32
		totalShownPercentage                      float32
		UsersFilesClickedPercentage               float32
		UsersSearchClickedPercentage              float32
		UsersRepositoriesClickedPercentage        float32
		UsersSavedSearchesClickedPercentage       float32
		UsersNewSavedSearchesClickedPercentage    float32
		percentUsersShown                         float32
	)
	if err := dbconn.Global.QueryRowContext(ctx, q).Scan(
		&recentFilesClicked,
		&recentSearchClickedPercentage,
		&recentRepositoriesClickedPercentage,
		&savedSearchesClickedPercentage,
		&newSavedSearchesClickedPercentage,
		&totalShownPercentage,
		&UsersFilesClickedPercentage,
		&UsersSearchClickedPercentage,
		&UsersRepositoriesClickedPercentage,
		&UsersSavedSearchesClickedPercentage,
		&UsersNewSavedSearchesClickedPercentage,
		&percentUsersShown,
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
