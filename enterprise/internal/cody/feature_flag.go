package cody

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// IsCodyEnabled determines if cody is enabled for the actor in the given context.
// If it is an unauthenticated request, cody is disabled.
// If authenticated it checks if cody is enabled for the deployment type
func IsCodyEnabled(ctx context.Context) bool {
	if err := licensing.Check(licensing.FeatureCody); err != nil {
		return false
	}

	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return false
	}

	if deploy.IsApp() {
		return isCodyEnabledInApp()
	}

	return isCodyEnabled(ctx)
}

// isCodyEnabled determines if cody is enabled for the actor in the given context
// for all deployment types except "app".
// If Completions aren't configured, cody is disabled.
// If Completions are not enabled, cody is disabled
// If CodyRestrictUsersFeatureFlag is set, the cody-experimental featureflag
// will determine access.
// Otherwise, all authenticated users are granted access.
func isCodyEnabled(ctx context.Context) bool {
	completionsConfig := conf.Get().Completions
	if completionsConfig == nil {
		return false
	}

	if !completionsConfig.Enabled {
		return false
	}

	experimentalFeatures := conf.Get().ExperimentalFeatures
	if experimentalFeatures != nil && experimentalFeatures.CodyRestrictUsersFeatureFlag {
		return featureflag.FromContext(ctx).GetBoolOr("cody-experimental", false)
	}
	return true
}

// isCodyEnabledInApp determines if cody is enabled within Sourcegraph App.
// If Completions are configured and enabled, cody is enabled.
// If the App user's dotcom auth token is present, cody is enabled.
// In all other cases Cody is disabled.
func isCodyEnabledInApp() bool {
	completionsConfig := conf.Get().Completions
	appConfig := conf.Get().App
	if completionsConfig != nil && completionsConfig.Enabled {
		return true
	}

	if appConfig != nil && len(appConfig.DotcomAuthToken) > 0 {
		return true
	}

	return false
}

var ErrRequiresVerifiedEmailAddress = errors.New("cody requires a verified email address")

func CheckVerifiedEmailRequirement(ctx context.Context, db database.DB, logger log.Logger) error {
	// Only check on dotcom
	if !envvar.SourcegraphDotComMode() {
		return nil
	}

	// Do not require if user is site-admin
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, db); err == nil {
		return nil
	}

	verified, err := backend.NewUserEmailsService(db, logger).CurrentActorHasVerifiedEmail(ctx)
	if err != nil {
		return err
	}
	if verified {
		return nil
	}

	return ErrRequiresVerifiedEmailAddress
}
