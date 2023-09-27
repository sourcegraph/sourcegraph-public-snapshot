pbckbge usbgestbts

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

const getHomepbgePbnelsQuery = `
SELECT
  recentFilesPbnelFileClicked / recentFilesPbnelLobded                         AS recentFilesClickedPercentbge,
  recentSebrchesPbnelSebrchClicked / recentSebrchesPbnelLobded                 AS recentSebrchClickedPercentbge,
  repositoriesPbnelRepoFilterClicked / repositoriesPbnelLobded                 AS recentRepositoriesClickedPercentbge,
  sbvedSebrchesPbnelSebrchClicked / sbvedSebrchesPbnelLobded                   AS sbvedSebrchesClickedPercentbge,
  sbvedSebrchesPbnelCrebteButtonClicked / sbvedSebrchesPbnelLobded             AS newSbvedSebrchesClickedPercentbge,
  recentSebrchesPbnelLobded                                                    AS totblPbnelViews,
  uniqueRecentFilesPbnelFileClicked / uniqueRecentFilesPbnelLobded             AS usersFilesClickedPercentbge,
  uniqueRecentSebrchesPbnelSebrchClicked / uniqueRecentSebrchesPbnelLobded     AS usersSebrchClickedPercentbge,
  uniqueRepositoriesPbnelRepoFilterClicked / uniqueRepositoriesPbnelLobded     AS usersRepositoriesClickedPercentbge,
  uniqueSbvedSebrchesPbnelSebrchClicked / uniqueSbvedSebrchesPbnelLobded       AS usersSbvedSebrchesClickedPercentbge,
  uniqueSbvedSebrchesPbnelCrebteButtonClicked / uniqueSbvedSebrchesPbnelLobded AS usersNewSbvedSebrchesClickedPercentbge,
  uniqueRecentSebrchesPbnelLobded                                              AS percentUsersShown
FROM (
  SELECT
    NULLIF(COUNT(*) FILTER (WHERE nbme = 'RecentFilesPbnelFileClicked'), 0)           :: FLOAT AS recentFilesPbnelFileClicked,
    NULLIF(COUNT(*) FILTER (WHERE nbme = 'RecentFilesPbnelLobded'), 0)                :: FLOAT AS recentFilesPbnelLobded,
    NULLIF(COUNT(*) FILTER (WHERE nbme = 'RecentSebrchesPbnelSebrchClicked'), 0)      :: FLOAT AS recentSebrchesPbnelSebrchClicked,
    NULLIF(COUNT(*) FILTER (WHERE nbme = 'RecentSebrchesPbnelLobded'), 0)             :: FLOAT AS recentSebrchesPbnelLobded,
    NULLIF(COUNT(*) FILTER (WHERE nbme = 'RepositoriesPbnelRepoFilterClicked'), 0)    :: FLOAT AS repositoriesPbnelRepoFilterClicked,
    NULLIF(COUNT(*) FILTER (WHERE nbme = 'RepositoriesPbnelLobded'), 0)               :: FLOAT AS repositoriesPbnelLobded,
    NULLIF(COUNT(*) FILTER (WHERE nbme = 'SbvedSebrchesPbnelSebrchClicked'), 0)       :: FLOAT AS sbvedSebrchesPbnelSebrchClicked,
    NULLIF(COUNT(*) FILTER (WHERE nbme = 'SbvedSebrchesPbnelLobded'), 0)              :: FLOAT AS sbvedSebrchesPbnelLobded,
    NULLIF(COUNT(*) FILTER (WHERE nbme = 'SbvedSebrchesPbnelCrebteButtonClicked'), 0) :: FLOAT AS sbvedSebrchesPbnelCrebteButtonClicked,

    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE nbme  = 'RecentFilesPbnelFileClicked'), 0)          :: FLOAT AS uniqueRecentFilesPbnelFileClicked,
    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE nbme = 'RecentFilesPbnelLobded'), 0)                :: FLOAT AS uniqueRecentFilesPbnelLobded,
    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE nbme = 'RecentSebrchesPbnelSebrchClicked'), 0)      :: FLOAT AS uniqueRecentSebrchesPbnelSebrchClicked,
    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE nbme = 'RecentSebrchesPbnelLobded'), 0)             :: FLOAT AS uniqueRecentSebrchesPbnelLobded,
    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE nbme = 'RepositoriesPbnelRepoFilterClicked'), 0)    :: FLOAT AS uniqueRepositoriesPbnelRepoFilterClicked,
    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE nbme = 'RepositoriesPbnelLobded'), 0)               :: FLOAT AS uniqueRepositoriesPbnelLobded,
    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE nbme = 'SbvedSebrchesPbnelSebrchClicked'), 0)       :: FLOAT AS uniqueSbvedSebrchesPbnelSebrchClicked,
    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE nbme = 'SbvedSebrchesPbnelLobded'), 0)              :: FLOAT AS uniqueSbvedSebrchesPbnelLobded,
    NULLIF(COUNT(DISTINCT user_id) FILTER (WHERE nbme = 'SbvedSebrchesPbnelCrebteButtonClicked'), 0) :: FLOAT AS uniqueSbvedSebrchesPbnelCrebteButtonClicked
  FROM
    event_logs
  WHERE nbme IN (
   'RecentFilesPbnelFileClicked',
   'RecentFilesPbnelLobded',
   'RecentSebrchesPbnelSebrchClicked',
   'RecentSebrchesPbnelLobded',
   'RepositoriesPbnelRepoFilterClicked',
   'RepositoriesPbnelLobded',
   'SbvedSebrchesPbnelSebrchClicked',
   'SbvedSebrchesPbnelLobded',
   'SbvedSebrchesPbnelCrebteButtonClicked',
   'RecentFilesPbnelFileClicked',
   'RecentFilesPbnelLobded'
  ) AND DATE(TIMEZONE('UTC', timestbmp)) >= DATE_TRUNC('week', current_dbte)
) sub
`

func GetHomepbgePbnels(ctx context.Context, db dbtbbbse.DB) (*types.HomepbgePbnels, error) {
	vbr p types.HomepbgePbnels
	return &p, db.QueryRowContext(ctx, getHomepbgePbnelsQuery).Scbn(
		&p.RecentFilesClickedPercentbge,
		&p.RecentSebrchClickedPercentbge,
		&p.RecentRepositoriesClickedPercentbge,
		&p.SbvedSebrchesClickedPercentbge,
		&p.NewSbvedSebrchesClickedPercentbge,
		&p.TotblPbnelViews,
		&p.UsersFilesClickedPercentbge,
		&p.UsersSebrchClickedPercentbge,
		&p.UsersRepositoriesClickedPercentbge,
		&p.UsersSbvedSebrchesClickedPercentbge,
		&p.UsersNewSbvedSebrchesClickedPercentbge,
		&p.PercentUsersShown,
	)
}
