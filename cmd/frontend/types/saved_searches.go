package types

// SavedSearch represents a saved search
type SavedSearch struct {
	ID              int32 // the globally unique DB ID
	Description     string
	Query           string // the literal search query to be ran
	Notify          bool // whether or not to notify the owner(s) of this saved search via email
	NotifySlack     bool // whether or not to notify the owner(s) of this saved search via Slack
	OwnerKind       string
	UserID          *int32 // if non-nil, the owner is this user. UserID/OrgID are mutually exclusive.
	OrgID           *int32 // if non-nil, the owner is this organization. UserID/OrgID are mutually exclusive.
	SlackWebhookURL *string
}
