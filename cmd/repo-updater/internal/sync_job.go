package internal

import "database/sql"

// SyncJob represents an external service that needs to be synced
type SyncJob struct {
	ID                int
	State             string
	FailureMessage    sql.NullString
	StartedAt         sql.NullTime
	FinishedAt        sql.NullTime
	ProcessAfter      sql.NullTime
	NumResets         int
	NumFailures       int
	ExternalServiceID int64
	NextSyncAt        sql.NullTime
}

// RecordID implements workerutil.Record and indicates the queued item id
func (s *SyncJob) RecordID() int {
	return s.ID
}
