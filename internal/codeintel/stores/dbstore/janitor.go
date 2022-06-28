package dbstore

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type SourcedCommits struct {
	RepositoryID   int
	RepositoryName string
	Commits        []string
}

// ScanSourcedCommits scans triples of repository ids/repository names/commits from the
// return value of `*Store.query`. The output of this function is ordered by repository
// identifier, then by commit.
func ScanSourcedCommits(rows *sql.Rows, queryErr error) (_ []SourcedCommits, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	sourcedCommitsMap := map[int]SourcedCommits{}
	for rows.Next() {
		var repositoryID int
		var repositoryName string
		var commit string
		if err := rows.Scan(&repositoryID, &repositoryName, &commit); err != nil {
			return nil, err
		}

		sourcedCommitsMap[repositoryID] = SourcedCommits{
			RepositoryID:   repositoryID,
			RepositoryName: repositoryName,
			Commits:        append(sourcedCommitsMap[repositoryID].Commits, commit),
		}
	}

	flattened := make([]SourcedCommits, 0, len(sourcedCommitsMap))
	for _, sourcedCommits := range sourcedCommitsMap {
		sort.Strings(sourcedCommits.Commits)
		flattened = append(flattened, sourcedCommits)
	}

	sort.Slice(flattened, func(i, j int) bool {
		return flattened[i].RepositoryID < flattened[j].RepositoryID
	})
	return flattened, nil
}

// DeleteOldAuditLogs removes lsif_upload audit log records older than the given max age.
func (s *Store) DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (_ int, err error) {
	ctx, _, endObservation := s.operations.deleteOldAuditLogs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	count, _, err := basestore.ScanFirstInt(s.Store.Query(ctx, sqlf.Sprintf(deleteOldAuditLogsQuery, now, int(maxAge/time.Second))))
	return count, err
}

const deleteOldAuditLogsQuery = `
-- source: internal/codeintel/stores/dbstore/janitor.go:DeleteOldAuditLogs
WITH deleted AS (
	DELETE FROM lsif_uploads_audit_logs
	WHERE %s - log_timestamp > (%s * '1 second'::interval)
	RETURNING upload_id
)
SELECT count(*) FROM deleted
`
