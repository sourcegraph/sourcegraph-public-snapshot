package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hooks"
	edb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/github"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/gitlab"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/schema"
	"gopkg.in/inconshreveable/log15.v2"
)

func initAuthz() {
	db.ExternalServices = edb.NewExternalServicesStore()

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

	graphqlbackend.AlertFuncs = append(graphqlbackend.AlertFuncs, func(args graphqlbackend.AlertFuncArgs) []*graphqlbackend.Alert {
		// 🚨 SECURITY: Only the site admin should ever see this (all other users will see a hard-block
		// license expiration screen) about this. Leaking this wouldn't be a security vulnerability, but
		// just in case this method is changed to return more information, we lock it down.
		if !args.IsSiteAdmin {
			return nil
		}

		info, err := licensing.GetConfiguredProductLicenseInfo()
		if err != nil {
			log15.Error("Error reading license key for Sourcegraph subscription.", "err", err)
			return []*graphqlbackend.Alert{{
				TypeValue:    graphqlbackend.AlertTypeError,
				MessageValue: "Error reading Sourcegraph license key. Check the logs for more information, or update the license key in the [**site configuration**](/site-admin/configuration).",
			}}
		}
		if info != nil && info.IsExpiredWithGracePeriod() {
			return []*graphqlbackend.Alert{{
				TypeValue:    graphqlbackend.AlertTypeError,
				MessageValue: "Sourcegraph license expired! All non-admin users are locked out of Sourcegraph. Update the license key in the [**site configuration**](/site-admin/configuration) or downgrade to only using Sourcegraph Core features.",
			}}
		}
		return nil
	})

	// Enforce the use of a valid license key by preventing all HTTP requests if the license is invalid
	// (due to a error in parsing or verification, or because the license has expired).
	hooks.PostAuthMiddleware = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := backend.CheckCurrentUserIsSiteAdmin(r.Context()); err != nil {
				// Site admins are exempt from license enforcement screens such that they can
				// easily update the license key.
				next.ServeHTTP(w, r)
				return
			}

			info, err := licensing.GetConfiguredProductLicenseInfo()
			if err != nil {
				log15.Error("Error reading license key for Sourcegraph subscription.", "err", err)
				licensing.WriteSubscriptionErrorResponse(w, http.StatusInternalServerError, "Error reading Sourcegraph license key", "Site admins may check the logs for more information. Update the license key in the [**site configuration**](/site-admin/configuration).")
				return
			}
			if info != nil && info.IsExpiredWithGracePeriod() {
				licensing.WriteSubscriptionErrorResponse(w, http.StatusForbidden, "Sourcegraph license expired", "To continue using Sourcegraph, a site admin must renew the Sourcegraph license (or downgrade to only using Sourcegraph Core features). Update the license key in the [**site configuration**](/site-admin/configuration).")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type ExternalServicesStore interface {
	ListGitLabConnections(context.Context) ([]*schema.GitLabConnection, error)
	ListGitHubConnections(context.Context) ([]*schema.GitHubConnection, error)
	ListBitbucketServerConnections(context.Context) ([]*schema.BitbucketServerConnection, error)
}

// authzProvidersFromConfig returns the set of permission-related providers derived from the site config.
// It also returns any validation problems with the config, separating these into "serious problems"
// and "warnings". "Serious problems" are those that should make Sourcegraph set authz.allowAccessByDefault
// to false. "Warnings" are all other validation problems.
func authzProvidersFromConfig(
	ctx context.Context,
	cfg *conf.Unified,
	s ExternalServicesStore,
	db *sql.DB, // Needed by Bitbucket Server authz provider
) (
	allowAccessByDefault bool,
	providers []authz.Provider,
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

	if ghConns, err := s.ListGitHubConnections(ctx); err != nil {
		seriousProblems = append(seriousProblems, fmt.Sprintf("Could not load GitHub external service configs: %s", err))
	} else {
		ghProviders, ghProblems, ghWarnings := github.NewAuthzProviders(ghConns)
		providers = append(providers, ghProviders...)
		seriousProblems = append(seriousProblems, ghProblems...)
		warnings = append(warnings, ghWarnings...)
	}

	if glConns, err := s.ListGitLabConnections(ctx); err != nil {
		seriousProblems = append(seriousProblems, fmt.Sprintf("Could not load GitLab external service configs: %s", err))
	} else {
		glProviders, glProblems, glWarnings := gitlab.NewAuthzProviders(cfg, glConns)
		providers = append(providers, glProviders...)
		seriousProblems = append(seriousProblems, glProblems...)
		warnings = append(warnings, glWarnings...)
	}

	if bbsConns, err := s.ListBitbucketServerConnections(ctx); err != nil {
		seriousProblems = append(seriousProblems, fmt.Sprintf("Could not load Bitbucket Server external service configs: %s", err))
	} else {
		bbsProviders, bbsProblems, bbsWarnings := bitbucketserver.NewAuthzProviders(bbsConns, db)
		providers = append(providers, bbsProviders...)
		seriousProblems = append(seriousProblems, bbsProblems...)
		warnings = append(warnings, bbsWarnings...)
	}

	// 🚨 SECURITY: Warn the admin when both code host authz provider and the Sourcegraph authz provider are configured.
	if cfg.SiteConfiguration.PermissionsUserMapping != nil &&
		cfg.SiteConfiguration.PermissionsUserMapping.Enabled && len(providers) > 0 {
		serviceTypes := make([]string, len(providers))
		for i := range providers {
			serviceTypes[i] = strconv.Quote(providers[i].ServiceType())
		}
		msg := fmt.Sprintf(
			"The Sourcegraph permissions (`permissions.userMapping`) cannot be enabled when %s authorization providers are in use. Blocking access to all repositories until the conflict is resolved.",
			strings.Join(serviceTypes, ", "))
		seriousProblems = append(seriousProblems, msg)
	}

	return allowAccessByDefault, providers, seriousProblems, warnings
}

func init() {
	// Report any authz provider problems in external configs.
	conf.ContributeWarning(func(cfg conf.Unified) (problems conf.Problems) {
		_, _, seriousProblems, warnings :=
			authzProvidersFromConfig(context.Background(), &cfg, db.ExternalServices, dbconn.Global)
		problems = append(problems, conf.NewExternalServiceProblems(seriousProblems...)...)
		problems = append(problems, conf.NewExternalServiceProblems(warnings...)...)
		return problems
	})
}
