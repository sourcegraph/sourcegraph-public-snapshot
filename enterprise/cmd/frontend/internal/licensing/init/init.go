package init

import (
	"context"
	"fmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hooks"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing/enforcement"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing/resolvers"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
)

// TODO(efritz) - de-globalize assignments in this function
// TODO(efritz) - refactor licensing packages - this is a huge mess!
func Init(
	ctx context.Context,
	observationCtx *observation.Context,
	db database.DB,
	codeIntelServices codeintel.Services,
	conf conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	// Enforce the license's max user count by preventing the creation of new users when the max is
	// reached.
	database.BeforeCreateUser = enforcement.NewBeforeCreateUserHook()

	// Enforce non-site admin roles in Free tier.
	database.AfterCreateUser = enforcement.NewAfterCreateUserHook()

	// Enforce site admin creation rules.
	database.BeforeSetUserIsSiteAdmin = enforcement.NewBeforeSetUserIsSiteAdmin()

	// Enforce the license's max external service count by preventing the creation of new external
	// services when the max is reached.
	database.BeforeCreateExternalService = enforcement.NewBeforeCreateExternalServiceHook()

	logger := log.Scoped("licensing", "licensing enforcement")
	hooks.GetLicenseInfo = func(isSiteAdmin bool) *hooks.LicenseInfo {
		if !isSiteAdmin {
			return nil
		}

		info, err := licensing.GetConfiguredProductLicenseInfo()
		if err != nil {
			logger.Error("Failed to get license info", log.Error(err))
			return nil
		}

		if info.Plan() == licensing.PlanFree0 {
			// We don't enforce anything on the free plan
			return nil
		}

		licenseInfo := &hooks.LicenseInfo{
			CurrentPlan: string(info.Plan()),
		}
		if info.Plan() == licensing.PlanBusiness0 {
			const codeScaleLimit = 100 * 1024 * 1024 * 1024
			licenseInfo.CodeScaleLimit = "100GiB"

			stats, err := usagestats.GetRepositories(ctx, db)
			if err != nil {
				logger.Error("Failed to get repository stats", log.Error(err))
				return nil
			}

			if stats.GitDirBytes >= codeScaleLimit {
				licenseInfo.CodeScaleExceededLimit = true
			} else if stats.GitDirBytes >= codeScaleLimit*0.9 {
				licenseInfo.CodeScaleCloseToLimit = true
			}
		}

		return licenseInfo
	}

	// Enforce the license's feature check for monitoring. If the license does not support the monitoring
	// feature, then alternative debug handlers will be invoked.
	// Uncomment this when licensing for FeatureMonitoring should be enforced.
	// See PR https://github.com/sourcegraph/sourcegraph/issues/42527 for more context.
	// app.SetPreMountGrafanaHook(enforcement.NewPreMountGrafanaHook())

	// Make the Site.productSubscription.productNameWithBrand GraphQL field (and other places) use the
	// proper product name.
	graphqlbackend.GetProductNameWithBrand = licensing.ProductNameWithBrand

	// Make the Site.productSubscription.actualUserCount and Site.productSubscription.actualUserCountDate
	// GraphQL fields return the proper max user count and timestamp on the current license.
	graphqlbackend.ActualUserCount = licensing.ActualUserCount
	graphqlbackend.ActualUserCountDate = licensing.ActualUserCountDate

	noLicenseMaximumAllowedUserCount := licensing.NoLicenseMaximumAllowedUserCount
	graphqlbackend.NoLicenseMaximumAllowedUserCount = &noLicenseMaximumAllowedUserCount

	noLicenseWarningUserCount := licensing.NoLicenseWarningUserCount
	graphqlbackend.NoLicenseWarningUserCount = &noLicenseWarningUserCount

	// Make the Site.productSubscription GraphQL field return the actual info about the product license,
	// if any.
	graphqlbackend.GetConfiguredProductLicenseInfo = func() (*graphqlbackend.ProductLicenseInfo, error) {
		info, err := licensing.GetConfiguredProductLicenseInfo()
		if info == nil || err != nil {
			return nil, err
		}
		return &graphqlbackend.ProductLicenseInfo{
			TagsValue:      info.Tags,
			UserCountValue: info.UserCount,
			ExpiresAtValue: info.ExpiresAt,
		}, nil
	}

	graphqlbackend.IsFreePlan = func(info *graphqlbackend.ProductLicenseInfo) bool {
		for _, tag := range info.Tags() {
			if tag == fmt.Sprintf("plan:%s", licensing.PlanFree0) {
				return true
			}
		}

		return false
	}

	enterpriseServices.LicenseResolver = resolvers.LicenseResolver{}

	goroutine.Go(func() {
		licensing.StartMaxUserCount(logger, &usersStore{
			db: db,
		})
	})
	if envvar.SourcegraphDotComMode() {
		goroutine.Go(func() {
			productsubscription.StartCheckForUpcomingLicenseExpirations(logger, db)
		})
	}

	return nil
}

type usersStore struct {
	db database.DB
}

func (u *usersStore) Count(ctx context.Context) (int, error) {
	return u.db.Users().Count(
		ctx,
		&database.UsersListOptions{
			ExcludeSourcegraphOperators: true,
		},
	)
}
