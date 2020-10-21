package init

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/graphqlbackend"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/schema"
)

// TODO(efritz) - de-globalize assignments in this function
// TODO(efritz) - refactor licensing packages - this is a huge mess!
func Init(ctx context.Context, enterpriseServices *enterprise.Services) error {
	// Enforce the license's max user count by preventing the creation of new users when the max is
	// reached.
	db.Users.PreCreateUser = licensing.NewPreCreateUserHook(&usersStore{})

	// Make the Site.productSubscription.productNameWithBrand GraphQL field (and other places) use the
	// proper product name.
	graphqlbackend.GetProductNameWithBrand = licensing.ProductNameWithBrand

	conf.Branding = func() *schema.Branding {
		if !licensing.EnforceTiers ||
			licensing.CheckFeature(licensing.FeatureBranding) == nil {
			branding := conf.Get().Branding
			if branding != nil && branding.BrandName == "" {
				bcopy := *branding
				bcopy.BrandName = "Sourcegraph"
				branding = &bcopy
			}
			return branding
		}
		return nil
	}

	graphqlbackend.AlertFuncs = append(graphqlbackend.AlertFuncs, func(args graphqlbackend.AlertFuncArgs) []*graphqlbackend.Alert {
		if !licensing.EnforceTiers {
			return nil
		}

		// Only site admins can act on this alert, so only show it to site admins.
		if !args.IsSiteAdmin {
			return nil
		}

		if licensing.IsFeatureEnabledLenient(licensing.FeatureBranding) {
			return nil
		}

		return []*graphqlbackend.Alert{{
			TypeValue:    graphqlbackend.AlertTypeError,
			MessageValue: "A Sourcegraph license is required to custom branding for the instance. [**Get a license.**](/site-admin/license)",
		}}
	})

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

	goroutine.Go(func() {
		licensing.StartMaxUserCount(&usersStore{})
	})
	if envvar.SourcegraphDotComMode() {
		goroutine.Go(productsubscription.StartCheckForUpcomingLicenseExpirations)
	}

	return nil
}

type usersStore struct{}

func (usersStore) Count(ctx context.Context) (int, error) {
	return db.Users.Count(ctx, nil)
}
