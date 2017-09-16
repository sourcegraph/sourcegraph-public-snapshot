package ui2

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/jscontext"
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
