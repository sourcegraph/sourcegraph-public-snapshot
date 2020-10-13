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
		recentFilesClickedPercentage              float64
		recentSearchClickedPercentage             float64
		recentRepositoriesClickedPercentage       float64
		savedSearchesClickedPercentage            float64
		newSavedSearchesClickedPercentage         float64
		totalPanelViews                           float64
		usersFilesClickedPercentage               float64
		usersSearchClickedPercentage              float64
		usersRepositoriesClickedPercentage        float64
		usersSavedSearchesClickedPercentage       float64
		usersNewSavedSearchesClickedPercentage    float64
		percentUsersShown                         float64
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

	return &types.HomepagePanels{
		RecentFilesClickedPercentage:           float64(recentFilesClickedPercentage),
		RecentSearchClickedPercentage:          float64(recentSearchClickedPercentage),
		RecentRepositoriesClickedPercentage:    float64(RecentRepositoriesClickedPercentage),
		SavedSearchesClickedPercentage:         float64(savedSearchesClickedPercentage),
		NewSavedSearchesClickedPercentage:      float64(newSavedSearchesClickedPercentage),
		TotalPanelViews:                        float64(totalPanelViews),
		UsersFilesClickedPercentage:            float64(usersFilesClickedPercentage),
		UsersSearchClickedPercentage:           float64(usersSearchClickedPercentage),
		UsersRepositoriesClickedPercentage:     float64(usersRepositoriesClickedPercentage),
		UsersSavedSearchesClickedPercentage:    float64(usersSavedSearchesClickedPercentage),
		UsersNewSavedSearchesClickedPercentage: float64(usersNewSavedSearchesClickedPercentage),
		PercentUsersShown:                      float64(PercentUsersShown)
	}, nil
}
