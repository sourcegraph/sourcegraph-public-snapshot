package bitbucketcloud

import "time"

// General types we need to be able to handle, but which don't have specific
// endpoints we need to implement methods for.

type Account struct {
	Links         Links         `json:"links"`
	Username      string        `json:"username"`
	Nickname      string        `json:"nickname"`
	AccountStatus AccountStatus `json:"account_status"`
	DisplayName   string        `json:"display_name"`
	Website       string        `json:"website"`
	CreatedOn     time.Time     `json:"created_on"`
	UUID          string        `json:"uuid"`
}

type Link struct {
	Href string `json:"href"`
	Name string `json:"name,omitempty"`
}

type Links map[string]Link

type Participant struct {
	User           User             `json:"user"`
	Role           ParticipantRole  `json:"role"`
	Approved       bool             `json:"approved"`
	State          ParticipantState `json:"state"`
	ParticipatedOn time.Time        `json:"participated_on"`
}

type RenderedMarkup struct {
	Raw    string `json:"raw"`
	Markup string `json:"markup"`
	HTML   string `json:"html"`
}

type AccountStatus string
type ParticipantRole string
type ParticipantState string

const (
	AccountStatusActive AccountStatus = "active"

	ParticipantRoleParticipant ParticipantRole = "PARTICIPANT"
	ParticipantRoleReviewer    ParticipantRole = "REVIEWER"

	ParticipantStateApproved         ParticipantState = "approved"
	ParticipantStateChangesRequested ParticipantState = "changes_requested"
	ParticipantStateNull             ParticipantState = "null"
)
