package types

// SavedSearch represents a saved search
type SavedSearch struct {
	ID              int32
	Description     string
	Query           string
	Notify          bool
	NotifySlack     bool
	UserID          *int32
	OrgID           *int32
	SlackWebhookURL *string
}
