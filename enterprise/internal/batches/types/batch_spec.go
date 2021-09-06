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

	c.Spec, err = batcheslib.ParseBatchSpec([]byte(rawSpec), batcheslib.ParseBatchSpecOptions{
		// Backend always supports all latest features.
		AllowArrayEnvironments: true,
		AllowTransformChanges:  true,
		AllowConditionalExec:   true,
	})

	return c, err
}

type BatchSpec struct {
	ID     int64
	RandID string

	RawSpec string
	Spec    *batcheslib.BatchSpec

	NamespaceUserID int32
	NamespaceOrgID  int32

	UserID int32

	State BatchSpecState

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Clone returns a clone of a BatchSpec.
func (cs *BatchSpec) Clone() *BatchSpec {
	cc := *cs
	return &cc
}

// RecordID is needed to implement the workerutil.Record interface.
func (bs *BatchSpec) RecordID() int { return int(bs.ID) }

// BatchSpecTTL specifies the TTL of BatchSpecs that haven't been applied
// yet. It's set to 1 week.
const BatchSpecTTL = 7 * 24 * time.Hour

// ExpiresAt returns the time when the BatchSpec will be deleted if not
// applied.
func (cs *BatchSpec) ExpiresAt() time.Time {
	return cs.CreatedAt.Add(BatchSpecTTL)
}

// BatchSpecState defines the possible states of a batch spec.
type BatchSpecState string

// BatchSpecState constants.
const (
	BatchSpecStateQueued     BatchSpecState = "QUEUED"
	BatchSpecStateProcessing BatchSpecState = "PROCESSING"
	BatchSpecStateErrored    BatchSpecState = "ERRORED"
	BatchSpecStateFailed     BatchSpecState = "FAILED"
	BatchSpecStateCompleted  BatchSpecState = "COMPLETED"
)

// Valid returns true if the given BatchSpecState is valid.
func (s BatchSpecState) Valid() bool {
	switch s {
	case BatchSpecStateQueued,
		BatchSpecStateProcessing,
		BatchSpecStateErrored,
		BatchSpecStateFailed,
		BatchSpecStateCompleted:
		return true
	default:
		return false
	}
}

// ToDB returns the database representation of the worker state. That's
// needed because we want to use UPPERCASE in the application and GraphQL layer,
// but need to use lowercase in the database to make it work with workerutil.Worker.
func (s BatchSpecState) ToDB() string { return strings.ToLower(string(s)) }
