package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

func GetHomepagePanels(ctx context.Context) (*types.HomepagePanels, error) {
	const q = `
	WITH sub AS (SELECT name, user_id FROM event_logs WHERE name LIKE '%Panel%' AND DATE_TRUNC('week', timestamp) = DATE_TRUNC('week', current_date))
		SELECT
  CAST(COUNT(*) FILTER (WHERE name = 'RecentFilesPanelFileClicked') AS FLOAT)/NULLIF(COUNT(*) FILTER (WHERE name = 'RecentFilesPanelLoaded'),0) AS recentFilesClickedPercentage,
  CAST(COUNT(*) FILTER (WHERE name = 'RecentSearchesPanelSearchClicked') AS FLOAT)/NULLIF(COUNT(*) FILTER (WHERE name = 'RecentSearchesPanelLoaded'),0) AS recentSearchClickedPercentage,
  CAST(COUNT(*) FILTER (WHERE name = 'RepositoriesPanelRepoFilterClicked') AS FLOAT)/NULLIF(COUNT(*) FILTER (WHERE name = 'RepositoriesPanelLoaded'),0) AS recentRepositoriesClickedPercentage,
  CAST(COUNT(*) FILTER (WHERE name = 'SavedSearchesPanelSearchClicked') AS FLOAT)/NULLIF(COUNT(*) FILTER (WHERE name = 'SavedSearchesPanelLoaded'),0) AS savedSearchesClickedPercentage,
  CAST(COUNT(*) FILTER (WHERE name = 'SavedSearchesPanelCreateButtonClicked') AS FLOAT)/NULLIF(COUNT(*) FILTER (WHERE name = 'SavedSearchesPanelLoaded'),0) AS newSavedSearchesClickedPercentage,
  COUNT(*) FILTER (WHERE name = 'RecentSearchesPanelLoaded') as totalPanelViews,
  CAST(COUNT(*) FILTER (WHERE name = 'SavedSearchesPanelCreateButtonClicked') AS FLOAT)/NULLIF(COUNT(*) FILTER (WHERE name = 'SavedSearchesPanelLoaded'),0) AS newSavedSearchesClickedPercentage,
  CAST(COUNT(DISTINCT user_id) FILTER (WHERE name = 'RecentFilesPanelFileClicked') AS FLOAT)/NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE name = 'RecentFilesPanelLoaded'),0) AS usersFilesClickedPercentage,
  CAST(COUNT(DISTINCT user_id) FILTER (WHERE name = 'RecentSearchesPanelSearchClicked') AS FLOAT)/NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE name = 'RecentSearchesPanelLoaded'),0) AS usersSearchClickedPercentage,
  CAST(COUNT(DISTINCT user_id) FILTER (WHERE name = 'RepositoriesPanelRepoFilterClicked') AS FLOAT)/NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE name = 'RepositoriesPanelLoaded'),0) AS usersRepositoriesClickedPercentage,
  CAST(COUNT(DISTINCT user_id) FILTER (WHERE name = 'SavedSearchesPanelSearchClicked') AS FLOAT)/NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE name = 'SavedSearchesPanelLoaded'),0) AS usersSavedSearchesClickedPercentage,
  CAST(COUNT(DISTINCT user_id) FILTER (WHERE name = 'SavedSearchesPanelCreateButtonClicked') AS FLOAT)/NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE name = 'SavedSearchesPanelLoaded'),0) AS usersNewSavedSearchesClickedPercentage,
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
		&usersFilesClickedPercentage,
		&usersSearchClickedPercentage,
		&usersRepositoriesClickedPercentage,
		&usersSavedSearchesClickedPercentage,
		&usersNewSavedSearchesClickedPercentage,
		&percentUsersShown,
	); err != nil {
		return nil, err
	}

	return &types.HomepagePanels{
		RecentFilesClickedPercentage:           float64(recentFilesClickedPercentage),
		RecentSearchClickedPercentage:          float64(recentSearchClickedPercentage),
		RecentRepositoriesClickedPercentage:    float64(recentRepositoriesClickedPercentage),
		SavedSearchesClickedPercentage:         float64(savedSearchesClickedPercentage),
		NewSavedSearchesClickedPercentage:      float64(newSavedSearchesClickedPercentage),
		TotalPanelViews:                        float64(totalPanelViews),
		UsersFilesClickedPercentage:            float64(usersFilesClickedPercentage),
		UsersSearchClickedPercentage:           float64(usersSearchClickedPercentage),
		UsersRepositoriesClickedPercentage:     float64(usersRepositoriesClickedPercentage),
		UsersSavedSearchesClickedPercentage:    float64(usersSavedSearchesClickedPercentage),
		UsersNewSavedSearchesClickedPercentage: float64(usersNewSavedSearchesClickedPercentage),
		PercentUsersShown:                      float64(percentUsersShown),
	}, nil
}
