package database

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *codeMonitorStore) HasAnyLastSearched(ctx context.Context, monitorID int64) (bool, error) {
	rawQuery := `
	SELECT COUNT(*) > 0
	FROM cm_last_searched
	WHERE monitor_id = %s
	`

	q := sqlf.Sprintf(rawQuery, monitorID)
	var hasLastSearched bool
	return hasLastSearched, s.QueryRow(ctx, q).Scan(&hasLastSearched)
}

func (s *codeMonitorStore) UpsertLastSearched(ctx context.Context, monitorID int64, repoID api.RepoID, commitOIDs []string) error {
	rawQuery := `
	INSERT INTO cm_last_searched (monitor_id, repo_id, commit_oids)
	VALUES (%s, %s, %s)
	ON CONFLICT (monitor_id, repo_id) DO UPDATE
	SET commit_oids = %s
	`

	// Appease non-null constraint on column
	if commitOIDs == nil {
		commitOIDs = []string{}
	}
	q := sqlf.Sprintf(rawQuery, monitorID, int64(repoID), pq.StringArray(commitOIDs), pq.StringArray(commitOIDs))
	return s.Exec(ctx, q)
}

func (s *codeMonitorStore) GetLastSearched(ctx context.Context, monitorID int64, repoID api.RepoID) ([]string, error) {
	rawQuery := `
	SELECT commit_oids
	FROM cm_last_searched
	WHERE monitor_id = %s
		AND repo_id = %s
	LIMIT 1
	`

	q := sqlf.Sprintf(rawQuery, monitorID, int64(repoID))
	var commitOIDs []string
	err := s.QueryRow(ctx, q).Scan((*pq.StringArray)(&commitOIDs))
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return commitOIDs, err
}
