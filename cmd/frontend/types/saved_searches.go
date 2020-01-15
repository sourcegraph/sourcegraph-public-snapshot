package types

// SavedSearch represents a saved search
type SavedSearch struct {
	ID              int32 // the globally unique DB ID
	Description     string
	Query           string  // the literal search query to be ran
	Notify          bool    // whether or not to notify the owner(s) of that saved search via email
	NotifySlack     bool    // whether or not to notify the owner(s) of that saved search via Slack
	UserID          *int32  // if non-nil, the owner is that user. UserID/OrgID are mutually exclusive.
	OrgID           *int32  // if non-nil, the owner is that organization. UserID/OrgID are mutually exclusive.
	SlackWebhookURL *string // if non-nil && NotifySlack == true, indicates that that Slack webhook URL should be used instead of the owners default Slack webhook.
}
