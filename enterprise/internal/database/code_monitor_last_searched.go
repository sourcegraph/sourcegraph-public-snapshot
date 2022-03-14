package database

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *codeMonitorStore) UpsertLastSearched(ctx context.Context, monitorID, argsHash int64, commitOIDs []string) error {
	rawQuery := `
	INSERT INTO cm_last_searched (monitor_id, args_hash, commit_oids)
	VALUES (%s, %s, %s)
	ON CONFLICT (monitor_id, args_hash) DO UPDATE
	SET commit_oids = %s
	`

	// Appease non-null constraint on column
	if commitOIDs == nil {
		commitOIDs = []string{}
	}
	q := sqlf.Sprintf(rawQuery, monitorID, argsHash, pq.StringArray(commitOIDs), pq.StringArray(commitOIDs))
	return s.Exec(ctx, q)
}

func (s *codeMonitorStore) GetLastSearched(ctx context.Context, monitorID, argsHash int64) ([]string, error) {
	rawQuery := `
	SELECT commit_oids
	FROM cm_last_searched
	WHERE monitor_id = %s
		AND args_hash = %s
	LIMIT 1
	`

	q := sqlf.Sprintf(rawQuery, monitorID, argsHash)
	var commitOIDs []string
	err := s.QueryRow(ctx, q).Scan((*pq.StringArray)(&commitOIDs))
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return commitOIDs, err
}
