package types

import (
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type BatchSpecExecutionState string

const (
	BatchSpecExecutionStateQueued     BatchSpecExecutionState = "queued"
	BatchSpecExecutionStateErrored    BatchSpecExecutionState = "errored"
	BatchSpecExecutionStateFailed     BatchSpecExecutionState = "failed"
	BatchSpecExecutionStateCompleted  BatchSpecExecutionState = "completed"
	BatchSpecExecutionStateProcessing BatchSpecExecutionState = "processing"
)

type BatchSpecExecution struct {
	ID              int64
	RandID          string
	State           BatchSpecExecutionState
	FailureMessage  *string
	StartedAt       *time.Time
	FinishedAt      *time.Time
	ProcessAfter    *time.Time
	NumResets       int64
	NumFailures     int64
	ExecutionLogs   []workerutil.ExecutionLogEntry
	WorkerHostname  string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	BatchSpec       string
	BatchSpecID     int64
	UserID          int32
	NamespaceUserID int32
	NamespaceOrgID  int32
	Cancel          bool
}

func (e BatchSpecExecution) RecordID() int {
	return int(e.ID)
}

func (e BatchSpecExecution) GQLState() string {
	if e.Cancel {
		if e.State == BatchSpecExecutionStateFailed {
			return "CANCELED"
		}
		return "CANCELING"
	}
	return strings.ToUpper(string(e.State))
}
