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
