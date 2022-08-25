package adminanalytics

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CodeIntelTopRepositories struct {
	LangId_     string  `json:"langId"`
	Name_       string  `json:"name"`
	Events_     float64 `json:"events"`
	Kind_       string  `json:"kind"`
	Precision_  string  `json:"precision"`
	HasPrecise_ bool    `json:"hasPrecise"`
}

func (s *CodeIntelTopRepositories) Name() string      { return s.Name_ }
func (s *CodeIntelTopRepositories) Language() string  { return s.LangId_ }
func (s *CodeIntelTopRepositories) Events() float64   { return s.Events_ }
func (s *CodeIntelTopRepositories) Kind() string      { return s.Kind_ }
func (s *CodeIntelTopRepositories) Precision() string { return s.Precision_ }
func (s *CodeIntelTopRepositories) HasPrecise() bool  { return s.HasPrecise_ }

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
		WITH events AS (
			SELECT *
			FROM (
				SELECT argument, (
					CASE
					WHEN name = 'codeintel.lsifDefinitions.xrepo'                                      THEN 'crossRepo'
					WHEN name = 'codeintel.lsifDefinitions'                                            THEN 'precise'
					WHEN name = 'codeintel.lsifReferences.xrepo'                                       THEN 'crossRepo'
					WHEN name = 'codeintel.lsifReferences'                                             THEN 'precise'
					WHEN name = 'codeintel.searchDefinitions.xrepo'                                    THEN 'crossRepo'
					WHEN name = 'codeintel.searchReferences.xrepo'                                     THEN 'crossRepo'
					WHEN name = 'findReferences'                    AND source = 'CODEHOSTINTEGRATION' THEN 'codeHost'
					WHEN name = 'findReferences'                    AND source = 'WEB'                 THEN 'inApp'
					WHEN name = 'goToDefinition.preloaded'          AND source = 'CODEHOSTINTEGRATION' THEN 'codeHost'
					WHEN name = 'goToDefinition.preloaded'          AND source = 'WEB'                 THEN 'inApp'
					WHEN name = 'goToDefinition'                    AND source = 'CODEHOSTINTEGRATION' THEN 'codeHost'
					WHEN name = 'goToDefinition'                    AND source = 'WEB'                 THEN 'inApp'
					WHEN name = 'codeintel.searchDefinitions'                                          THEN 'inApp'
					ELSE NULL
					END
				) AS event_kind, (
					CASE
					WHEN name = 'codeintel.lsifDefinitions.xrepo' THEN 'precise'
					WHEN name = 'codeintel.lsifDefinitions'       THEN 'precise'
					WHEN name = 'codeintel.lsifHover'             THEN 'precise'
					WHEN name = 'codeintel.lsifReferences.xrepo'  THEN 'precise'
					WHEN name = 'codeintel.lsifReferences'        THEN 'precise'
					ELSE                                               'search-based'
					END
				) AS event_precision
				FROM event_logs
				WHERE timestamp BETWEEN $1 AND $2
			) AS _
			WHERE event_kind IS NOT NULL
		), top_repos AS (
			SELECT
				repo.id AS repo_id,
				repo.name AS repo_name,
				EXISTS (SELECT 1 FROM lsif_uploads WHERE repository_id = repo.id AND state = 'completed') AS has_precise
			FROM events
			JOIN repo ON repo.id = (argument->>'repositoryId')::int
			GROUP BY repo.id, repo.name
			ORDER BY COUNT(*) DESC
			LIMIT 5
		)
		SELECT repo_name, lang, event_kind, event_precision, event_count, has_precise
		FROM
			top_repos,
			LATERAL (
				SELECT (argument->>'languageId')::text AS lang, event_kind, event_precision, COUNT(1) AS event_count
				FROM events
				WHERE (argument->>'repositoryId')::int = repo_id
				GROUP BY (argument->>'languageId')::text, event_kind, event_precision
			) AS langs;
	`, from.Format(time.RFC3339), now.Format(time.RFC3339))
	if err != nil {
		return nil, errors.Wrap(err, "GetCodeIntelTopRepositories SQL query")
	}
	defer rows.Close()

	items := []*CodeIntelTopRepositories{}
	for rows.Next() {
		var item CodeIntelTopRepositories

		if err := rows.Scan(&item.Name_, &item.LangId_, &item.Kind_, &item.Precision_, &item.Events_, &item.HasPrecise_); err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	if _, err := setArrayToCache(cacheKey, items); err != nil {
		return nil, err
	}

	return items, nil
}
