package types

import "time"

// SavedSearch represents a saved search.
type SavedSearch struct {
	ID          int32 // the globally unique DB ID
	Description string
	Query       string    // the search query
	Owner       Namespace // the owner
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
