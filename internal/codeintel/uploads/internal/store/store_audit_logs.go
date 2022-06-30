package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// DeleteOldAuditLogs removes lsif_upload audit log records older than the given max age.
func (s *store) DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (count int, err error) {
	ctx, _, endObservation := s.operations.deleteOldAuditLogs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	query := sqlf.Sprintf(deleteOldAuditLogsQuery, now, int(maxAge/time.Second))
	count, _, err = basestore.ScanFirstInt(s.db.Query(ctx, query))
	return count, err
}

const deleteOldAuditLogsQuery = `
-- source: internal/codeintel/uploads/internal/store/store_audit_logs.go:DeleteOldAuditLogs
WITH deleted AS (
	DELETE FROM lsif_uploads_audit_logs
	WHERE %s - log_timestamp > (%s * '1 second'::interval)
	RETURNING upload_id
)
SELECT count(*) FROM deleted
`
