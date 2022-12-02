package types

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type RepoMetadata struct {
	RepoID         api.RepoID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Ignored        bool
	State          string
	FailureMessage *string
	StartedAt      *time.Time
	FinishedAt     *time.Time
	ProcessAfter   *time.Time
	NumResets      int32
	NumFailures    int32
	ExecutionLogs  []workerutil.ExecutionLogEntry
}

func (meta *RepoMetadata) Cursor() int64 { return int64(meta.RepoID) }
func (meta *RepoMetadata) RecordID() int { return int(meta.RepoID) }
