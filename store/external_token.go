package store

import (
	"errors"
	"time"
)

// ExternalAuthToken is an OAuth(2) access token for an external site, such as GitHub.
type ExternalAuthToken struct {
	User        int       // the user whose resources this token enables access to
	Host        string    // the host that this token enables access to (e.g., "github.com")
	Token       string    // the OAuth2 access token
	Scope       string    // the scope listing the permissions the token is entitled to
	ExtUID      int       `db:"ext_uid"`      // the external user id corresponding to this token
	RefreshedAt time.Time `db:"refreshed_at"` // the last time this token was refreshed

	// ClientID is the application client ID this token was granted
	// to. The same client ID must be used when using this access
	// token.
	ClientID string `db:"client_id"`

	Disabled bool // whether this token is disabled (it can be disabled if auth failures occur)

	AuthFailureCount        int        `db:"auth_failure_count"`                      // the number of auth failures encountered since FirstAuthFailureAt
	FirstAuthFailureAt      *time.Time `db:"first_auth_failure_at" json:",omitempty"` // the date of the first auth failure encountered
	FirstAuthFailureMessage string     `db:"first_auth_failure_message" json:"-"`     // error message accompanying the first auth failure
}

var (
	// ErrNoExternalAuthToken occurs when no external auth token exists
	// for a given user and host.
	ErrNoExternalAuthToken = errors.New("no external auth token found for user and host")

	// ErrExternalAuthTokenDisabled occurs when an external auth token
	// has been disabled, likely due to auth failures.
	ErrExternalAuthTokenDisabled = errors.New("external auth token is disabled")
)
