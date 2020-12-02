package campaigns

import "time"

// CampaignState defines the possible states of a Campaign
type CampaignState string

const (
	CampaignStateAny    CampaignState = "ANY"
	CampaignStateOpen   CampaignState = "OPEN"
	CampaignStateClosed CampaignState = "CLOSED"
)

// A Campaign of changesets over multiple Repos over time.
type Campaign struct {
	ID          int64
	Name        string
	Description string

	CampaignSpecID int64

	InitialApplierID int32
	LastApplierID    int32
	LastAppliedAt    time.Time

	NamespaceUserID int32
	NamespaceOrgID  int32

	ClosedAt time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Clone returns a clone of a Campaign.
func (c *Campaign) Clone() *Campaign {
	cc := *c
	return &cc
}

// Closed returns true when the ClosedAt timestamp has been set.
func (c *Campaign) Closed() bool { return !c.ClosedAt.IsZero() }
