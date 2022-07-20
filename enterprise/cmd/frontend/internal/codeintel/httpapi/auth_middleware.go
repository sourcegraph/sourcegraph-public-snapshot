package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
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
)

var DefaultValidatorByCodeHost = AuthValidatorMap{
	"github.com": enforceAuthViaGitHub,
	"gitlab.com": enforceAuthViaGitLab,
}

var errVerificationNotSupported = errors.New(strings.Join([]string{
	"verification is supported for the following code hosts: github.com, gitlab.com",
	"please request support for additional code host verification at https://github.com/sourcegraph/sourcegraph/issues/4967",
}, " - "))

// authMiddleware wraps the given upload handler with an authorization check. On each initial upload
// request, the target repository is checked against the supplied auth validators. The matching validator
// is invoked, which coordinates with a remote code host's permissions API to determine if the current
// request contains sufficient evidence of authorship for the target repository.
//
// When LSIF auth is not enforced on the instance, this middleware no-ops.
func authMiddleware(next http.Handler, db database.DB, authValidators AuthValidatorMap, operation *observation.Operation) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusCode, err := func() (_ int, err error) {
			ctx, trace, endObservation := operation.With(r.Context(), &err, observation.Args{})
			defer endObservation(1, observation.Args{})

			repositoryName := getQuery(r, "repository")

			// Skip auth check if it's not enabled in the instance's site configuration, if this
			// user is a site admin (who can upload LSIF to any repository on the instance), or
			// if the request a subsequent request of a multi-part upload.
			if !conf.Get().LsifEnforceAuth || isSiteAdmin(ctx, db) || hasWriteAccess(ctx, repositoryName, db) || hasQuery(r, "uploadId") {
				trace.Log(log.Event("bypassing code host auth check"))
				return 0, nil
			}

			query := r.URL.Query()

			for codeHost, validator := range authValidators {
				if !strings.HasPrefix(repositoryName, codeHost) {
					continue
				}
				trace.Log(log.String("codeHost", codeHost))

				return validator(ctx, query, repositoryName)
			}

			return http.StatusUnprocessableEntity, errVerificationNotSupported
		}()
		if err != nil {
			if statusCode >= 500 {
				log15.Error("codeintel.httpapi: failed to authorize request", "error", err)
			}

			http.Error(w, fmt.Sprintf("failed to authorize request: %s", err.Error()), statusCode)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func hasWriteAccess(ctx context.Context, repoName string, db database.DB) bool {
	user, err := database.Users(db).GetByCurrentAuthUser(ctx)
	if err != nil {
		if errcode.IsNotFound(err) || err == database.ErrNoCurrentUser {
			return false
		}
		log15.Error("codeintel.httpapi: failed to get current user", "error", err)
		return false
	}

	repo, err := db.Repos().GetByName(ctx, api.RepoName(repoName))
	if err != nil {
		log15.Error("codeintel.httpapi: failed to get repo from name", "error", err)
		return false
	}

	repos, err := db.Authz().AuthorizedRepos(ctx, &database.AuthorizedReposArgs{
		Repos:  []*types.Repo{repo},
		UserID: user.ID,
		Perm:   authz.Write,
		Type:   authz.PermRepos,
	})
	if err != nil {
		log15.Error("codeintel.httpapi: failed to get authorized repos", "error", err, "user", user.ID, "repo", repo.ID)
		return false
	}

	return len(repos) > 0
}

func isSiteAdmin(ctx context.Context, db database.DB) bool {
	user, err := db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		if errcode.IsNotFound(err) || err == database.ErrNoCurrentUser {
			return false
		}

		log15.Error("codeintel.httpapi: failed to get current user", "error", err)
		return false
	}

	return user != nil && user.SiteAdmin
}
