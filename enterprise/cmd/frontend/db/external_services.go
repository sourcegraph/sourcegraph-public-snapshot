package db

import (
	"sourcegraph.com/cmd/frontend/db"
	"sourcegraph.com/enterprise/cmd/frontend/internal/authz"
	"sourcegraph.com/schema"
)

// NewExternalServicesStore returns an OSS db.ExternalServicesStore set with
// enterprise validators.
func NewExternalServicesStore() *db.ExternalServicesStore {
	return &db.ExternalServicesStore{
		GitHubValidators: []func(*schema.GitHubConnection) error{
			authz.ValidateGitHubAuthz,
		},
		GitLabValidators: []func(*schema.GitLabConnection, []schema.AuthProviders) error{
			authz.ValidateGitLabAuthz,
		},
		BitbucketServerValidators: []func(*schema.BitbucketServerConnection, []schema.AuthProviders) error{
			authz.ValidateBitbucketServerAuthz,
		},
	}
}
