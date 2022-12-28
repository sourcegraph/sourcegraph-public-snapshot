package types

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
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
	ExecutionLogs   []workerutil.ExecutionLogEntry
	WorkerHostname  string
	Cancel          bool
}

func (j *OutboundWebhookJob) RecordID() int {
	return int(j.ID)
}
