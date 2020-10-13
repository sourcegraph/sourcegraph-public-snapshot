package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

func GetHomepagePanels(ctx context.Context) (*types.GetHomepagePanels, error) {
	const q = `
	WITH sub AS (SELECT name, user_id FROM event_logs WHERE name LIKE '%Panel%' AND DATE_TRUNC('week', timestamp) = DATE_TRUNC('week', current_date))
		SELECT
  CAST(COUNT(*) FILTER (WHERE name = 'RecentFilesPanelFileClicked') AS FLOAT)/COUNT(*) FILTER (WHERE name = 'RecentFilesPanelLoaded') AS recentFilesClickedPercentage,
  CAST(COUNT(*) FILTER (WHERE name = 'RecentSearchesPanelSearchClicked') AS FLOAT)/COUNT(*) FILTER (WHERE name = 'RecentSearchesPanelLoaded') AS recentSearchClickedPercentage,
  CAST(COUNT(*) FILTER (WHERE name = 'RepositoriesPanelRepoFilterClicked') AS FLOAT)/COUNT(*) FILTER (WHERE name = 'RepositoriesPanelLoaded') AS recentRepositoriesClickedPercentage,
  CAST(COUNT(*) FILTER (WHERE name = 'SavedSearchesPanelSearchClicked') AS FLOAT)/COUNT(*) FILTER (WHERE name = 'SavedSearchesPanelLoaded') AS savedSearchesClickedPercentage,
  CAST(COUNT(*) FILTER (WHERE name = 'SavedSearchesPanelCreateButtonClicked') AS FLOAT)/COUNT(*) FILTER (WHERE name = 'SavedSearchesPanelLoaded') AS newSavedSearchesClickedPercentage,
  COUNT(*) FILTER (WHERE name = 'RecentSearchesPanelLoaded') as totalPanelViews,
  CAST(COUNT(*) FILTER (WHERE name = 'SavedSearchesPanelCreateButtonClicked') AS FLOAT)/COUNT(*) FILTER (WHERE name = 'SavedSearchesPanelLoaded') AS newSavedSearchesClickedPercentage,
  CAST(COUNT(DISTINCT user_id) FILTER (WHERE name = 'RecentFilesPanelFileClicked') AS FLOAT)/COUNT(DISTINCT user_id) FILTER (WHERE name = 'RecentFilesPanelLoaded') AS UsersFilesClickedPercentage,
  CAST(COUNT(DISTINCT user_id) FILTER (WHERE name = 'RecentSearchesPanelSearchClicked') AS FLOAT)/COUNT(DISTINCT user_id) FILTER (WHERE name = 'RecentSearchesPanelLoaded') AS UsersSearchClickedPercentage,
  CAST(COUNT(DISTINCT user_id) FILTER (WHERE name = 'RepositoriesPanelRepoFilterClicked') AS FLOAT)/COUNT(DISTINCT user_id) FILTER (WHERE name = 'RepositoriesPanelLoaded') AS UsersRepositoriesClickedPercentage,
  CAST(COUNT(DISTINCT user_id) FILTER (WHERE name = 'SavedSearchesPanelSearchClicked') AS FLOAT)/COUNT(DISTINCT user_id) FILTER (WHERE name = 'SavedSearchesPanelLoaded') AS UsersSavedSearchesClickedPercentage,
  CAST(COUNT(DISTINCT user_id) FILTER (WHERE name = 'SavedSearchesPanelCreateButtonClicked') AS FLOAT)/COUNT(DISTINCT user_id) FILTER (WHERE name = 'SavedSearchesPanelLoaded') AS UsersNewSavedSearchesClickedPercentage,
  COUNT(DISTINCT user_id) FILTER (WHERE name = 'RecentSearchesPanelLoaded') AS percentUsersShown
FROM
  sub`
	var (
		recentFilesClickedPercentage              float32
		recentSearchClickedPercentage             float32
		recentRepositoriesClickedPercentage       float32
		savedSearchesClickedPercentage            float32
		newSavedSearchesClickedPercentage         float32
		totalPanelViews                           float32
		usersFilesClickedPercentage               float32
		usersSearchClickedPercentage              float32
		usersRepositoriesClickedPercentage        float32
		usersSavedSearchesClickedPercentage       float32
		usersNewSavedSearchesClickedPercentage    float32
		percentUsersShown                         float32
	)
	if err := dbconn.Global.QueryRowContext(ctx, q).Scan(
		&recentFilesClickedPercentage,
		&recentSearchClickedPercentage,
		&recentRepositoriesClickedPercentage,
		&savedSearchesClickedPercentage,
		&newSavedSearchesClickedPercentage,
		&totalPanelViews,
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
		RecentFilesClickedPercentage:           float32(recentFilesClickedPercentage),
		RecentSearchClickedPercentage:          float32(recentSearchClickedPercentage),
		RecentRepositoriesClickedPercentage:    float32(RecentRepositoriesClickedPercentage),
		SavedSearchesClickedPercentage:         float32(savedSearchesClickedPercentage),
		NewSavedSearchesClickedPercentage:      float32(newSavedSearchesClickedPercentage),
		TotalPanelViews:                        float32(totalPanelViews),
		UsersFilesClickedPercentage:            float32(usersFilesClickedPercentage),
		UsersSearchClickedPercentage:           float32(usersSearchClickedPercentage),
		UsersRepositoriesClickedPercentage:     float32(usersRepositoriesClickedPercentage),
		UsersSavedSearchesClickedPercentage:    float32(usersSavedSearchesClickedPercentage),
		UsersNewSavedSearchesClickedPercentage: float32(usersNewSavedSearchesClickedPercentage),
		PercentUsersShown:                      float32(PercentUsersShown)
	}, nil
}
