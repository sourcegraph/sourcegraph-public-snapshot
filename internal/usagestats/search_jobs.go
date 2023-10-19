package usagestats

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type evenLoadFunc func(ctx context.Context, db database.DB, stats *types.SearchJobsUsageStatistics, now time.Time) error

type eventLoader struct {
	now        time.Time
	operations map[string]evenLoadFunc
}

func newEventLoader(now time.Time) *eventLoader {
	return &eventLoader{now: now, operations: make(map[string]evenLoadFunc)}
}

func (p *eventLoader) withOperation(name string, loadFunc evenLoadFunc) {
	p.operations[name] = loadFunc
}

func (p *eventLoader) generate(ctx context.Context, db database.DB) *types.SearchJobsUsageStatistics {
	stats := &types.SearchJobsUsageStatistics{}
	logger := log.Scoped("search jobs ping loader")

	for name, loadFunc := range p.operations {
		err := loadFunc(ctx, db, stats, p.now)
		if err != nil {
			logger.Error("search jobs pings loading error, skipping ping", log.String("name", name), log.Error(err))
		}
	}
	return stats
}

func GetSearchJobsUsageStatistics(ctx context.Context, db database.DB) (*types.SearchJobsUsageStatistics, error) {
	loader := newEventLoader(timeNow())

	loader.withOperation("weeklyUsage", weeklySearchJobsUsage)
	loader.withOperation("bannerViews", GetWeeklySearchFormViews)
	loader.withOperation("validationErrors", GetWeeklySearchFormValidationErrors)

	return loader.generate(ctx, db), nil
}

func weeklySearchJobsUsage(ctx context.Context, db database.DB, stats *types.SearchJobsUsageStatistics, now time.Time) error {
	const searchJobsWeeklyEventsQuery = `
    SELECT
		COUNT(*) FILTER (WHERE name = 'ViewSearchJobsListPage')                       	AS weekly_search_jobs_page_views,
		COUNT(*) FILTER (WHERE name = 'SearchJobsCreateClick')                       	AS weekly_search_jobs_create_clicks,
		COUNT(*) FILTER (WHERE name = 'SearchJobsResultDownloadClick') 				    AS weekly_search_jobs_download_clicks,
		COUNT(*) FILTER (WHERE name = 'SearchJobsResultViewLogsClick') 				    AS weekly_search_jobs_view_logs_clicks,
		COUNT(distinct user_id) FILTER (WHERE name = 'ViewSearchJobsListPage')        	AS weekly_search_jobs_unique_page_views,
		COUNT(distinct user_id) FILTER (WHERE name = 'SearchJobsResultDownloadClick')  	AS weekly_search_jobs_unique_download_clicks,
		COUNT(distinct user_id) FILTER (WHERE name = 'SearchJobsResultViewLogsClick') 	AS weekly_search_jobs_unique_view_logs_clicks
	FROM event_logs
	WHERE name in ('ViewSearchJobsListPage', 'SearchJobsCreateClick', 'SearchJobsResultDownloadClick', 'SearchJobsResultViewLogsClick')
		AND timestamp > DATE_TRUNC('week', $1::timestamp);
	`

	if err := db.QueryRowContext(ctx, searchJobsWeeklyEventsQuery, timeNow()).Scan(
		&stats.WeeklySearchJobsPageViews,
		&stats.WeeklySearchJobsCreateClick,
		&stats.WeeklySearchJobsDownloadClicks,
		&stats.WeeklySearchJobsViewLogsClicks,
		&stats.WeeklySearchJobsUniquePageViews,
		&stats.WeeklySearchJobsUniqueDownloadClicks,
		&stats.WeeklySearchJobsUniqueViewLogsClicks,
	); err != nil {
		return err
	}
	return nil
}

func GetWeeklySearchFormViews(ctx context.Context, db database.DB, stats *types.SearchJobsUsageStatistics, now time.Time) error {
	const getWeeklySearchFormViewsQuery = `
		SELECT COUNT(*), argument::json->>'validState' as argument FROM event_logs
		WHERE name = 'SearchJobsSearchFormShown' AND timestamp > DATE_TRUNC('week', $1::TIMESTAMP)
		GROUP BY argument;
	`
	rows, err := db.QueryContext(ctx, getWeeklySearchFormViewsQuery, timeNow())
	weeklySearchJobsSearchFormShownByValidState := []types.SearchJobsSearchFormShownPing{}

	if err != nil {
		return errors.Wrap(err, "GetWeeklySearchFormViews")
	}
	defer rows.Close()

	for rows.Next() {
		weeklySearchJobsSearchFormShown := types.SearchJobsSearchFormShownPing{}
		if err := rows.Scan(
			&weeklySearchJobsSearchFormShown.TotalCount,
			&weeklySearchJobsSearchFormShown.ValidState,
		); err != nil {
			return errors.Wrap(err, "GetWeeklySearchFormViews")
		}
		weeklySearchJobsSearchFormShownByValidState = append(weeklySearchJobsSearchFormShownByValidState, weeklySearchJobsSearchFormShown)
	}

	stats.WeeklySearchJobsSearchFormShown = weeklySearchJobsSearchFormShownByValidState

	return nil
}

func GetWeeklySearchFormValidationErrors(ctx context.Context, db database.DB, stats *types.SearchJobsUsageStatistics, now time.Time) error {
	const getSearchJobsAggregatedQuery = `
		SELECT COUNT(*) as count, argument::json->>'errors' as errors FROM event_logs
		WHERE name = 'SearchJobsValidationErrors' AND timestamp > DATE_TRUNC('week', $1::TIMESTAMP)
		GROUP BY errors
		ORDER BY count DESC, errors
	`

	rows, err := db.QueryContext(ctx, getSearchJobsAggregatedQuery, timeNow())
	if err != nil {
		return errors.Wrap(err, "GetWeeklySearchFormValidationErrors")
	}
	defer rows.Close()

	errorsAggregate := []types.SearchJobsValidationErrorPing{}
	for rows.Next() {
		var v types.SearchJobsValidationErrorPing
		if err := rows.Scan(
			&v.TotalCount,
			dbutil.JSONMessage(&v.Errors),
		); err != nil {
			return errors.Wrap(err, "GetWeeklySearchFormViews")
		}

		errorsAggregate = append(errorsAggregate, v)
	}

	stats.WeeklySearchJobsValidationErrors = errorsAggregate

	return errors.Wrap(rows.Err(), "GetWeeklySearchFormViews")
}
