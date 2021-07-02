package types

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type BatchSpecExecution struct {
	ID             int64                          `json:"id"`
	State          string                         `json:"state"`
	FailureMessage *string                        `json:"failureMessage"`
	StartedAt      *time.Time                     `json:"startedAt"`
	FinishedAt     *time.Time                     `json:"finishedAt"`
	ProcessAfter   *time.Time                     `json:"processAfter"`
	NumResets      int64                          `json:"numResets"`
	NumFailures    int64                          `json:"numFailures"`
	ExecutionLogs  []workerutil.ExecutionLogEntry `json:"execution_logs"`
	WorkerHostname string                         `json:"worker_hostname"`
	CreatedAt      time.Time                      `json:"created_at"`
	UpdatedAt      time.Time                      `json:"updated_at"`
	BatchSpec      string                         `json:"batch_spec"`
	BatchSpecID    int64                          `json:"batch_spec_id"`
}

func (i BatchSpecExecution) RecordID() int {
	return int(i.ID)
}
