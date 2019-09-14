package a8n

import (
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
)

// A Campaign of changesets over multiple Repos over time.
type Campaign struct {
	ID              int64
	Name            string
	Description     string
	AuthorID        int32
	NamespaceUserID int32
	NamespaceOrgID  int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ChangesetIDs    []int64
}

// Clone returns a clone of a Campaign.
func (c *Campaign) Clone() *Campaign {
	cc := *c
	cc.ChangesetIDs = c.ChangesetIDs[:len(c.ChangesetIDs):len(c.ChangesetIDs)]
	return &cc
}

// ChangesetState defines the possible states of a Changeset.
type ChangesetState string

// ChangesetState constants.
const (
	ChangesetStateOpen   ChangesetState = "OPEN"
	ChangesetStateClosed ChangesetState = "CLOSED"
	ChangesetStateMerged ChangesetState = "MERGED"
)

// Valid returns true if the given Changeset is valid.
func (s ChangesetState) Valid() bool {
	switch s {
	case ChangesetStateOpen,
		ChangesetStateClosed,
		ChangesetStateMerged:
		return true
	default:
		return false
	}
}

// A Changeset is a changeset on a code host belonging to a Repository and many
// Campaigns.
type Changeset struct {
	ID                  int64
	RepoID              int32
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Metadata            interface{}
	CampaignIDs         []int64
	ExternalID          string
	ExternalServiceType string
}

// Clone returns a clone of a Changeset.
func (t *Changeset) Clone() *Changeset {
	tt := *t
	tt.CampaignIDs = t.CampaignIDs[:len(t.CampaignIDs):len(t.CampaignIDs)]
	return &tt
}

// Title of the Changeset.
func (t *Changeset) Title() (string, error) {
	switch m := t.Metadata.(type) {
	case *github.PullRequest:
		return m.Title, nil
	default:
		return "", errors.New("unknown changeset type")
	}
}

// Body of the Changeset.
func (t *Changeset) Body() (string, error) {
	switch m := t.Metadata.(type) {
	case *github.PullRequest:
		return m.Body, nil
	default:
		return "", errors.New("unknown changeset type")
	}
}

// State of a Changeset.
func (t *Changeset) State() (s ChangesetState, err error) {
	switch m := t.Metadata.(type) {
	case *github.PullRequest:
		s = ChangesetState(m.State)
	default:
		return "", errors.New("unknown changeset type")
	}

	if !s.Valid() {
		return "", errors.Errorf("changeset state %q invalid", s)
	}

	return s, nil
}

// URL of a Changeset.
func (t *Changeset) URL() (s string, err error) {
	switch m := t.Metadata.(type) {
	case *github.PullRequest:
		return m.URL, nil
	default:
		return "", errors.New("unknown changeset type")
	}
}
