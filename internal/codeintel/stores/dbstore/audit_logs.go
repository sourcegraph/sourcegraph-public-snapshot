package dbstore

import (
	"context"
	"time"

	"github.com/jackc/pgtype"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type UploadLog struct {
	LogTimestamp      time.Time
	RecordDeletedAt   *time.Time
	UploadID          int
	Commit            string
	Root              string
	RepositoryID      int
	UploadedAt        time.Time
	Indexer           string
	IndexerVersion    *string
	UploadSize        *int
	AssociatedIndexID *int
	TransitionColumns []map[string]*string
	Reason            *string
	Operation         string
}

func scanUploadAuditLog(s dbutil.Scanner) (log UploadLog, _ error) {
	hstores := pgtype.HstoreArray{}
	err := s.Scan(
		&log.LogTimestamp,
		&log.RecordDeletedAt,
		&log.UploadID,
		&log.Commit,
		&log.Root,
		&log.RepositoryID,
		&log.UploadedAt,
		&log.Indexer,
		&log.IndexerVersion,
		&log.UploadSize,
		&log.AssociatedIndexID,
		&hstores,
		&log.Reason,
		&log.Operation,
	)

	for _, hstore := range hstores.Elements {
		m := make(map[string]*string)
		if err := hstore.AssignTo(&m); err != nil {
			return log, err
		}
		log.TransitionColumns = append(log.TransitionColumns, m)
	}

	return log, err
}

var scanUploadAuditLogs = basestore.NewSliceScanner(scanUploadAuditLog)

// GetAuditLogsForUpload returns all the audit logs for the given upload ID in order of entry
// from oldest to newest, according to the auto-incremented internal sequence field.
func (s *Store) GetAuditLogsForUpload(ctx context.Context, uploadID int) (_ []UploadLog, err error) {
	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.Store))
	if err != nil {
		return nil, err
	}

	return scanUploadAuditLogs(s.Store.Query(ctx, sqlf.Sprintf(getAuditLogsForUploadQuery, uploadID, authzConds)))
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
