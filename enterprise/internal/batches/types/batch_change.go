package types

import (
	"strings"
	"time"
)

// BatchChangeState defines the possible states of a BatchChange
type BatchChangeState string

const (
	BatchChangeStateOpen   BatchChangeState = "OPEN"
	BatchChangeStateClosed BatchChangeState = "CLOSED"
	BatchChangeStateDraft  BatchChangeState = "DRAFT"
)

// A BatchChange of changesets over multiple Repos over time.
type BatchChange struct {
	ID          int64
	Name        string
	Description string

	BatchSpecID int64

	CreatorID     int32
	LastApplierID int32
	LastAppliedAt time.Time

	NamespaceUserID int32
	NamespaceOrgID  int32

	ClosedAt time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Clone returns a clone of a BatchChange.
func (c *BatchChange) Clone() *BatchChange {
	cc := *c
	return &cc
}

// Closed returns true when the ClosedAt timestamp has been set.
func (c *BatchChange) Closed() bool { return !c.ClosedAt.IsZero() }

// IsDraft returns true when the BatchChange is a draft ("shallow") Batch
// Change, i.e. it's associated with a BatchSpec but it hasn't been applied
// yet.
func (c *BatchChange) IsDraft() bool { return c.LastAppliedAt.IsZero() }

// ToGraphQL returns the GraphQL representation of the state.
func (s BatchChangeState) ToGraphQL() string { return strings.ToUpper(string(s)) }
