package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
)

func (r *siteResolver) NeedsRepositoryConfiguration(ctx context.Context) (bool, error) {
	// 🚨 SECURITY: The site alerts may contain sensitive data, so only site
	// admins may view them.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		// TODO(dax): This should return err once the site flags query is fixed for users
		return false, nil
	}

	return needsRepositoryConfiguration(ctx, r.db)
}

func needsRepositoryConfiguration(ctx context.Context, db database.DB) (bool, error) {
	kinds := make([]string, 0, len(database.ExternalServiceKinds))
	for kind, config := range database.ExternalServiceKinds {
		if config.CodeHost {
			kinds = append(kinds, kind)
		}
	}

	count, err := db.ExternalServices().Count(ctx, database.ExternalServicesListOptions{
		Kinds: kinds,
	})
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (*siteResolver) SendsEmailVerificationEmails() bool { return conf.EmailVerificationRequired() }

func (r *siteResolver) FreeUsersExceeded(ctx context.Context) (bool, error) {
	if dotcom.SourcegraphDotComMode() {
		return false, nil
	}

	info, err := getConfiguredProductLicenseInfo()
	if err != nil {
		return false, err
	}
	// Only show alert if the license is a free plan.
	if !info.Plan.IsFreePlan() {
		return false, nil
	}

	userCount, err := r.db.Users().Count(
		ctx,
		&database.UsersListOptions{
			ExcludeSourcegraphOperators: true,
		},
	)
	if err != nil {
		return false, err
	}

	return licensing.NoLicenseWarningUserCount <= int32(userCount), nil
}

func (r *siteResolver) ExternalServicesFromFile() bool { return envvar.ExtsvcConfigFile() != "" }
func (r *siteResolver) AllowEditExternalServicesWithFile() bool {
	return envvar.ExtsvcConfigAllowEdits()
}
