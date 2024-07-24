package types

import "time"

// SavedSearch represents a saved search.
type SavedSearch struct {
	ID               int32 // the globally unique DB ID
	Description      string
	Query            string    // the search query
	Draft            bool      // whether the prompt is a draft
	Owner            Namespace // the owner
	VisibilitySecret bool      // the visibility state (if false, public)

	CreatedAt     time.Time // when this saved search was created
	CreatedByUser *int32    // the user that created this saved search
	UpdatedAt     time.Time // when this saved search was last updated
	UpdatedByUser *int32    // the user that last updated this saved search
}
