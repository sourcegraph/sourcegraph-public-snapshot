package authz

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

type ExternalServicesStore interface {
	ListGitLabConnections(context.Context) ([]*schema.GitLabConnection, error)
	ListGitHubConnections(context.Context) ([]*schema.GitHubConnection, error)
	ListBitbucketServerConnections(context.Context) ([]*schema.BitbucketServerConnection, error)
}

// ProviderRegister is a function that returns authz providers and any serious problems and warnings found.
type ProviderRegister = func(
	context.Context,
	*conf.Unified,
	ExternalServicesStore,
	*sql.DB,
) (_ []authz.Provider, problems []string, warnings []string)

var providerRegisters []ProviderRegister

func NewProviderRegister(r ProviderRegister) {
	providerRegisters = append(providerRegisters, r)
}

// ProvidersFromConfig returns the set of permission-related providers derived from the site config.
// It also returns any validation problems with the config, separating these into "serious problems"
// and "warnings". "Serious problems" are those that should make Sourcegraph set
// authz.allowAccessByDefault to false. "Warnings" are all other validation problems.
func ProvidersFromConfig(
	ctx context.Context,
	cfg *conf.Unified,
	s ExternalServicesStore,
	db *sql.DB, // Needed by Bitbucket Server authz provider
) (
	allowAccessByDefault bool,
	authzProviders []authz.Provider,
	seriousProblems []string,
	warnings []string,
) {
	allowAccessByDefault = true
	defer func() {
		if len(seriousProblems) > 0 {
			log15.Error("Repository authz config was invalid (errors are visible in the UI as an admin user, you should fix ASAP). Restricting access to repositories by default for now to be safe.", "seriousProblems", seriousProblems)
			allowAccessByDefault = false
		}
	}()

	for _, r := range providerRegisters {
		ps, problems, warns := r(ctx, cfg, s, db)
		authzProviders = append(authzProviders, ps...)
		seriousProblems = append(seriousProblems, problems...)
		warnings = append(warnings, warns...)
	}

	return allowAccessByDefault, authzProviders, seriousProblems, warnings
}

func init() {
	conf.ContributeWarning(func(cfg conf.Unified) (problems conf.Problems) {
		_, _, seriousProblems, warnings := ProvidersFromConfig(context.Background(), &cfg, db.ExternalServices, dbconn.Global)
		problems = append(problems, conf.NewExternalServiceProblems(seriousProblems...)...)
		problems = append(problems, conf.NewExternalServiceProblems(warnings...)...)
		return problems
	})
}
