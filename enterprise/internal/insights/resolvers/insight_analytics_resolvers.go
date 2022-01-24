package resolvers

import (
	"context"
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var _ graphqlbackend.CompareTwoInsightSeriesDataPointResolver = &compareTwoInsightSeriesDataPointResolver{}

type compareTwoInsightSeriesDataPointResolver struct {
	dataPoint compareTwoInsightSeriesDataPoint
}

func (c *compareTwoInsightSeriesDataPointResolver) FirstSeriesValue() *int32 {
	return c.dataPoint.FirstSeriesValue
}

func (c *compareTwoInsightSeriesDataPointResolver) SecondSeriesValue() *int32 {
	return c.dataPoint.SecondSeriesValue
}

func (c *compareTwoInsightSeriesDataPointResolver) Diff() int32 {
	return c.dataPoint.Diff
}

func (c *compareTwoInsightSeriesDataPointResolver) RepoName() string {
	return c.dataPoint.RepoName
}

func (c *compareTwoInsightSeriesDataPointResolver) Time() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: c.dataPoint.Time}
}

type compareTwoInsightSeriesDataPoint struct {
	FirstSeriesValue  *int32
	SecondSeriesValue *int32
	Diff              int32
	RepoName          string
	Time              time.Time
}

func (r *Resolver) CompareTwoInsightSeries(ctx context.Context, args graphqlbackend.CompareTwoInsightSeriesArgs) ([]graphqlbackend.CompareTwoInsightSeriesDataPointResolver, error) {
	var resolvers []graphqlbackend.CompareTwoInsightSeriesDataPointResolver

	// resolvers = append(resolvers, &compareTwoInsightSeriesDataPointResolver{dataPoint: compareTwoInsightSeriesDataPoint{
	// 	FirstSeriesValue:  nilInt(5),
	// 	SecondSeriesValue: nilInt(8),
	// 	Diff:              3,
	// 	RepoName:          "github.com/sourcegraph/sourcegraph",
	// 	Time:              graphqlbackend.DateTime{Time: time.Now()},
	// }})
	//
	// resolvers = append(resolvers, &compareTwoInsightSeriesDataPointResolver{dataPoint: compareTwoInsightSeriesDataPoint{
	// 	FirstSeriesValue:  nilInt(2),
	// 	SecondSeriesValue: nilInt(2),
	// 	Diff:              0,
	// 	RepoName:          "github.com/sourcegraph/handbook",
	// 	Time:              graphqlbackend.DateTime{Time: time.Now()},
	// }})

	points, err := compareTwoSeriesQuery(ctx, args.Input, r.baseInsightResolver)
	if err != nil {
		return nil, errors.Wrap(err, "CompareTwoInsightSeriesQuery")
	}

	for i := range points {
		point := points[i]
		resolvers = append(resolvers, &compareTwoInsightSeriesDataPointResolver{point})
	}

	return resolvers, nil
}

func compareTwoSeriesQuery(ctx context.Context, input graphqlbackend.CompareTwoInsightSeriesInput, base baseInsightResolver) ([]compareTwoInsightSeriesDataPoint, error) {
	store := basestore.NewWithDB(base.insightsDB, sql.TxOptions{})

	q := sqlf.Sprintf(compareSql, input.FirstSeriesId, input.SecondSeriesId)
	return scanComparedataPoint(store.Query(ctx, q))
}

const compareSql = `
SELECT compare.first_val,
       compare.second_val,
       COALESCE(ABS(compare.second_val - compare.first_val), compare.first_val, compare.second_val)::int AS diff,
       name,
       time::date
FROM (SELECT name, first.value::int AS first_val, second.value::int AS second_val, first.time
      FROM (SELECT repo_id, value, repo_name_id, time
            FROM series_points
            WHERE series_id = %s) AS first
               FULL JOIN (SELECT repo_id, value, repo_name_id, time
                          FROM series_points
                          WHERE series_id = %s) AS second
                         ON first.repo_id = second.repo_id and first.time = second.time
               LEFT JOIN repo_names ON first.repo_name_id = repo_names.id) AS compare
ORDER BY diff DESC
LIMIT 10;
`

func scanComparedataPoint(rows *sql.Rows, queryErr error) (_ []compareTwoInsightSeriesDataPoint, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	results := make([]compareTwoInsightSeriesDataPoint, 0)
	for rows.Next() {
		var temp compareTwoInsightSeriesDataPoint
		if err := rows.Scan(
			&temp.FirstSeriesValue,
			&temp.SecondSeriesValue,
			&temp.Diff,
			&temp.RepoName,
			&temp.Time,
		); err != nil {
			return []compareTwoInsightSeriesDataPoint{}, err
		}
		results = append(results, temp)
	}
	return results, nil
}
