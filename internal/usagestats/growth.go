// Package usagestats provides an interface to update and access information about
// individual and aggregate Sourcegraph users' activity levels.
package usagestats

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetGrowthStatistics(ctx context.Context, db database.DB) (*types.GrowthStatistics, error) {
	const q = `
-- source: internal/usagestats/growth.go:GetGrowthStatistics
WITH all_usage_by_user_and_month AS (
    SELECT user_id,
           DATE_TRUNC('month', timestamp) AS month_active
      FROM event_logs
     GROUP BY user_id,
              month_active),
recent_usage_by_user AS (
    SELECT users.id,
           BOOL_OR(CASE WHEN DATE_TRUNC('month', month_active) = DATE_TRUNC('month', now()) THEN TRUE ELSE FALSE END) AS current_month,
           BOOL_OR(CASE WHEN DATE_TRUNC('month', month_active) = DATE_TRUNC('month', now()) - INTERVAL '1 month' THEN TRUE ELSE FALSE END) AS previous_month,
           DATE_TRUNC('month', DATE(users.created_at)) AS created_month,
           DATE_TRUNC('month', DATE(users.deleted_at)) AS deleted_month
      FROM users
      LEFT JOIN all_usage_by_user_and_month ON all_usage_by_user_and_month.user_id = users.id
     GROUP BY id,
              created_month,
              deleted_month)
SELECT COUNT(*) FILTER ( WHERE recent_usage_by_user.created_month = DATE_TRUNC('month', now())) AS created_users,
       COUNT(*) FILTER ( WHERE recent_usage_by_user.deleted_month = DATE_TRUNC('month', now())) AS deleted_users,
       COUNT(*) FILTER (
                 WHERE current_month = TRUE
                   AND previous_month = FALSE
                   AND created_month < DATE_TRUNC('month', now())
                   AND (deleted_month < DATE_TRUNC('month', now()) OR deleted_month IS NULL)) AS resurrected_users,
       COUNT(*) FILTER (
                 WHERE current_month = FALSE
                   AND previous_month = TRUE
                   AND created_month < DATE_TRUNC('month', now())
                   AND (deleted_month < DATE_TRUNC('month', now()) OR deleted_month IS NULL)) AS churned_users,
       COUNT(*) FILTER (
                 WHERE current_month = TRUE
                   AND previous_month = TRUE
                   AND created_month < DATE_TRUNC('month', now())
                   AND (deleted_month < DATE_TRUNC('month', now()) OR deleted_month IS NULL)) AS retained_users
  FROM recent_usage_by_user
    `
	var (
		createdUsers     int
		deletedUsers     int
		resurrectedUsers int
		churnedUsers     int
		retainedUsers    int
	)
	if err := db.QueryRowContext(ctx, q).Scan(
		&createdUsers,
		&deletedUsers,
		&resurrectedUsers,
		&churnedUsers,
		&retainedUsers,
	); err != nil {
		return nil, err
	}

	return &types.GrowthStatistics{
		DeletedUsers:     int32(deletedUsers),
		CreatedUsers:     int32(createdUsers),
		ResurrectedUsers: int32(resurrectedUsers),
		ChurnedUsers:     int32(churnedUsers),
		RetainedUsers:    int32(retainedUsers),
	}, nil
}

func GetCTAUsage(ctx context.Context, db database.DB) (*types.CTAUsage, error) {
	// aggregatedUserIDQueryFragment is a query fragment that can be used to canonicalize the
	// values of the user_id and anonymous_user_id fields (assumed in scope) int a unified value.
	const aggregatedUserIDQueryFragment = `
CASE WHEN user_id = 0
  -- It's faster to group by an int rather than text, so we convert
  -- the anonymous_user_id to an int, rather than the user_id to text.
  THEN ('x' || substr(md5(anonymous_user_id), 1, 8))::bit(32)::int
  ELSE user_id
END
`

	const q = `
 -- source: internal/usagestats/growth.go:GetCTAUsage
 WITH events_for_today AS (
     (SELECT name,
            ` + aggregatedUserIDQueryFragment + ` AS user_id,
            DATE_TRUNC('day', timestamp) AS day,
            argument->>'page' AS page
      FROM event_logs
     WHERE name IN ('InstallBrowserExtensionCTAShown',
                    'InstallBrowserExtensionCTAClicked',
                    'InstallIDEExtensionCTAShown',
                    'InstallIDEExtensionCTAClicked')
       AND argument->>'page' IN ('file', 'search')
       AND DATE_TRUNC('day', timestamp) = DATE_TRUNC('day', $1::timestamp)
     ) UNION ALL (
      SELECT NULL AS name,
             NULL AS user_id,
             DATE_TRUNC('day', $1::timestamp) AS day,
             NULL AS page
    )
 )
 SELECT day,
        COUNT(DISTINCT user_id) FILTER (WHERE name = 'InstallBrowserExtensionCTAShown' AND page = 'file'),
        COUNT(DISTINCT user_id) FILTER (WHERE name = 'InstallBrowserExtensionCTAClicked' AND page = 'file'),
        COUNT(DISTINCT user_id) FILTER (WHERE name = 'InstallBrowserExtensionCTAShown' AND page = 'search'),
        COUNT(DISTINCT user_id) FILTER (WHERE name = 'InstallBrowserExtensionCTAClicked' AND page = 'search'),
        COUNT(*) FILTER (WHERE name = 'InstallBrowserExtensionCTAShown' AND page = 'file'),
        COUNT(*) FILTER (WHERE name = 'InstallBrowserExtensionCTAClicked' AND page = 'file'),
        COUNT(*) FILTER (WHERE name = 'InstallBrowserExtensionCTAShown' AND page = 'search'),
        COUNT(*) FILTER (WHERE name = 'InstallBrowserExtensionCTAClicked' AND page = 'search'),
        COUNT(DISTINCT user_id) FILTER (WHERE name = 'InstallIDEExtensionCTAShown' AND page = 'file'),
        COUNT(DISTINCT user_id) FILTER (WHERE name = 'InstallIDEExtensionCTAClicked' AND page = 'file'),
        COUNT(DISTINCT user_id) FILTER (WHERE name = 'InstallIDEExtensionCTAShown' AND page = 'search'),
        COUNT(DISTINCT user_id) FILTER (WHERE name = 'InstallIDEExtensionCTAClicked' AND page = 'search'),
        COUNT(*) FILTER (WHERE name = 'InstallIDEExtensionCTAShown' AND page = 'file'),
        COUNT(*) FILTER (WHERE name = 'InstallIDEExtensionCTAClicked' AND page = 'file'),
        COUNT(*) FILTER (WHERE name = 'InstallIDEExtensionCTAShown' AND page = 'search'),
        COUNT(*) FILTER (WHERE name = 'InstallIDEExtensionCTAClicked' AND page = 'search')
   FROM events_for_today
  GROUP BY day
`

	var (
		day                        time.Time
		bextFilePageUserDisplays   int32
		bextFilePageUserClicks     int32
		bextSearchPageUserDisplays int32
		bextSearchPageUserClicks   int32
		bextFilePageDisplays       int32
		bextFilePageClicks         int32
		bextSearchPageDisplays     int32
		bextSearchPageClicks       int32
		ideFilePageUserDisplays    int32
		ideFilePageUserClicks      int32
		ideSearchPageUserDisplays  int32
		ideSearchPageUserClicks    int32
		ideFilePageDisplays        int32
		ideFilePageClicks          int32
		ideSearchPageDisplays      int32
		ideSearchPageClicks        int32
	)
	now := timeNow()
	if err := db.QueryRowContext(ctx, q, now).Scan(
		&day,
		&bextFilePageUserDisplays,
		&bextFilePageUserClicks,
		&bextSearchPageUserDisplays,
		&bextSearchPageUserClicks,
		&bextFilePageDisplays,
		&bextFilePageClicks,
		&bextSearchPageDisplays,
		&bextSearchPageClicks,
		&ideFilePageUserDisplays,
		&ideFilePageUserClicks,
		&ideSearchPageUserDisplays,
		&ideSearchPageUserClicks,
		&ideFilePageDisplays,
		&ideFilePageClicks,
		&ideSearchPageDisplays,
		&ideSearchPageClicks,
	); err != nil {
		return nil, err
	}

	return &types.CTAUsage{
		DailyBrowserExtensionCTA: types.FileAndSearchPageUserAndEventCounts{
			StartTime:             day,
			DisplayedOnFilePage:   types.UserAndEventCount{UserCount: bextFilePageUserDisplays, EventCount: bextFilePageDisplays},
			DisplayedOnSearchPage: types.UserAndEventCount{UserCount: bextSearchPageUserDisplays, EventCount: bextSearchPageDisplays},
			ClickedOnFilePage:     types.UserAndEventCount{UserCount: bextFilePageUserClicks, EventCount: bextFilePageClicks},
			ClickedOnSearchPage:   types.UserAndEventCount{UserCount: bextSearchPageUserClicks, EventCount: bextSearchPageClicks},
		},
		DailyIDEExtensionCTA: types.FileAndSearchPageUserAndEventCounts{
			StartTime:             day,
			DisplayedOnFilePage:   types.UserAndEventCount{UserCount: ideFilePageUserDisplays, EventCount: ideFilePageDisplays},
			DisplayedOnSearchPage: types.UserAndEventCount{UserCount: ideSearchPageUserDisplays, EventCount: ideSearchPageDisplays},
			ClickedOnFilePage:     types.UserAndEventCount{UserCount: ideFilePageUserClicks, EventCount: ideFilePageClicks},
			ClickedOnSearchPage:   types.UserAndEventCount{UserCount: ideSearchPageUserClicks, EventCount: ideSearchPageClicks},
		},
	}, nil
}
