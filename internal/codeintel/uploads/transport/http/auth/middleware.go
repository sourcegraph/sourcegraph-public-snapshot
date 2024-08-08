package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	sglog "github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type (
	AuthValidator    func(context.Context, url.Values, string) (int, error)
	AuthValidatorMap = map[string]AuthValidator
	RepoStore        interface {
		GetByName(context.Context, api.RepoName) (*types.Repo, error)
	}
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
	repoStore RepoStore,
	authValidators AuthValidatorMap,
	operation *observation.Operation,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusCode, err := func() (_ int, err error) {
			ctx, trace, endObservation := operation.With(r.Context(), &err, observation.Args{})
			defer endObservation(1, observation.Args{})

			user, err := isLoggedIn(ctx, userStore, trace)
			if err != nil {
				return http.StatusUnauthorized, err
			}

			query := r.URL.Query()
			repositoryName := getQuery(r, "repository")

			if user.isSiteAdmin() {
				// If `AuthzEnforceForSiteAdmins` is set we should still check repo permissions for site admins
				if conf.Get().AuthzEnforceForSiteAdmins {
					statusCode, err := user.canAccessRepository(ctx, repoStore, repositoryName, trace)
					if err != nil {
						return statusCode, err
					}
				}
				return 0, nil
			}

			// Non site-admin users can upload indices if auth is not enabled in the instance's site configuration
			if !conf.Get().LsifEnforceAuth {
				statusCode, err := user.canAccessRepository(ctx, repoStore, repositoryName, trace)
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

type loggedInUserDoNotCreateThisTypeDirectly struct {
	user *types.User
}

func isLoggedIn(ctx context.Context, userStore UserStore, trace observation.TraceLogger) (out loggedInUserDoNotCreateThisTypeDirectly, err error) {
	user, err := userStore.GetByCurrentAuthUser(ctx)
	if err == nil {
		out.user = user
	}
	if err != nil && !(errcode.IsNotFound(err) || err == database.ErrNoCurrentUser) {
		trace.Error("codeintel.httpapi: failed to find current user", sglog.Error(err))
	}
	return out, err
}

func (u *loggedInUserDoNotCreateThisTypeDirectly) canAccessRepository(
	ctx context.Context,
	repoStore RepoStore,
	repositoryName string,
	trace observation.TraceLogger,
) (int, error) {
	if u.user == nil {
		panic("This method should not be called if the user was not known")
	}
	if _, err := repoStore.GetByName(ctx, api.RepoName(repositoryName)); err != nil {
		trace.AddEvent("siteadmin failed auth check")
		return errcode.HTTP(err), err
	}
	return 0, nil
}

func (u *loggedInUserDoNotCreateThisTypeDirectly) isSiteAdmin() bool {
	if u.user == nil {
		panic("This method should not be called if the user was not known")
	}
	return u.user.SiteAdmin
}
