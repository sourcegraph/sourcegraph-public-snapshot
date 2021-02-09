package database

import (
	"github.com/sourcegraph/sourcegraph/internal/authz/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/authz/github"
	"github.com/sourcegraph/sourcegraph/internal/authz/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/authz/perforce"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewExternalServicesStore returns an OSS database.ExternalServicesStore set with
// enterprise validators.
func NewExternalServicesStore(d dbutil.DB) *database.ExternalServiceStore {
	es := database.ExternalServices(d)
	es.GitHubValidators = []func(*schema.GitHubConnection) error{
		github.ValidateAuthz,
	}
	es.GitLabValidators = []func(*schema.GitLabConnection, []schema.AuthProviders) error{
		gitlab.ValidateAuthz,
	}
	es.BitbucketServerValidators = []func(*schema.BitbucketServerConnection) error{
		bitbucketserver.ValidateAuthz,
	}
	es.PerforceValidators = []func(connection *schema.PerforceConnection) error{
		perforce.ValidateAuthz,
	}

	return es
}
