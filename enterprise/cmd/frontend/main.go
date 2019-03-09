// Command frontend contains the enterprise frontend implementation.
//
// It wraps the open source frontend command and merely injects a few
// proprietary things into it via e.g. blank/underscore imports in this file
// which register side effects with the frontend package.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hooks"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/graphqlbackend"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/httpapi"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry"
	"gopkg.in/inconshreveable/log15.v2"
)

func main() {
	// Make the Site.productSubscription.productNameWithBrand GraphQL field (and other places) use the
	// proper product name.
	graphqlbackend.GetProductNameWithBrand = licensing.ProductNameWithBrand

	// Make the Site.productSubscription.actualUserCount and Site.productSubscription.actualUserCountDate
	// GraphQL fields return the proper max user count and timestamp on the current license.
	graphqlbackend.ActualUserCount = licensing.ActualUserCount
	graphqlbackend.ActualUserCountDate = licensing.ActualUserCountDate

	noLicenseMaximumAllowedUserCount := licensing.NoLicenseMaximumAllowedUserCount
	graphqlbackend.NoLicenseMaximumAllowedUserCount = &noLicenseMaximumAllowedUserCount

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

	// Enforce the use of a valid license key by preventing all HTTP requests if the license is invalid
	// (due to a error in parsing or verification, or because the license has expired).
	hooks.PreAuthMiddleware = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info, err := licensing.GetConfiguredProductLicenseInfo()
			if err != nil {
				log15.Error("Error reading license key for Sourcegraph subscription.", "err", err)
				licensing.WriteSubscriptionErrorResponse(w, http.StatusInternalServerError, "Error reading Sourcegraph license key", "Site admins may check the logs for more information. Update the license key in the Sourcegraph management console (https://docs.sourcegraph.com/admin/management_console).")
				return
			}
			if info != nil && info.IsExpiredWithGracePeriod() {
				licensing.WriteSubscriptionErrorResponse(w, http.StatusForbidden, "Sourcegraph license expired", "To continue using Sourcegraph, a site admin must renew the Sourcegraph license (or downgrade to only using Sourcegraph Core features). Update the license key in the Sourcegraph management console.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	hooks.AfterDBInit = func() {
		authz.Init(db.ExternalServices)

		// Enforce the license's max user count by preventing the creation of new users when the max is
		// reached.
		db.Users.PreCreateUser = licensing.NewPreCreateUserHook(&usersStore{})

		go licensing.StartMaxUserCount(&usersStore{})
	}

	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		log.Println("enterprise edition")
	}

	shared.Main()
}

type usersStore struct{}

func (usersStore) Count(ctx context.Context) (int, error) {
	return db.Users.Count(ctx, nil)
}
