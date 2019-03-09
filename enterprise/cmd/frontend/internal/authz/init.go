package authz

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

type ExternalServicesStore interface {
	ListGitLabConnections(context.Context) ([]*schema.GitLabConnection, error)
	ListGitHubConnections(context.Context) ([]*schema.GitHubConnection, error)
}

func Init(s ExternalServicesStore) {
	// Warn about usage of auth providers that are not enabled by the license.
	graphqlbackend.AlertFuncs = append(graphqlbackend.AlertFuncs, func(args graphqlbackend.AlertFuncArgs) []*graphqlbackend.Alert {
		// Only site admins can act on this alert, so only show it to site admins.
		if !args.IsSiteAdmin {
			return nil
		}

		if licensing.IsFeatureEnabledLenient(licensing.FeatureACLs) {
			return nil
		}

		var authzTypes []string
		ctx := context.Background()

		githubs, err := s.ListGitHubConnections(ctx)
		if err != nil {
			return []*graphqlbackend.Alert{{
				TypeValue:    graphqlbackend.AlertTypeError,
				MessageValue: fmt.Sprintf("Unable to fetch GitHub external services: %s", err),
			}}
		}
		for _, g := range githubs {
			if g.Authorization != nil {
				authzTypes = append(authzTypes, "GitHub")
				break
			}
		}

		gitlabs, err := s.ListGitLabConnections(ctx)
		if err != nil {
			return []*graphqlbackend.Alert{{
				TypeValue:    graphqlbackend.AlertTypeError,
				MessageValue: fmt.Sprintf("Unable to fetch GitLab external services: %s", err),
			}}
		}
		for _, g := range gitlabs {
			if g.Authorization != nil {
				authzTypes = append(authzTypes, "GitLab")
				break
			}
		}

		if len(authzTypes) > 0 {
			return []*graphqlbackend.Alert{{
				TypeValue:    graphqlbackend.AlertTypeError,
				MessageValue: fmt.Sprintf("A Sourcegraph license is required to enable repository permissions for the following code hosts: %s. [**Get a license.**](/site-admin/license)", strings.Join(authzTypes, ", ")),
			}}
		}
		return nil
	})

	path, _ := os.Executable()
	isTest := filepath.Ext(path) == ".test"
	if isTest {
		return
	}

	ctx := context.Background()
	go func() {
		t := time.NewTicker(5 * time.Second)
		for range t.C {
			allowAccessByDefault, authzProviders, _, _ := providersFromConfig(ctx, conf.Get(), s)
			authz.SetProviders(allowAccessByDefault, authzProviders)
		}
	}()
}

// providersFromConfig returns the set of permission-related providers derived from the site config.
// It also returns any validation problems with the config, separating these into "serious problems"
// and "warnings".  "Serious problems" are those that should make Sourcegraph set
// authz.allowAccessByDefault to false. "Warnings" are all other validation problems.
func providersFromConfig(
	ctx context.Context,
	cfg *conf.Unified,
	s ExternalServicesStore,
) (
	allowAccessByDefault bool,
	authzProviders []authz.Provider,
	seriousProblems []string,
	warnings []string,
) {
	allowAccessByDefault = true
	defer func() {
		if len(seriousProblems) > 0 {
			log15.Error("Repository authz config was invalid (errors are visible in the UI as an admin user, you should fix ASAP). Restricting access to repositories by default for now to be safe.")
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

	return allowAccessByDefault, authzProviders, seriousProblems, warnings
}
