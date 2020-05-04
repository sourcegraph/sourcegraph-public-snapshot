package extsvc

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Account represents a row in the `user_external_accounts` table. See the GraphQL API's
// corresponding fields in "ExternalAccount" for documentation.
type Account struct {
	ID          int32
	UserID      int32
	AccountSpec // ServiceType, ServiceID, ClientID, AccountID
	AccountData // AuthData, Data
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AccountSpec specifies a user external account by its external identifier (i.e., by the
// identifier provided by the account's owner service), instead of by our database's serial
// ID. See the GraphQL API's corresponding fields in "ExternalAccount" for documentation.
type AccountSpec struct {
	ServiceType string
	ServiceID   string
	ClientID    string
	AccountID   string
}

// AccountData contains data that can be freely updated in the user external account after it
// has been created. See the GraphQL API's corresponding fields for documentation.
type AccountData struct {
	AuthData *json.RawMessage
	Data     *json.RawMessage
}

// Repository contains necessary information to identify an external repository on the code host.
type Repository struct {
	// URI is the full name for this repository, e.g. "github.com/user/repo".
	URI string
	api.ExternalRepoSpec
}

// Accounts contains a list of accounts that belong to the same external service.
// All fields have a same meaning to AccountSpec. See GraphQL API's corresponding fields
// in "ExternalAccount" for documentation.
type Accounts struct {
	ServiceType string
	ServiceID   string
	AccountIDs  []string
}

// TracingFields returns tracing fields for the opentracing log.
func (s *Accounts) TracingFields() []otlog.Field {
	return []otlog.Field{
		otlog.String("Accounts.ServiceType", s.ServiceType),
		otlog.String("Accounts.Perm", s.ServiceID),
		otlog.Int("Accounts.AccountIDs.Count", len(s.AccountIDs)),
	}
}

// AccountID is a descriptive type for the external identifier of an external account on the
// code host. It can be the string representation of an integer (e.g. GitLab), a GraphQL ID
// (e.g. GitHub), or a username (e.g. Bitbucket Server) depends on the code host type.
type AccountID string

// RepoID is a descriptive type for the external identifier of an external repository on the
// code host. It can be the string representation of an integer (e.g. GitLab and Bitbucket
// Server) or a GraphQL ID (e.g. GitHub) depends on the code host type.
type RepoID string

// ParseConfig attempts to unmarshal the given JSON config into a configuration struct defined in the schema package.
func ParseConfig(kind, config string) (cfg interface{}, _ error) {
	switch strings.ToLower(kind) {
	case "awscodecommit":
		cfg = &schema.AWSCodeCommitConnection{}
	case "bitbucketserver":
		cfg = &schema.BitbucketServerConnection{}
	case "github":
		cfg = &schema.GitHubConnection{}
	case "gitlab":
		cfg = &schema.GitLabConnection{}
	case "gitolite":
		cfg = &schema.GitoliteConnection{}
	case "phabricator":
		cfg = &schema.PhabricatorConnection{}
	case "other":
		cfg = &schema.OtherExternalServiceConnection{}
	default:
		return nil, fmt.Errorf("unknown external service kind %q", kind)
	}
	return cfg, jsonc.Unmarshal(config, cfg)
}

const IDParam = "externalServiceID"

func WebhookURL(kind string, externalServiceID int64, externalURL string) (string, error) {
	var path string
	switch strings.ToLower(kind) {
	case "github":
		path = "github-webhooks"
	case "bitbucketserver":
		path = "bitbucket-server-webhooks"
	default:
		return "", fmt.Errorf("unknown external service kind: %q", kind)
	}
	// eg. https://example.com/.api/github-webhooks?externalServiceID=1
	return fmt.Sprintf("%s/.api/%s?%s=%d", externalURL, path, IDParam, externalServiceID), nil
}
