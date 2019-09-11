package a8n

import "time"

// A Campaign of changesets (i.e. ChangeSets and Issues) over multiple Repos over time.
type Campaign struct {
	ID              int64
	Name            string
	Description     string
	AuthorID        int32
	NamespaceUserID int32
	NamespaceOrgID  int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ChangeSetIDs    []int64
}

// Clone returns a clone of a Campaign.
func (c *Campaign) Clone() *Campaign {
	cc := *c
	cc.ChangeSetIDs = c.ChangeSetIDs[:len(c.ChangeSetIDs):len(c.ChangeSetIDs)]
	return &cc
}

// A ChangeSet is a sum type representing either a ChangeSet or an Issue
// belonging to a Repository and a Campaign.
type ChangeSet struct {
	ID          int64
	RepoID      int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Metadata    interface{}
	CampaignIDs []int64
	ExternalID  string
}

// Clone returns a clone of a ChangeSet.
func (t *ChangeSet) Clone() *ChangeSet {
	tt := *t
	tt.CampaignIDs = t.CampaignIDs[:len(t.CampaignIDs):len(t.CampaignIDs)]
	return &tt
}
