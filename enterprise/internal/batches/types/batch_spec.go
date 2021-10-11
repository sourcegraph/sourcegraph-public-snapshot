package types

import (
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

	// CreatedFromRaw is true when the BatchSpec was created through the
	// createBatchSpecFromRaw GraphQL mutation, which means that it's meant to be
	// executed server-side.
	CreatedFromRaw bool

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
