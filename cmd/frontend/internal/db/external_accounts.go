package db

// ExternalAccountSpec specifies a user external account by its external identifier (i.e., by
// the identifier provided by the account's owner service), instead of by our database's serial
// ID. See the GraphQL API's corresponding fields for documentation.
type ExternalAccountSpec struct {
	ServiceType string
	ServiceID   string
	AccountID   string
}
