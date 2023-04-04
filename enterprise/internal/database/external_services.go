package database

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/authz/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/authz/github"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/authz/gitlab"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/authz/perforce"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

var ValidateExternalServiceConfig = database.MakeValidateExternalServiceConfigFunc([]func(*types.GitHubConnection) error{github.ValidateAuthz},
	[]func(*schema.GitLabConnection, []schema.AuthProviders) error{gitlab.ValidateAuthz},
	[]func(*schema.BitbucketServerConnection) error{bitbucketserver.ValidateAuthz},
	[]func(connection *schema.PerforceConnection) error{perforce.ValidateAuthz},
	[]func(connection *schema.AzureDevOpsConnection) error{func(_ *schema.AzureDevOpsConnection) error { return nil }}) // TODO: @varsanojidan switch this with actual authz once its implemented.
