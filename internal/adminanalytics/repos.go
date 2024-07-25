package adminanalytics

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type Repos struct {
	DB    database.DB
	Cache KeyValue
}

func (r *Repos) Summary(ctx context.Context) (*ReposSummary, error) {
	cacheKey := "Repos:Summary"
	if summary, err := getItemFromCache[ReposSummary](r.Cache, cacheKey); err == nil {
		return summary, nil
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

	if err := r.DB.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...).Scan(&data.Count, &data.PreciseCodeIntelCount); err != nil {
		return nil, err
	}

	summary := &ReposSummary{data}

	err := setItemToCache(r.Cache, cacheKey, summary)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

type ReposSummary struct {
	Data ReposSummaryData
}

type ReposSummaryData struct {
	Count                 float64
	PreciseCodeIntelCount float64
}

func (s *ReposSummary) Count() float64 { return s.Data.Count }

func (s *ReposSummary) PreciseCodeIntelCount() float64 { return s.Data.PreciseCodeIntelCount }

func (s *Repos) CacheAll(ctx context.Context) error {
	if _, err := s.Summary(ctx); err != nil {
		return err
	}

	return nil
}
