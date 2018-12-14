package authz

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func init() {
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

		githubs, err := db.ExternalServices.ListGitHubConnections(ctx)
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

		gitlabs, err := db.ExternalServices.ListGitLabConnections(ctx)
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
}

func init() {
	path, _ := os.Executable()
	isTest := filepath.Ext(path) == ".test"
	if isTest {
		return
	}

	ctx := context.Background()
	go func() {
		t := time.NewTicker(5 * time.Second)
		for range t.C {
			allowAccessByDefault, authzProviders, _, _ := providersFromConfig(ctx, conf.Get())
			authz.SetProviders(allowAccessByDefault, authzProviders)
		}
	}()
}

// providersFromConfig returns the set of permission-related providers derived from the site config.
// It also returns any validation problems with the config, separating these into "serious problems"
// and "warnings".  "Serious problems" are those that should make Sourcegraph set
// authz.allowAccessByDefault to false. "Warnings" are all other validation problems.
func providersFromConfig(ctx context.Context, cfg *conf.Unified) (
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

	glp, glproblems, glwarnings := gitlabProviders(ctx, cfg)
	authzProviders = append(authzProviders, glp...)
	seriousProblems = append(seriousProblems, glproblems...)
	warnings = append(warnings, glwarnings...)

	ghp, ghproblems, ghwarnings := githubProviders(ctx)
	authzProviders = append(authzProviders, ghp...)
	seriousProblems = append(seriousProblems, ghproblems...)
	warnings = append(warnings, ghwarnings...)

	return allowAccessByDefault, authzProviders, seriousProblems, warnings
}
