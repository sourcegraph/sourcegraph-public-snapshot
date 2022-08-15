package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func (r *siteResolver) NeedsRepositoryConfiguration(ctx context.Context) (bool, error) {
	if envvar.SourcegraphDotComMode() {
		return false, nil
	}

	// 🚨 SECURITY: The site alerts may contain sensitive data, so only site
	// admins may view them.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
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

func (*siteResolver) DisableBuiltInSearches() bool {
	return conf.Get().DisableBuiltInSearches
}

func (*siteResolver) SendsEmailVerificationEmails() bool { return conf.EmailVerificationRequired() }

func (r *siteResolver) FreeUsersExceeded(ctx context.Context) (bool, error) {
	if envvar.SourcegraphDotComMode() {
		return false, nil
	}

	// If a license exists, warnings never need to be shown.
	if info, err := GetConfiguredProductLicenseInfo(); info != nil {
		return false, err
	}
	// If OSS, warnings never need to be shown.
	if NoLicenseWarningUserCount == nil {
		return false, nil
	}

	userCount, err := r.db.Users().Count(ctx, nil)
	if err != nil {
		return false, err
	}

	return *NoLicenseWarningUserCount <= int32(userCount), nil
}
