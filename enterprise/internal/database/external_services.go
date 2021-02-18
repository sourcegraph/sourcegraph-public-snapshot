package database

import (
	"github.com/sourcegraph/sourcegraph/internal/authz/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/authz/github"
	"github.com/sourcegraph/sourcegraph/internal/authz/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewExternalServicesStore returns an OSS database.ExternalServicesStore set with
// enterprise validators.
func NewExternalServicesStore(d dbutil.DB, key encryption.Key) *database.ExternalServiceStore {
	es := database.ExternalServices(d, key)
	es.GitHubValidators = []func(*schema.GitHubConnection) error{
		github.ValidateAuthz,
	}
	es.GitLabValidators = []func(*schema.GitLabConnection, []schema.AuthProviders) error{
		gitlab.ValidateAuthz,
	}
	es.BitbucketServerValidators = []func(*schema.BitbucketServerConnection) error{
		bitbucketserver.ValidateAuthz,
	}

	return es
}
