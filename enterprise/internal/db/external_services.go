package db

import (
	"github.com/sourcegraph/sourcegraph/internal/authz/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/authz/github"
	"github.com/sourcegraph/sourcegraph/internal/authz/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewExternalServicesStore returns an OSS db.ExternalServicesStore set with
// enterprise validators.
func NewExternalServicesStore(d dbutil.DB) *db.ExternalServiceStore {
	es := db.ExternalServices(d)
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
