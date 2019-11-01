package authz

import (
	"context"
	"database/sql"
	"fmt"

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

// ProvidersFromConfig returns the set of permission-related providers derived from the site config.
// It also returns any validation problems with the config, separating these into "serious problems"
// and "warnings".  "Serious problems" are those that should make Sourcegraph set
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

	if gitlabs, err := s.ListGitLabConnections(ctx); err != nil {
		seriousProblems = append(seriousProblems, fmt.Sprintf("Could not load GitLab external service configs: %s", err))
	} else {
		glp, glproblems, glwarnings := gitlabProviders(ctx, cfg, gitlabs)
		authzProviders = append(authzProviders, glp...)
		seriousProblems = append(seriousProblems, glproblems...)
		warnings = append(warnings, glwarnings...)
	}

	if githubs, err := s.ListGitHubConnections(ctx); err != nil {
		seriousProblems = append(seriousProblems, fmt.Sprintf("Could not load GitHub external service configs: %s", err))
	} else {
		ghp, ghproblems, ghwarnings := githubProviders(ctx, githubs)
		authzProviders = append(authzProviders, ghp...)
		seriousProblems = append(seriousProblems, ghproblems...)
		warnings = append(warnings, ghwarnings...)
	}

	if bitbucketServers, err := s.ListBitbucketServerConnections(ctx); err != nil {
		seriousProblems = append(seriousProblems, fmt.Sprintf("Could not load Bitbucket Server external service configs: %s", err))
	} else {
		ps, problems, warns := bitbucketServerProviders(ctx, db, cfg, bitbucketServers)
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
