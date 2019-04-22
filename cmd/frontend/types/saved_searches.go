package types

// SavedSearches represents a saved search
type SavedSearch struct {
	ID          string
	Description string
	Query       string
	Notify      bool
	NotifySlack bool
	OwnerKind   string
	UserID      *int32
	OrgID       *int32
}
