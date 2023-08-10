package types

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

// State returns the user-visible state, collapsing the other state fields into
// one.
func (c *BatchChange) State() BatchChangeState {
	if c.Closed() {
		return BatchChangeStateClosed
	} else if c.IsDraft() {
		return BatchChangeStateDraft
	}
	return BatchChangeStateOpen
}

func (c *BatchChange) URL(ctx context.Context, namespaceName string) (string, error) {
	// To build the absolute URL, we need to know where Sourcegraph is!
	extURL, err := url.Parse(conf.Get().ExternalURL)
	if err != nil {
		return "", errors.Wrap(err, "parsing external Sourcegraph URL")
	}

	// This needs to be kept consistent with resolvers.batchChangeURL().
	// (Refactoring the resolver to use the same function is difficult due to
	// the different querying and caching behaviour in GraphQL resolvers, so we
	// simply replicate the logic here.)
	u := extURL.ResolveReference(&url.URL{Path: namespaceURL(c.NamespaceOrgID, namespaceName) + "/batch-changes/" + c.Name})

	return u.String(), nil
}

// ToGraphQL returns the GraphQL representation of the state.
func (s BatchChangeState) ToGraphQL() string { return strings.ToUpper(string(s)) }

func namespaceURL(orgID int32, namespaceName string) string {
	prefix := "/users/"
	if orgID != 0 {
		prefix = "/organizations/"
	}

	return prefix + namespaceName
}
