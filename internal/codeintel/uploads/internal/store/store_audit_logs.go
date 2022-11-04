package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetAuditLogsForUpload returns all the audit logs for the given upload ID in order of entry
// from oldest to newest, according to the auto-incremented internal sequence field.
func (s *store) GetAuditLogsForUpload(ctx context.Context, uploadID int) (_ []shared.UploadLog, err error) {
	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.db))
	if err != nil {
		return nil, err
	}

	return scanUploadAuditLogs(s.db.Query(ctx, sqlf.Sprintf(getAuditLogsForUploadQuery, uploadID, authzConds)))
}

const getAuditLogsForUploadQuery = `
-- source: internal/codeintel/stores/dbstore/audit_logs.go:GetAuditLogsForUpload
SELECT
	u.log_timestamp,
	u.record_deleted_at,
	u.upload_id,
	u.commit,
	u.root,
	u.repository_id,
	u.uploaded_at,
	u.indexer,
	u.indexer_version,
	u.upload_size,
	u.associated_index_id,
	u.transition_columns,
	u.reason,
	u.operation
FROM lsif_uploads_audit_logs u
JOIN repo ON repo.id = u.repository_id
WHERE u.upload_id = %s AND %s
ORDER BY u.sequence
`

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
