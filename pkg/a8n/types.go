package a8n

import "time"

// A Campaign of threads (i.e. ChangeSets and Issues) over multiple Repos over time.
type Campaign struct {
	ID              int64
	Name            string
	Description     string
	AuthorID        int32
	NamespaceUserID int32
	NamespaceOrgID  int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// A Thread is a sum type representing either a ChangeSet or an Issue
// belonging to a Repository and a Campaign.
type Thread struct {
	ID         int64
	CampaignID int64
	RepoID     int32
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Metadata   interface{}
}
