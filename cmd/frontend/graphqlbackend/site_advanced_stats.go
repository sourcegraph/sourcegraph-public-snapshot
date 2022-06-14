package graphqlbackend

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var (
	changesetCreated = `
	SELECT CAST(event_logs.timestamp AS DATE) as date,
		COUNT(event_logs.*) as total_count,
		COUNT(DISTINCT event_logs.anonymous_user_id) as unique_count,
		COUNT(users.id) as registered_count
	FROM users
		RIGHT JOIN event_logs ON users.id = event_logs.user_id
		AND event_logs.name IN ('ViewSignIn')
	WHERE event_logs.timestamp %s
	GROUP BY date
	`
	eventQueries = map[string]string{
		"CHANGESETS_CREATED": changesetCreated,
		"CHANGESETS_MERGED":  changesetCreated,
	}
)

func (r *siteResolver) EventStatistics(ctx context.Context, args *struct {
	From   string
	To     string
	Events *[]string
}) (*siteEventStatisticsResolver, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	dateRangeCond, err := makeDateRangeCond(args.From, args.To)
	if err != nil {
		return nil, err
	}
	var queries []*sqlf.Query
	if args.Events == nil {
		for _, query := range eventQueries {
			queries = append(queries, sqlf.Sprintf(query, dateRangeCond))
		}
	} else {
		for _, event := range *args.Events {
			if query, ok := eventQueries[event]; ok {
				queries = append(queries, sqlf.Sprintf(query, dateRangeCond))
			} else {
				// TODO: throw error or ignore?
			}
		}
	}
	query := sqlf.Sprintf(`
		WITH result AS (%s)
		SELECT date,
			SUM(total_count),
			SUM(unique_count),
			SUM(registered_count)
		FROM result
		GROUP BY date
		ORDER BY date
		`, sqlf.Join(queries, "UNION ALL"))

	rows, err := r.db.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var nodes []*types.EventStatisticsNode
	for rows.Next() {
		var node types.EventStatisticsNode
		if err := rows.Scan(&node.Date, &node.TotalCount, &node.UniqueCount, &node.RegisteredCount); err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}
	return &siteEventStatisticsResolver{nodes}, nil
}

func makeDateRangeCond(from string, to string) (*sqlf.Query, error) {
	fromTime, err := time.Parse(time.RFC3339, from)
	if err != nil {
		return nil, err
	}
	toTime, err := time.Parse(time.RFC3339, to)
	if err != nil {
		return nil, err
	}
	return sqlf.Sprintf(`BETWEEN %s AND %s`, fromTime.Format(time.RFC3339), toTime.Format(time.RFC3339)), nil
}

type siteEventStatisticsResolver struct {
	nodes []*types.EventStatisticsNode
}

func (s *siteEventStatisticsResolver) Nodes() []*siteEventStatisticsNodeResolver {
	resolvers := make([]*siteEventStatisticsNodeResolver, 0, len(s.nodes))
	for _, node := range s.nodes {
		resolvers = append(resolvers, &siteEventStatisticsNodeResolver{node})
	}
	return resolvers
}

type siteEventStatisticsNodeResolver struct {
	node *types.EventStatisticsNode
}

func (s *siteEventStatisticsNodeResolver) Date() string {
	return s.node.Date.Format(time.RFC3339)
}

func (s *siteEventStatisticsNodeResolver) TotalCount() int32 {
	return s.node.TotalCount
}

func (s *siteEventStatisticsNodeResolver) UniqueCount() int32 {
	return s.node.UniqueCount
}

func (s *siteEventStatisticsNodeResolver) RegisteredCount() int32 {
	return s.node.RegisteredCount
}
