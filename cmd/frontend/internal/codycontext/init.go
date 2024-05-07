package codycontext

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Init(
	ctx context.Context,
	observationCtx *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	_ *enterprise.Services,
) error {
	logger := observationCtx.Logger.Scoped("codycontext")
	maybeValidateSiteConfig(ctx, logger, db)
	return nil
}

// maybeValidateSiteConfig adds a config validator that checks if the "cody-context-filters-enabled" feature flag is enabled,
// and if it's not and site config has `CodyContextFilters` field defined, returns a validation error.
//
// We perform this check to prevent users from setting `CodyContextFilters` in the site config as it is not supported
// by Cody IDE clients (VSCode and JetBrains) yet.
//
// `CodyContextFilters` field is respected only by enterprise instances, thus this check is not needed for dotcom.
//
// TODO: remove this check after `CodyContextFilters` support is added to the IDE clients.
func maybeValidateSiteConfig(ctx context.Context, logger log.Logger, db database.DB) {
	if dotcom.SourcegraphDotComMode() {
		return
	}

	conf.ContributeValidator(func(confQuerier conftypes.SiteConfigQuerier) conf.Problems {
		if confQuerier.SiteConfig().CodyContextFilters == nil {
			return nil
		}
		enabled, err := checkFeatureFlagEnabled(ctx, db)
		if err != nil {
			logger.Error("Failed to get feature flag value", log.Error(err))
			return nil
		}
		if enabled {
			return nil
		}
		return conf.NewSiteProblems("\"cody.contextFilters\" param can't be set as it is not supported by Cody IDE clients (VS Code and JetBrains) yet. For information on when IDE support will be available, please visit our documentation: https://sourcegraph.com/docs/cody/capabilities/ignore-context#cody-ignore.")
	})
}

// TODO: remove this check after `CodyContextFilters` support is added to the IDE clients.
func checkFeatureFlagEnabled(ctx context.Context, db database.DB) (bool, error) {
	flag, err := db.FeatureFlags().GetFeatureFlag(ctx, "cody-context-filters-enabled")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	return flag != nil && flag.Bool.Value, nil
}
