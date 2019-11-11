package db

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/github"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/gitlab"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewExternalServicesStore returns an OSS db.ExternalServicesStore set with
// enterprise validators.
func NewExternalServicesStore() *db.ExternalServicesStore {
	return &db.ExternalServicesStore{
		GitHubValidators: []func(*schema.GitHubConnection) error{
			github.ValidateAuthz,
		},
		GitLabValidators: []func(*schema.GitLabConnection, []schema.AuthProviders) error{
			gitlab.ValidateAuthz,
		},
		BitbucketServerValidators: []func(*schema.BitbucketServerConnection) error{
			bitbucketserver.ValidateAuthz,
		},
	}
}
