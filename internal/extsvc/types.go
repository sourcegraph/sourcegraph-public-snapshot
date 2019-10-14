package extsvc

import (
	"encoding/json"
	"time"
)

// ExternalAccount represents a row in the `user_external_accounts` table. See the GraphQL API's
// corresponding fields for documentation.
type ExternalAccount struct {
	ID                  int32
	UserID              int32
	ExternalAccountSpec // ServiceType, ServiceID, ClientID, AccountID
	ExternalAccountData // AuthData, AccountData
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// ExternalAccountSpec specifies a user external account by its external identifier (i.e., by
// the identifier provided by the account's owner service), instead of by our database's serial
// ID. See the GraphQL API's corresponding fields for documentation.
type ExternalAccountSpec struct {
	ServiceType string
	ServiceID   string
	ClientID    string
	AccountID   string
}

// ExternalAccountData contains data that can be freely updated in the user external account
// after it has been created. See the GraphQL API's corresponding fields for documentation.
type ExternalAccountData struct {
	AuthData    *json.RawMessage
	AccountData *json.RawMessage
}
