package types

import (
	"strings"
	"time"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

// NewBatchSpecFromRaw parses and validates the given rawSpec, and returns a BatchSpec
// containing the result.
func NewBatchSpecFromRaw(rawSpec string) (_ *BatchSpec, err error) {
	c := &BatchSpec{RawSpec: rawSpec}

	c.Spec, err = batcheslib.ParseBatchSpec([]byte(rawSpec))

	return c, err
}

type BatchSpec struct {
	ID     int64
	RandID string

	RawSpec string
	Spec    *batcheslib.BatchSpec

	NamespaceUserID int32
	NamespaceOrgID  int32

	UserID        int32
	BatchChangeID int64

	// CreatedFromRaw is true when the BatchSpec was created through the
	// createBatchSpecFromRaw GraphQL mutation, which means that it's meant to be
	// executed server-side.
	CreatedFromRaw bool

	AllowUnsupported bool
	AllowIgnored     bool
	NoCache          bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Clone returns a clone of a BatchSpec.
func (cs *BatchSpec) Clone() *BatchSpec {
	cc := *cs
	return &cc
}

// BatchSpecTTL specifies the TTL of BatchSpecs that haven't been applied
// yet. It's set to 1 week.
const BatchSpecTTL = 7 * 24 * time.Hour

// ExpiresAt returns the time when the BatchSpec will be deleted if not
// applied.
func (cs *BatchSpec) ExpiresAt() time.Time {
	return cs.CreatedAt.Add(BatchSpecTTL)
}

type BatchSpecStats struct {
	ResolutionDone bool

	Workspaces        int
	SkippedWorkspaces int
	CachedWorkspaces  int
	Executions        int

	Queued     int
	Processing int
	Completed  int
	Canceling  int
	Canceled   int
	Failed     int

	StartedAt  time.Time
	FinishedAt time.Time
}

// BatchSpecState defines the possible states of a BatchSpec that was created
// to be executed server-side. Client-side batch specs (created with src-cli)
// are always in state "completed".
//
// Some variants of this state are only computed in the BatchSpecResolver.
type BatchSpecState string

const (
	BatchSpecStatePending    BatchSpecState = "pending"
	BatchSpecStateQueued     BatchSpecState = "queued"
	BatchSpecStateProcessing BatchSpecState = "processing"
	BatchSpecStateErrored    BatchSpecState = "errored"
	BatchSpecStateFailed     BatchSpecState = "failed"
	BatchSpecStateCompleted  BatchSpecState = "completed"
	BatchSpecStateCanceled   BatchSpecState = "canceled"
	BatchSpecStateCanceling  BatchSpecState = "canceling"
)

// ToGraphQL returns the GraphQL representation of the state.
func (s BatchSpecState) ToGraphQL() string { return strings.ToUpper(string(s)) }

// Cancelable returns whether the state is one in which the BatchSpec can be
// canceled.
func (s BatchSpecState) Cancelable() bool {
	return s == BatchSpecStateQueued || s == BatchSpecStateProcessing
}

// Started returns whether the execution of the BatchSpec has started.
func (s BatchSpecState) Started() bool {
	return s != BatchSpecStateQueued && s != BatchSpecStatePending
}

// Finished returns whether the execution of the BatchSpec has finished.
func (s BatchSpecState) Finished() bool {
	return s == BatchSpecStateCompleted ||
		s == BatchSpecStateFailed ||
		s == BatchSpecStateErrored ||
		s == BatchSpecStateCanceled
}

// FinishedAndNotCanceled returns whether the execution of the BatchSpec ran
// through and finished without being canceled.
func (s BatchSpecState) FinishedAndNotCanceled() bool {
	return s == BatchSpecStateCompleted || s == BatchSpecStateFailed
}

// ComputeBatchSpecState computes the BatchSpecState based on the given stats.
func ComputeBatchSpecState(spec *BatchSpec, stats BatchSpecStats) BatchSpecState {
	if !spec.CreatedFromRaw {
		return BatchSpecStateCompleted
	}

	if !stats.ResolutionDone {
		return BatchSpecStatePending
	}

	if stats.Workspaces == 0 {
		return BatchSpecStateCompleted
	}

	if stats.SkippedWorkspaces == stats.Workspaces {
		return BatchSpecStateCompleted
	}

	if stats.Executions == 0 {
		return BatchSpecStatePending
	}

	if stats.Queued == stats.Executions {
		return BatchSpecStateQueued
	}

	if stats.Completed == stats.Executions {
		return BatchSpecStateCompleted
	}

	if stats.Canceled == stats.Executions {
		return BatchSpecStateCanceled
	}

	if stats.Failed+stats.Completed == stats.Executions {
		return BatchSpecStateFailed
	}

	if stats.Canceled+stats.Failed+stats.Completed == stats.Executions {
		return BatchSpecStateCanceled
	}

	if stats.Canceling+stats.Failed+stats.Completed+stats.Canceled == stats.Executions {
		return BatchSpecStateCanceling
	}

	if stats.Canceling > 0 || stats.Processing > 0 {
		return BatchSpecStateProcessing
	}

	if (stats.Completed > 0 || stats.Failed > 0 || stats.Canceled > 0) && stats.Queued > 0 {
		return BatchSpecStateProcessing
	}

	return "INVALID"
}

// BatchSpecSource defines the possible sources for creating a BatchSpec. Client-side
// batch specs (created with src-cli) are said to have the "local" source, and batch specs
// created for server-side execution are said to have the "remote" source.
type BatchSpecSource string

const (
	BatchSpecSourceLocal  BatchSpecState = "local"
	BatchSpecSourceRemote BatchSpecState = "remote"
)

func (s BatchSpecSource) ToGraphQL() string { return strings.ToUpper(string(s)) }
