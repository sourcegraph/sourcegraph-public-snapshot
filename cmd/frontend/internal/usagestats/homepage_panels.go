package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

const getHomepagePanelsQuery = `
-- source: cmd/frontend/internal/usagestats/homepage_panels.go:GetHomePagePanels
SELECT
  recentFilesPanelFileClicked / recentFilesPanelLoaded                         AS recentFilesClickedPercentage,
  recentSearchesPanelSearchClicked / recentSearchesPanelLoaded                 AS recentSearchClickedPercentage,
  repositoriesPanelRepoFilterClicked / repositoriesPanelLoaded                 AS recentRepositoriesClickedPercentage,
  savedSearchesPanelSearchClicked / savedSearchesPanelLoaded                   AS savedSearchesClickedPercentage,
  savedSearchesPanelCreateButtonClicked / savedSearchesPanelLoaded             AS newSavedSearchesClickedPercentage,
  recentSearchesPanelLoaded                                                    AS totalPanelViews,
  recentFilesPanelFileClicked / uniqueRecentFilesPanelLoaded                   AS usersFilesClickedPercentage,
  uniqueRecentSearchesPanelSearchClicked / uniqueRecentSearchesPanelLoaded     AS usersSearchClickedPercentage,
  uniqueRepositoriesPanelRepoFilterClicked / uniqueRepositoriesPanelLoaded     AS usersRepositoriesClickedPercentage,
  uniqueSavedSearchesPanelSearchClicked / uniqueSavedSearchesPanelLoaded       AS usersSavedSearchesClickedPercentage,
  uniqueSavedSearchesPanelCreateButtonClicked / uniqueSavedSearchesPanelLoaded AS usersNewSavedSearchesClickedPercentage,
  uniqueRecentSearchesPanelLoaded                                              AS percentUsersShown
FROM (
  SELECT
    NULLIF(COUNT(*) FILTER (WHERE name = 'RecentFilesPanelFileClicked'), 0)           :: FLOAT AS recentFilesPanelFileClicked,
    NULLIF(COUNT(*) FILTER (WHERE name = 'RecentFilesPanelLoaded'), 0)                :: FLOAT AS recentFilesPanelLoaded,
    NULLIF(COUNT(*) FILTER (WHERE name = 'RecentSearchesPanelSearchClicked'), 0)      :: FLOAT AS recentSearchesPanelSearchClicked,
    NULLIF(COUNT(*) FILTER (WHERE name = 'RecentSearchesPanelLoaded'), 0)             :: FLOAT AS recentSearchesPanelLoaded,
    NULLIF(COUNT(*) FILTER (WHERE name = 'RepositoriesPanelRepoFilterClicked'), 0)    :: FLOAT AS repositoriesPanelRepoFilterClicked,
    NULLIF(COUNT(*) FILTER (WHERE name = 'RepositoriesPanelLoaded'), 0)               :: FLOAT AS repositoriesPanelLoaded,
    NULLIF(COUNT(*) FILTER (WHERE name = 'SavedSearchesPanelSearchClicked'), 0)       :: FLOAT AS savedSearchesPanelSearchClicked,
    NULLIF(COUNT(*) FILTER (WHERE name = 'SavedSearchesPanelLoaded'), 0)              :: FLOAT AS savedSearchesPanelLoaded,
    NULLIF(COUNT(*) FILTER (WHERE name = 'SavedSearchesPanelCreateButtonClicked'), 0) :: FLOAT AS savedSearchesPanelCreateButtonClicked,

    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE name = 'RecentFilesPanelLoaded'), 0)                :: FLOAT AS uniqueRecentFilesPanelLoaded,
    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE name = 'RecentSearchesPanelSearchClicked'), 0)      :: FLOAT AS uniqueRecentSearchesPanelSearchClicked,
    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE name = 'RecentSearchesPanelLoaded'), 0)             :: FLOAT AS uniqueRecentSearchesPanelLoaded,
    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE name = 'RepositoriesPanelRepoFilterClicked'), 0)    :: FLOAT AS uniqueRepositoriesPanelRepoFilterClicked,
    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE name = 'RepositoriesPanelLoaded'), 0)               :: FLOAT AS uniqueRepositoriesPanelLoaded,
    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE name = 'SavedSearchesPanelSearchClicked'), 0)       :: FLOAT AS uniqueSavedSearchesPanelSearchClicked,
    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE name = 'SavedSearchesPanelLoaded'), 0)              :: FLOAT AS uniqueSavedSearchesPanelLoaded,
    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE name = 'SavedSearchesPanelCreateButtonClicked'), 0) :: FLOAT AS uniqueSavedSearchesPanelCreateButtonClicked 
  FROM
    event_logs
  WHERE name IN (
   'RecentFilesPanelFileClicked',
   'RecentFilesPanelLoaded',
   'RecentSearchesPanelSearchClicked',
   'RecentSearchesPanelLoaded',
   'RepositoriesPanelRepoFilterClicked',
   'RepositoriesPanelLoaded',
   'SavedSearchesPanelSearchClicked',
   'SavedSearchesPanelLoaded',
   'SavedSearchesPanelCreateButtonClicked',
   'RecentFilesPanelFileClicked',
   'RecentFilesPanelLoaded'
  ) AND DATE_TRUNC('week', DATE(TIMEZONE('UTC', timestamp))) = DATE_TRUNC('week', current_date)
) sub
`

func GetHomepagePanels(ctx context.Context) (*types.HomepagePanels, error) {
	var (
		recentFilesClickedPercentage           float64
		recentSearchClickedPercentage          float64
		recentRepositoriesClickedPercentage    float64
		savedSearchesClickedPercentage         float64
		newSavedSearchesClickedPercentage      float64
		totalPanelViews                        float64
		usersFilesClickedPercentage            float64
		usersSearchClickedPercentage           float64
		usersRepositoriesClickedPercentage     float64
		usersSavedSearchesClickedPercentage    float64
		usersNewSavedSearchesClickedPercentage float64
		percentUsersShown                      float64
	)
	if err := dbconn.Global.QueryRowContext(ctx, getHomepagePanelsQuery).Scan(
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
