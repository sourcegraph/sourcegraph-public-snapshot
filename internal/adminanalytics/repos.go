package adminanalytics

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type Repos struct {
	DB database.DB
}

func (r *Repos) Summary(ctx context.Context, cache bool) (*ReposSummary, error) {
	cacheKey := "Repos:Summary"
	if cache == true {
		if summary, err := getItemFromCache[ReposSummary](cacheKey); err == nil {
			return summary, nil
		}
	}

	query := sqlf.Sprintf(`
	SELECT
		COUNT(DISTINCT repo.id) as total_repo_count,
		COUNT(DISTINCT lsif_uploads.repository_id) as lsif_index_repo_count
	FROM
		repo
		LEFT JOIN lsif_uploads ON lsif_uploads.repository_id = repo.id
	`)
	var data ReposSummaryData

	if err := r.DB.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...).Scan(&data.TotalCount, &data.PreciseCodeIntelCount); err != nil {
		return nil, err
	}

	summary := &ReposSummary{data}


	if _, err := setItemToCache(cacheKey, summary); err != nil {
		return nil, err
	}

	return summary, nil
}

type ReposSummary struct {
	Data ReposSummaryData
}

type ReposSummaryData struct {
	TotalCount                 int32
	PreciseCodeIntelCount int32
}

func (s *ReposSummary) TotalCount() int32 { return s.Data.TotalCount }

func (s *ReposSummary) PreciseCodeIntelCount() int32 { return s.Data.PreciseCodeIntelCount }
