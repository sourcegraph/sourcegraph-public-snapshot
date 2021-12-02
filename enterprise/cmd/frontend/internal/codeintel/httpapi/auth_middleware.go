package httpapi

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

type AuthValidator func(context.Context, url.Values, string) (int, error)
type AuthValidatorMap = map[string]AuthValidator

var DefaultValidatorByCodeHost = AuthValidatorMap{
	"github.com": enforceAuthViaGitHub,
}

// authMiddleware wraps the given upload handler with an authorization check. On each initial upload
// request, the target repository is checked against the supplied auth validators. The matching validator
// is invoked, which coordinates with a remote code host's permissions API to determine if the current
// request contains sufficient evidence of authorship for the target repository.
//
// When LSIF auth is not enforced on the instance, this middleware no-ops.
func authMiddleware(next http.Handler, db dbutil.DB, authValidators AuthValidatorMap) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Skip auth check if it's not enabled in the instance's site configuration, if this
		// user is a site admin (who can upload LSIF to any repository on the instance), or
		// if the request a subsequent request of a multi-part upload.
		if !conf.Get().LsifEnforceAuth || isSiteAdmin(ctx, db) || hasQuery(r, "uploadId") {
			next.ServeHTTP(w, r)
			return
		}

		query := r.URL.Query()
		repositoryName := getQuery(r, "repository")

		for codeHost, validator := range authValidators {
			if !strings.HasPrefix(repositoryName, codeHost) {
				continue
			}

			if statusCode, err := validator(ctx, query, repositoryName); err != nil {
				if statusCode >= 500 {
					log15.Error("codeintel.httpapi: failed to authorize request", "error", err)
				}

				http.Error(w, err.Error(), statusCode)
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		log15.Warn("codeintel.httpapi: verification not supported for code host", "repositoryName", repositoryName)
		http.Error(w, "verification not supported for code host - see https://github.com/sourcegraph/sourcegraph/issues/4967", http.StatusUnprocessableEntity)
	})
}

func isSiteAdmin(ctx context.Context, db dbutil.DB) bool {
	user, err := database.Users(db).GetByCurrentAuthUser(ctx)
	if err != nil {
		if errcode.IsNotFound(err) || err == database.ErrNoCurrentUser {
			return false
		}

		log15.Error("codeintel.httpapi: failed to get up current user", "error", err)
		return false
	}

	return user != nil && user.SiteAdmin
}
