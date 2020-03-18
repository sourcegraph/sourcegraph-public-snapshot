package extsvc

import (
	"encoding/json"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
)

// ExternalAccount represents a row in the `user_external_accounts` table. See the GraphQL API's
// corresponding fields in "ExternalAccount" for documentation.
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
// ID. See the GraphQL API's corresponding fields in "ExternalAccount" for documentation.
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

// ExternalAccounts contains a list of accounts that belong to the same external service.
// All fields have a same meaning to ExternalAccountSpec. See GraphQL API's corresponding
// fields in "ExternalAccount" for documentation.
type ExternalAccounts struct {
	ServiceType string
	ServiceID   string
	AccountIDs  []string
}

// TracingFields returns tracing fields for the opentracing log.
func (s *ExternalAccounts) TracingFields() []otlog.Field {
	return []otlog.Field{
		otlog.String("ExternalAccounts.ServiceType", s.ServiceType),
		otlog.String("ExternalAccounts.Perm", s.ServiceID),
		otlog.Int("ExternalAccounts.AccountIDs.Count", len(s.AccountIDs)),
	}
}

// ExternalAccountID is a descriptive type for the external identifier of an external account
// on the code host. It can be the string representation of an integer (e.g. GitLab) or a
// GraphQL ID (e.g. GitHub) depends on the code host type.
type ExternalAccountID string

// ExternalRepoID is a descriptive type for the external identifier of an external repository
// on the code host. It can be the string representation of an integer (e.g. GitLab) or a
// GraphQL ID (e.g. GitHub) depends on the code host type.
type ExternalRepoID string
