package init

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/licensing/enforcement"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/licensing/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var (
	enableUpcomingLicenseExpirationChecker = env.MustGetBool("DOTCOM_ENABLE_UPCOMING_LICENSE_EXPIRATION_CHECKER", true,
		"If false, we do not monitor for upcoming license expirations to post in Slack.")
	enableAnomalousLicenseChecker = env.MustGetBool("DOTCOM_ENABLE_ANOMALOUS_LICENSE_CHECKER", true,
		"If false, we do not monitor for anomalous license checks to post in Slack.")
)

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

	enterpriseServices.LicenseResolver = resolvers.LicenseResolver{}

	if dotcom.SourcegraphDotComMode() {
		logger := log.Scoped("licensing")
		if enableUpcomingLicenseExpirationChecker {
			goroutine.Go(func() {
				productsubscription.StartCheckForUpcomingLicenseExpirations(logger, db)
			})
		}
		if enableAnomalousLicenseChecker {
			goroutine.Go(func() {
				productsubscription.StartCheckForAnomalousLicenseUsage(logger, db)
			})
		}
	}

	return nil
}
