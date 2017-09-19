package ui2

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/jscontext"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

type Common struct {
	Context  jscontext.JSContext
	AssetURL string
	Title    string
	Error    *pageError

	// The fields below have zero values when not on a repo page.
	Repo    *sourcegraph.Repo
	Rev     string                  // unresolved / user-specified revision (e.x.: "@master")
	RevSpec sourcegraph.RepoRevSpec // resolved SHA1 revision
}

// repoShortName trims the first path element of the given repo uri if it has
// at least two path components.
func repoShortName(uri string) string {
	split := strings.Split(uri, "/")
	if len(split) < 2 {
		return uri
	}
	return strings.Join(split[1:], "/")
}

// newCommon builds a *Common data structure, returning an error if one occurs.
//
// In the event of the repository having been renamed, the request is handled
// by newCommon and nil, nil is returned. Basic usage looks like:
//
// 	common, err := newCommon(w, r, serveError)
// 	if err != nil {
// 		return err
// 	}
// 	if common == nil {
// 		return nil // request was handled
// 	}
//
func newCommon(w http.ResponseWriter, r *http.Request, title string, serveError func(w http.ResponseWriter, r *http.Request, err error, statusCode int)) (*Common, error) {
	common := &Common{
		Context:  jscontext.NewJSContextFromRequest(r),
		AssetURL: assets.URL("/").String(),
		Title:    title,
	}

	if _, ok := mux.Vars(r)["Repo"]; ok {
		// Common repo pages (blob, tree, etc).
		var err error
		common.Repo, common.RevSpec, err = handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
		if err != nil {
			if e, ok := err.(*handlerutil.URLMovedError); ok {
				// The repository has been renamed, e.g. "github.com/docker/docker"
				// was renamed to "github.com/moby/moby" -> redirect the user now.
				http.Redirect(w, r, "/"+e.NewURL, http.StatusMovedPermanently)
				return nil, nil
			}
			if legacyerr.ErrCode(err) == legacyerr.NotFound {
				// Repo does not exist.
				serveError(w, r, err, http.StatusNotFound)
				return nil, nil
			}
			return nil, err
		}
		if common.Repo.Private {
			serveError(w, r, errors.New("accessing private repositories is forbidden"), http.StatusNotFound)
			return nil, nil
		}
		if common.Repo.URI == "github.com/sourcegraphtest/Always500Test" {
			return nil, errors.New("error caused by Always500Test repo URI")
		}
		common.Rev = mux.Vars(r)["Rev"]
	}
	return common, nil
}

type handlerFunc func(w http.ResponseWriter, r *http.Request) error

func serveBasicPageString(title string) handlerFunc {
	return serveBasicPage(func(c *Common, r *http.Request) string {
		return title
	})
}

func serveBasicPage(title func(c *Common, r *http.Request) string) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		common, err := newCommon(w, r, "", serveError)
		if err != nil {
			return err
		}
		if common == nil {
			return nil // request was handled
		}
		common.Title = title(common, r)
		return renderTemplate(w, "app.html", common)
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) error {
	common, err := newCommon(w, r, "Sourcegraph", serveError)
	if err != nil {
		return err
	}
	if common == nil {
		return nil // request was handled
	}

	if !envvar.DeploymentOnPrem() && !actor.FromContext(r.Context()).IsAuthenticated() {
		// The user is not signed in and we are not on-prem so we are going to
		// redirect to about.sourcegraph.com.
		u, err := url.Parse(aboutRedirectScheme + "://" + aboutRedirectHost)
		if err != nil {
			return err
		}
		q := url.Values{}
		if r.Host != "sourcegraph.com" {
			// This allows about.sourcegraph.com to properly redirect back to
			// dev or staging environment after sign in.
			q.Set("host", r.Host)
		}
		u.RawQuery = q.Encode()
		http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
		return nil
	}

	// sourcegraph.com (not about) homepage. There is none, redirect them to /search.
	r.URL.Path = "/search"
	http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
	return nil
}

func serveRepo(w http.ResponseWriter, r *http.Request) error {
	common, err := newCommon(w, r, "", serveError)
	if err != nil {
		return err
	}
	if common == nil {
		return nil // request was handled
	}
	// e.g. "gorilla/mux - Sourcegraph"
	common.Title = fmt.Sprintf("%s - Sourcegraph", repoShortName(common.Repo.URI))

	q := r.URL.Query()
	if search := q.Get("q"); search != "" {
		// Redirect old search URLs:
		//
		// 	/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7?q=ErrMethodMismatch&utm_source=chrome-extension
		//
		// To new ones:
		//
		// 	/search?q=route&repo=github.com/gorilla/mux&repo=github.com/kubernetes/kubernetes&matchCase=false&matchWord=false
		//
		r.URL.Path = "/search"
		q.Set("repo", common.Repo.URI)
		q.Set("matchCase", "false")
		q.Set("matchWord", "false")
		r.URL.RawQuery = q.Encode()
		http.Redirect(w, r, r.URL.String(), http.StatusPermanentRedirect)
		return nil
	}
	return renderTemplate(w, "app.html", common)
}
