package graphqlbackend

import (
	"context"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type siteAnalyticsResolver struct {
	db database.DB
}
type siteAnalyticsSearchResolver struct {
	dateRange string
	db        database.DB
}

type siteAnalyticsStatItemResolver struct {
	nodesQuery   *sqlf.Query
	summaryQuery *sqlf.Query
	db           database.DB
}

func (r *siteResolver) Analytics(ctx context.Context) (*siteAnalyticsResolver, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	return &siteAnalyticsResolver{r.db}, nil
}

func (r *siteAnalyticsResolver) Search(ctx context.Context, args *struct {
	DateRange *string
}) (*siteAnalyticsSearchResolver, error) {
	return &siteAnalyticsSearchResolver{dateRange: *args.DateRange, db: r.db}, nil
}

func makeDateParameters(dateRange string, dateColumnName string) (*sqlf.Query, *sqlf.Query, error) {
	now := time.Now()
	var from time.Time
	var groupBy string

	if dateRange == "LAST_THREE_MONTHS" {
		from = now.AddDate(0, -3, 0)
		groupBy = "week"
	} else if dateRange == "LAST_MONTH" {
		from = now.AddDate(0, -1, 0)
		groupBy = "day"
	} else if dateRange == "LAST_WEEK" {
		from = now.AddDate(0, 0, -7)
		groupBy = "day"
	} else {
		return nil, nil, errors.New("Invalid date range")
	}

	return sqlf.Sprintf(fmt.Sprintf(`date_trunc('%s', %s::date)`, groupBy, dateColumnName)), sqlf.Sprintf(`BETWEEN %s AND %s`, from.Format(time.RFC3339), now.Format(time.RFC3339)), nil
}

func (r *siteAnalyticsSearchResolver) Searches(ctx context.Context) (*siteAnalyticsStatItemResolver, error) {
	dateSelectParam, dateRangeCond, err := makeDateParameters(r.dateRange, "event_logs.timestamp")
	if err != nil {
		return nil, err
	}
	nodesQuery := sqlf.Sprintf(`
		SELECT %s AS date,
			COUNT(event_logs.*) AS total_count,
			COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
			COUNT(DISTINCT users.id) AS registered_users
		FROM users
			RIGHT JOIN event_logs ON users.id = event_logs.user_id
			AND event_logs.name IN ('SearchResultsQueried')
		WHERE event_logs.timestamp %s
		GROUP BY date
	`, dateSelectParam, dateRangeCond)

	summaryQuery := sqlf.Sprintf(`
		SELECT
			COUNT(event_logs.*) AS total_count,
			COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
			COUNT(DISTINCT users.id) AS registered_users
		FROM users
			RIGHT JOIN event_logs ON users.id = event_logs.user_id
			AND event_logs.name IN ('SearchResultsQueried')
		WHERE event_logs.timestamp %s
	`, dateRangeCond)

	return &siteAnalyticsStatItemResolver{nodesQuery: nodesQuery, summaryQuery: summaryQuery, db: r.db}, nil
}

// TODO: fileViews: AnalyticsStatItem!
// TODO: fileOpens: AnalyticsStatItem!
// TODO: sharedSearches: AnalyticsStatItem!

type AnalyticsStatItemNodeResolver struct {
	date            time.Time
	count           int32
	uniqueUsers     int32
	registeredUsers int32
}

func (r *siteAnalyticsStatItemResolver) Nodes(ctx context.Context) ([]*AnalyticsStatItemNodeResolver, error) {
	rows, err := r.db.QueryContext(ctx, r.nodesQuery.Query(sqlf.PostgresBindVar), r.nodesQuery.Args()...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	resolvers := make([]*AnalyticsStatItemNodeResolver, 0)
	for rows.Next() {
		var date time.Time
		var count, uniqueUsers, registeredUsers int32

		if err := rows.Scan(&date, &count, &uniqueUsers, &registeredUsers); err != nil {
			return nil, err
		}

		resolvers = append(resolvers, &AnalyticsStatItemNodeResolver{
			date:            date,
			count:           count,
			uniqueUsers:     uniqueUsers,
			registeredUsers: registeredUsers,
		})
	}

	return resolvers, nil
}

func (r *AnalyticsStatItemNodeResolver) Date() string {
	return r.date.Format(time.RFC3339)
}

func (r *AnalyticsStatItemNodeResolver) Count() int32 {
	return r.count
}

func (r *AnalyticsStatItemNodeResolver) UniqueUsers() int32 {
	return r.uniqueUsers
}

func (r *AnalyticsStatItemNodeResolver) RegisteredUsers() int32 {
	return r.registeredUsers
}

type AnalyticsStatItemSummaryResolver struct {
	totalCount           int32
	totalUniqueUsers     int32
	totalRegisteredUsers int32
}

func (s *AnalyticsStatItemSummaryResolver) TotalCount() (int32, error) {
	return s.totalCount, nil
}

func (s *AnalyticsStatItemSummaryResolver) TotalUniqueUsers() (int32, error) {
	return s.totalUniqueUsers, nil
}

func (s *AnalyticsStatItemSummaryResolver) TotalRegisteredUsers() (int32, error) {
	return s.totalRegisteredUsers, nil
}

func (r *siteAnalyticsStatItemResolver) Summary(ctx context.Context) (*AnalyticsStatItemSummaryResolver, error) {
	var totalCount, totalUniqueUsers, totalRegisteredUsers int32

	if err := r.db.QueryRowContext(ctx, r.summaryQuery.Query(sqlf.PostgresBindVar), r.summaryQuery.Args()...).Scan(&totalCount, &totalUniqueUsers, &totalRegisteredUsers); err != nil {
		return nil, err
	}

	return &AnalyticsStatItemSummaryResolver{
		totalCount:           totalCount,
		totalUniqueUsers:     totalUniqueUsers,
		totalRegisteredUsers: totalRegisteredUsers,
	}, nil
}
