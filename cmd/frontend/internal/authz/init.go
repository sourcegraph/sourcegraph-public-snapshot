package authz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/authz/resolvers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/authz/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/authz/providers"
	srp "github.com/sourcegraph/sourcegraph/internal/authz/subrepoperms"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

var clock = timeutil.Now

func Init(
	ctx context.Context,
	observationCtx *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	database.ValidateExternalServiceConfig = providers.ValidateExternalServiceConfig
	database.AuthzWith = func(other basestore.ShareableStore) database.AuthzStore {
		return database.NewAuthzStore(observationCtx.Logger, db, clock)
	}

	// TODO(nsc): use c
	// Report any authz provider problems in external configs.
	conf.ContributeWarning(func(cfg conftypes.SiteConfigQuerier) (problems conf.Problems) {
		_, providers, seriousProblems, warnings, _ := providers.ProvidersFromConfig(ctx, cfg, db)
		problems = append(problems, conf.NewExternalServiceProblems(seriousProblems...)...)

		// Validating the connection may make a cross service call, so we should use an
		// internal actor.
		ctx := actor.WithInternalActor(ctx)

		// Add connection validation issue
		for _, p := range providers {
			if err := p.ValidateConnection(ctx); err != nil {
				warnings = append(warnings, fmt.Sprintf("%s provider %q: %s", p.ServiceType(), p.ServiceID(), err))
			}
		}

		problems = append(problems, conf.NewExternalServiceProblems(warnings...)...)
		return problems
	})

	enterpriseServices.PermissionsGitHubWebhook = webhooks.NewGitHubWebhook(log.Scoped("PermissionsGitHubWebhook"))

	authz.DefaultSubRepoPermsChecker = srp.NewSubRepoPermsClient(db.SubRepoPerms())

	graphqlbackend.AlertFuncs = append(graphqlbackend.AlertFuncs, func(args graphqlbackend.AlertFuncArgs) []*graphqlbackend.Alert {
		if licensing.IsLicenseValid() {
			return nil
		}

		reason := licensing.GetLicenseInvalidReason()

		return []*graphqlbackend.Alert{{
			TypeValue:    graphqlbackend.AlertTypeError,
			MessageValue: fmt.Sprintf("The Sourcegraph license key is invalid. Reason: %s. To continue using Sourcegraph, a site admin must renew the Sourcegraph license (or downgrade to only using Sourcegraph Free features). Update the license key in the [**site configuration**](/site-admin/configuration). Please contact Sourcegraph support for more information.", reason),
		}}
	})

	// Warn about usage of authz providers that are not enabled by the license.
	graphqlbackend.AlertFuncs = append(graphqlbackend.AlertFuncs, func(args graphqlbackend.AlertFuncArgs) []*graphqlbackend.Alert {
		// Only site admins can act on this alert, so only show it to site admins.
		if !args.IsSiteAdmin {
			return nil
		}

		if licensing.IsFeatureEnabledLenient(licensing.FeatureACLs) {
			return nil
		}

		_, _, _, _, invalidConnections := providers.ProvidersFromConfig(ctx, conf.Get(), db)

		// We currently support three types of authz providers: GitHub, GitLab and Bitbucket Server.
		authzTypes := make(map[string]struct{}, 3)
		for _, conn := range invalidConnections {
			authzTypes[conn] = struct{}{}
		}

		authzNames := make([]string, 0, len(authzTypes))
		for t := range authzTypes {
			switch t {
			case extsvc.TypeGitHub:
				authzNames = append(authzNames, "GitHub")
			case extsvc.TypeGitLab:
				authzNames = append(authzNames, "GitLab")
			case extsvc.TypeBitbucketServer:
				authzNames = append(authzNames, "Bitbucket Server")
			default:
				authzNames = append(authzNames, t)
			}
		}

		if len(authzNames) == 0 {
			return nil
		}

		return []*graphqlbackend.Alert{{
			TypeValue:    graphqlbackend.AlertTypeError,
			MessageValue: fmt.Sprintf("A Sourcegraph license is required to enable repository permissions for the following code hosts: %s. [**Get a license.**](/site-admin/license)", strings.Join(authzNames, ", ")),
		}}
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
			observationCtx.Logger.Error("Error reading license key for Sourcegraph subscription.", log.Error(err))
			return []*graphqlbackend.Alert{{
				TypeValue:    graphqlbackend.AlertTypeError,
				MessageValue: "Error reading Sourcegraph license key. Check the logs for more information, or update the license key in the [**site configuration**](/site-admin/configuration).",
			}}
		}
		if info != nil && info.IsExpired() {
			return []*graphqlbackend.Alert{{
				TypeValue:    graphqlbackend.AlertTypeError,
				MessageValue: "Sourcegraph license expired! All non-admin users are locked out of Sourcegraph. Update the license key in the [**site configuration**](/site-admin/configuration) or downgrade to only using Sourcegraph Free features.",
			}}
		}
		if info != nil && info.IsExpiringSoon() {
			return []*graphqlbackend.Alert{{
				TypeValue:    graphqlbackend.AlertTypeWarning,
				MessageValue: fmt.Sprintf("Sourcegraph license will expire soon! Expires on: %s. Update the license key in the [**site configuration**](/site-admin/configuration) or downgrade to only using Sourcegraph Free features.", info.ExpiresAt.UTC().Truncate(time.Hour).Format(time.UnixDate)),
			}}
		}
		return nil
	})

	go func() {
		t := time.NewTicker(5 * time.Second)
		for range t.C {
			allowAccessByDefault, authzProviders, _, _, _ := providers.ProvidersFromConfig(ctx, conf.Get(), db)
			authz.SetProviders(allowAccessByDefault, authzProviders)
		}
	}()

	enterpriseServices.AuthzResolver = resolvers.NewResolver(observationCtx, db)
	return nil
}
