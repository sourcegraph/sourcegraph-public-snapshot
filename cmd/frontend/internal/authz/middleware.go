package authz

import (
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/licensing/enforcement"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
)

// Enforce the use of a valid license key by preventing all HTTP requests if the license is invalid
// (due to an error in parsing or verification, or because the license has expired).
func PostAuthMiddleware(logger log.Logger, db database.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := actor.FromContext(r.Context())
		// Ignore not authenticated users, because we need to allow site admins
		// to sign in to set a license.
		if !a.IsAuthenticated() {
			next.ServeHTTP(w, r)
			return
		}

		siteadminOrHandler := func(handler func()) {
			err := auth.CheckCurrentUserIsSiteAdmin(r.Context(), db)
			if err == nil {
				// User is site admin, let them proceed.
				next.ServeHTTP(w, r)
				return
			}
			if err != auth.ErrMustBeSiteAdmin {
				logger.Error("Error checking current user is site admin", log.Error(err))
				http.Error(w, "Error checking current user is site admin. Site admins may check the logs for more information.", http.StatusInternalServerError)
				return
			}

			handler()
		}

		// Check if there are any license issues. If so, don't let the request go through.
		// Exception: Site admins are exempt from license enforcement screens so that they
		// can easily update the license key. We only fetch the user if we don't have a license,
		// to save that DB lookup in most cases.
		info, err := licensing.GetConfiguredProductLicenseInfo()
		if err != nil {
			logger.Error("Error reading license key for Sourcegraph subscription.", log.Error(err))
			siteadminOrHandler(func() {
				enforcement.WriteSubscriptionErrorResponse(w, http.StatusInternalServerError, "Error reading Sourcegraph license key", "Site admins may check the logs for more information. Update the license key in the [**site configuration**](/site-admin/configuration).")
			})
			return
		}
		if info != nil && info.IsExpired() {
			siteadminOrHandler(func() {
				enforcement.WriteSubscriptionErrorResponse(w, http.StatusForbidden, "Sourcegraph license expired", "To continue using Sourcegraph, a site admin must renew the Sourcegraph license (or downgrade to only using Sourcegraph Free features). Update the license key in the [**site configuration**](/site-admin/configuration).")
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
