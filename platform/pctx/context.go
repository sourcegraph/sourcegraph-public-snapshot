package pctx

import (
	"fmt"
	"net/http"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"github.com/justinas/nosurf"
	"github.com/sourcegraph/mux"
	"golang.org/x/net/context"
	approuter "src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/conf"
)

type contextKey int

// Context item keys for data that platform applications have access to.
const (
	csrfTokenKey = iota
	repoRevSpecKey
	baseURIKey
)

// CSRFToken is the token that should be used in forms and other
// actions to prevent cross-site request forgeries.
func CSRFToken(ctx context.Context) string {
	return ctx.Value(csrfTokenKey).(string)
}

// RepoRevSpec is the repository revision spec. This is encoded in the
// HTTP request URL but the URL is trimmed before the request is
// forwarded to platform apps, so it must be passed in the context.
func RepoRevSpec(ctx context.Context) (sourcegraph.RepoRevSpec, bool) {
	rrs, exists := ctx.Value(repoRevSpecKey).(sourcegraph.RepoRevSpec)
	return rrs, exists
}

// BaseURI is the base URI of the repo frame.
func BaseURI(ctx context.Context) string {
	return ctx.Value(baseURIKey).(string)
}

// WithRepoFrameInfo computes the context to be passed to a repository
// frame application handler.
func WithRepoFrameInfo(ctx context.Context, r *http.Request) (context.Context, error) {
	repoRevSpec, err := sourcegraph.UnmarshalRepoRevSpec(mux.Vars(r))
	if err != nil {
		return nil, err
	}
	baseURI, err := repoFrameBaseURI(ctx, r)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, csrfTokenKey, nosurf.Token(r))
	ctx = context.WithValue(ctx, repoRevSpecKey, repoRevSpec)
	ctx = context.WithValue(ctx, baseURIKey, baseURI)
	return ctx, nil
}

// repoFrameBaseURI computes the root URI of an application repository frame.
// Repository frames often will contain their own URL subrouters.
func repoFrameBaseURI(ctx context.Context, r *http.Request) (string, error) {
	vars := mux.Vars(r)

	urlVars := []string{
		"Repo", vars["Repo"],
		"App", vars["App"],
		"AppPath", "",
	}
	if resolvedRev, exists := vars["ResolvedRev"]; exists {
		urlVars = append(urlVars, "ResolvedRev", resolvedRev)
	} else {
		urlVars = append(urlVars, "Rev", vars["Rev"], "CommitID", vars["CommitID"])
	}

	baseURI, err := approuter.New(nil).Get(approuter.RepoAppFrame).URLPath(urlVars...)
	if err != nil {
		return "", fmt.Errorf("could not produce base URL for app request url %s: %s", r.URL, err)
	}
	return conf.AppURL(ctx).ResolveReference(baseURI).String(), nil
}
