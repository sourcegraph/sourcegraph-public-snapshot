package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	sglog "github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type (
	AuthValidator    func(context.Context, url.Values, string) (int, error)
	AuthValidatorMap = map[string]AuthValidator
)

var DefaultValidatorByCodeHost = AuthValidatorMap{
	"github.com": enforceAuthViaGitHub,
	"gitlab.com": enforceAuthViaGitLab,
}

var errVerificationNotSupported = errors.New(strings.Join([]string{
	"verification is supported for the following code hosts: github.com, gitlab.com",
	"please request support for additional code host verification at https://github.com/sourcegraph/sourcegraph/issues/4967",
}, " - "))

// AuthMiddleware wraps the given upload handler with an authorization check. On each initial upload
// request, the target repository is checked against the supplied auth validators. The matching validator
// is invoked, which coordinates with a remote code host's permissions API to determine if the current
// request contains sufficient evidence of authorship for the target repository.
//
// When LSIF auth is not enforced on the instance, this middleware no-ops.
func AuthMiddleware(
	next http.Handler,
	userStore UserStore,
	repoStore backend.ReposService,
	authValidators AuthValidatorMap,
	operation *observation.Operation,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusCode, err := func() (_ int, err error) {
			ctx, trace, endObservation := operation.With(r.Context(), &err, observation.Args{})
			defer endObservation(1, observation.Args{})

			query := r.URL.Query()
			repositoryName := getQuery(r, "repository")

			if isSiteAdmin(ctx, userStore, operation.Logger) {
				// If `AuthzEnforceForSiteAdmins` is set we should still check repo permissions for site admins
				if conf.Get().AuthzEnforceForSiteAdmins {
					statusCode, err := userCanAccessRepository(ctx, repoStore, repositoryName, trace)
					if err != nil {
						return statusCode, err
					}
				}
				return 0, nil
			}

			// Non site-admin users can upload indices if auth is not enabled in the instance's site configuration
			if !conf.Get().LsifEnforceAuth {
				statusCode, err := userCanAccessRepository(ctx, repoStore, repositoryName, trace)
				if err != nil {
					return statusCode, err
				}
				trace.AddEvent("bypassing code host auth check")
				return 0, nil
			}

			for codeHost, validator := range authValidators {
				if !strings.HasPrefix(repositoryName, codeHost) {
					continue
				}
				trace.AddEvent("TODO Domain Owner", attribute.String("codeHost", codeHost))

				return validator(ctx, query, repositoryName)
			}

			return http.StatusUnprocessableEntity, errVerificationNotSupported
		}()
		if err != nil {
			if statusCode >= 500 {
				operation.Logger.Error("codeintel.httpapi: failed to authorize request", sglog.Error(err))
			}

			http.Error(w, fmt.Sprintf("failed to authorize request: %s", err.Error()), statusCode)
			return
		}

		// 🚨 SECURITY: Bypass authz here; we've already determined that the current request is
		// authorized to view the target repository; they are either a site admin or the code
		// host has explicit listed them with some level of access (depending on the code host).
		internalReq := r.WithContext(actor.WithInternalActor(r.Context()))
		next.ServeHTTP(w, internalReq)
	})
}

func userCanAccessRepository(ctx context.Context, repoStore backend.ReposService, repositoryName string, trace observation.TraceLogger) (int, error) {
	_, err := repoStore.GetByName(ctx, api.RepoName(repositoryName))
	if err != nil {
		trace.AddEvent("siteadmin failed auth check")
		return errcode.HTTP(err), err
	}
	return 0, nil
}

func isSiteAdmin(ctx context.Context, userStore UserStore, logger sglog.Logger) bool {
	user, err := userStore.GetByCurrentAuthUser(ctx)
	if err != nil {
		if errcode.IsNotFound(err) || err == database.ErrNoCurrentUser {
			return false
		}

		logger.Error("codeintel.httpapi: failed to get up current user", sglog.Error(err))
		return false
	}

	return user != nil && user.SiteAdmin
}
