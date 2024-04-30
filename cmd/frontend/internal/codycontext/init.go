package codycontext

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Init(
	ctx context.Context,
	_ *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	_ *enterprise.Services,
) error {
	fmt.Println("codycontext.Init")
	return maybeValidateSiteConfig(ctx, db)
}

// maybeValidateSiteConfig checks "cody-context-filters-enabled" feature flag value and, if it's set, adds a config validator
// that returns an error if `CodyContextFilters` are set in the site config.
//
// We perform this check to prevent users from setting `CodyContextFilters` in the site config as it is not supported
// by Cody IDE clients (VSCode and JetBrains) yet.
//
// TODO: remove this check after `CodyContextFilters` support is added to the IDE clients.
func maybeValidateSiteConfig(ctx context.Context, db database.DB) error {
	enabled, err := checkFeatureFlagEnabled(ctx, db)
	if err != nil {
		return err
	}
	if !enabled {
		conf.ContributeValidator(func(confQuerier conftypes.SiteConfigQuerier) (problems conf.Problems) {
			if confQuerier.SiteConfig().CodyContextFilters != nil {
				problems = append(problems, conf.NewSiteProblem("\"cody.contextFilters\" param can't be set as it is not supported by Cody IDE clients (VS Code and JetBrains) yet. For information on when IDE support will be available, please visit our documentation: https://sourcegraph.com/docs/cody/capabilities/ignore-context#cody-ignore."))
			}
			return problems
		})
	}
	return nil
}

// TODO: remove this check after `CodyContextFilters` support is added to the IDE clients.
func checkFeatureFlagEnabled(ctx context.Context, db database.DB) (bool, error) {
	flag, err := db.FeatureFlags().GetFeatureFlag(ctx, "cody-context-filters-enabled")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	return flag != nil && flag.Bool.Value, nil
}
