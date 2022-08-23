package adminanalytics

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type CodeIntelTopRepositories struct {
	Name_   string  `json:"name"`
	Events_ float64 `json:"events"`
}

func (s *CodeIntelTopRepositories) Name() string    { return s.Name_ }
func (s *CodeIntelTopRepositories) Events() float64 { return s.Events_ }

func GetCodeIntelTopRepositories(ctx context.Context, db database.DB, cache bool, dateRange string) ([]*CodeIntelTopRepositories, error) {
	cacheKey := fmt.Sprintf(`CodeIntelTopRepositories:%s`, dateRange)

	if cache == true {
		if nodes, err := getArrayFromCache[CodeIntelTopRepositories](cacheKey); err == nil {
			return nodes, nil
		}
	}

	now := time.Now()
	from, err := getFromDate(dateRange, now)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, `
		SELECT name, events
		FROM (
			SELECT (argument->>'repositoryId')::int as id, COUNT(*) AS events
			FROM event_logs
			WHERE
				timestamp BETWEEN $1 AND $2 AND
				name IN (
					'codeintel.searchDefinitions',
					'codeintel.searchDefinitions.xrepo',
					'codeintel.searchReferences',
					'codeintel.searchReferences.xrepo',
					'codeintel.lsifDefinitions',
					'codeintel.lsifDefinitions.xrepo',
					'codeintel.lsifReferences',
					'codeintel.lsifReferences.xrepo'
				)
			GROUP BY argument->>'repositoryId'
			LIMIT 5
		) sub
		JOIN repo ON repo.id = sub.id;
	`, from.Format(time.RFC3339), now.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*CodeIntelTopRepositories{}
	for rows.Next() {
		var item CodeIntelTopRepositories

		if err := rows.Scan(&item.Name_, &item.Events_); err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	if _, err := setArrayToCache(cacheKey, items); err != nil {
		return nil, err
	}

	return items, nil
}
