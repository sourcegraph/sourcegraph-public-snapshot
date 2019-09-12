package a8n

import "time"

// A Campaign of changesets (i.e. Changesets and Issues) over multiple Repos over time.
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

// A Changeset is a sum type representing either a Changeset or an Issue
// belonging to a Repository and a Campaign.
type Changeset struct {
	ID          int64
	RepoID      int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Metadata    interface{}
	CampaignIDs []int64
	ExternalID  string
}

// Clone returns a clone of a Changeset.
func (t *Changeset) Clone() *Changeset {
	tt := *t
	tt.CampaignIDs = t.CampaignIDs[:len(t.CampaignIDs):len(t.CampaignIDs)]
	return &tt
}
