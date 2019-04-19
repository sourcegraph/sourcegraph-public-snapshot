package db

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/schema"
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
	}
}
