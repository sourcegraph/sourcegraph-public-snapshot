package types

import (
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/executor"
)

type OutboundWebhookJob struct {
	ID int64

	EventType string
	Scope     *string
	Payload   *encryption.Encryptable

	State           string
	FailureMessage  *string
	QueuedAt        time.Time
	StartedAt       *time.Time
	FinishedAt      *time.Time
	ProcessAfter    *time.Time
	NumResets       int
	NumFailures     int
	LastHeartbeatAt time.Time
	ExecutionLogs   []executor.ExecutionLogEntry
	WorkerHostname  string
	Cancel          bool
}

func (j *OutboundWebhookJob) RecordID() int {
	return int(j.ID)
}

func (j *OutboundWebhookJob) RecordUID() string {
	return strconv.FormatInt(j.ID, 10)
}
