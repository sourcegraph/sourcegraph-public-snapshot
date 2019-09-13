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

func (t *Changeset) Title() (string, error) {
	switch m := t.Metadata.(type) {
	case *github.PullRequest:
		return m.Title, nil
	default:
		return "", errors.New("unknown changeset type")
	}
}

func (t *Changeset) Body() (string, error) {
	switch m := t.Metadata.(type) {
	case *github.PullRequest:
		return m.Body, nil
	default:
		return "", errors.New("unknown changeset type")
	}
}
