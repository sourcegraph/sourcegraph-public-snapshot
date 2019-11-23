package ui

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assetsutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/jscontext"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/handlerutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/routevar"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

type InjectedHTML struct {
	HeadTop    template.HTML
	HeadBottom template.HTML
	BodyTop    template.HTML
	BodyBottom template.HTML
}

type Metadata struct {
	// Title is the title of the page for Twitter cards, OpenGraph, etc.
	// e.g. "Open in Sourcegraph"
	Title string

	// Description is the description of the page for Twitter cards, OpenGraph,
	// etc. e.g. "View this link in Sourcegraph Editor."
	Description string

	// ShowPreview controls whether or not OpenGraph/Twitter card/etc metadata is rendered.
	ShowPreview bool
}

type Common struct {
	Injected InjectedHTML
	Metadata *Metadata
	Context  jscontext.JSContext
	AssetURL string
	Title    string
	Error    *pageError

	WebpackDevServer bool // whether the Webpack dev server is running (WEBPACK_DEV_SERVER env var)

	// The fields below have zero values when not on a repo page.
	Repo         *types.Repo
	Rev          string // unresolved / user-specified revision (e.x.: "@master")
	api.CommitID        // resolved SHA1 revision
}

var webpackDevServer, _ = strconv.ParseBool(os.Getenv("WEBPACK_DEV_SERVER"))

// repoShortName trims the first path element of the given repo name if it has
// at least two path components.
func repoShortName(name api.RepoName) string {
	split := strings.Split(string(name), "/")
	if len(split) < 2 {
		return string(name)
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
// In the case of a repository that is cloning, a Common data structure is
// returned but it has an incomplete RevSpec.
func newCommon(w http.ResponseWriter, r *http.Request, title string, serveError func(w http.ResponseWriter, r *http.Request, err error, statusCode int)) (*Common, error) {
	common := &Common{
		Injected: InjectedHTML{
			HeadTop:    template.HTML(conf.Get().Critical.HtmlHeadTop),
			HeadBottom: template.HTML(conf.Get().Critical.HtmlHeadBottom),
			BodyTop:    template.HTML(conf.Get().Critical.HtmlBodyTop),
			BodyBottom: template.HTML(conf.Get().Critical.HtmlBodyBottom),
		},
		Context:  jscontext.NewJSContextFromRequest(r),
		AssetURL: assetsutil.URL("").String(),
		Title:    title,
		Metadata: &Metadata{
			Title:       conf.BrandName(),
			Description: "Sourcegraph is a web-based code search and navigation tool for dev teams. Search, navigate, and review code. Find answers.",
			ShowPreview: r.URL.Path == "/sign-in" && r.URL.RawQuery == "returnTo=%2F",
		},

		WebpackDevServer: webpackDevServer,
	}

	if _, ok := mux.Vars(r)["Repo"]; ok {
		// Common repo pages (blob, tree, etc).
		var err error
		common.Repo, common.CommitID, err = handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
		isRepoEmptyError := routevar.ToRepoRev(mux.Vars(r)).Rev == "" && gitserver.IsRevisionNotFound(errors.Cause(err)) // should reply with HTTP 200
		if err != nil && !isRepoEmptyError {
			if e, ok := err.(*handlerutil.URLMovedError); ok {
				// The repository has been renamed, e.g. "github.com/docker/docker"
				// was renamed to "github.com/moby/moby" -> redirect the user now.
				err = handlerutil.RedirectToNewRepoName(w, r, e.NewRepo)
				if err != nil {
					return nil, errors.Wrap(err, "when sending renamed repository redirect response")
				}

				return nil, nil
			}
			if e, ok := err.(backend.ErrRepoSeeOther); ok {
				// Repo does not exist here, redirect to the recommended location.
				u, err := url.Parse(e.RedirectURL)
				if err != nil {
					return nil, err
				}
				u.Path, u.RawQuery = r.URL.Path, r.URL.RawQuery
				http.Redirect(w, r, u.String(), http.StatusSeeOther)
				return nil, nil
			}
			if gitserver.IsRevisionNotFound(errors.Cause(err)) {
				// Revision does not exist.
				serveError(w, r, err, http.StatusNotFound)
				return nil, nil
			}
			if _, ok := errors.Cause(err).(*gitserver.RepoNotCloneableErr); ok {
				if errcode.IsNotFound(err) {
					// Repository is not found.
					serveError(w, r, err, http.StatusNotFound)
					return nil, nil
				}

				// Repository is not clonable.
				dangerouslyServeError(w, r, errors.New("repository could not be cloned"), http.StatusInternalServerError)
				return nil, nil
			}
			if vcs.IsRepoNotExist(err) {
				if vcs.IsCloneInProgress(err) {
					// Repo is cloning.
					return common, nil
				}
				// Repo does not exist.
				serveError(w, r, err, http.StatusNotFound)
				return nil, nil
			}
			if errcode.IsNotFound(err) || errors.Cause(err) == repoupdater.ErrNotFound {
				// Repo does not exist.
				serveError(w, r, err, http.StatusNotFound)
				return nil, nil
			}
			if errors.Cause(err) == repoupdater.ErrUnauthorized {
				// Not authorized to access repository.
				serveError(w, r, err, http.StatusUnauthorized)
				return nil, nil
			}
			return nil, err
		}
		if common.Repo.Name == "github.com/sourcegraphtest/Always500Test" {
			return nil, errors.New("error caused by Always500Test repo name")
		}
		common.Rev = mux.Vars(r)["Rev"]
		// Update gitserver contents for a repo whenever it is visited.
		go func() {
			ctx := context.Background()
			gitserverRepo, err := backend.GitRepo(ctx, common.Repo)
			if err != nil {
				log15.Error("backend.GitRepo", "error", err)
				return
			}

			_, err = repoupdater.DefaultClient.EnqueueRepoUpdate(ctx, gitserverRepo)
			if err != nil {
				log15.Error("EnqueueRepoUpdate", "error", err)
			}
		}()
	}
	return common, nil
}

type handlerFunc func(w http.ResponseWriter, r *http.Request) error

func serveBrandedPageString(titles ...string) handlerFunc {
	return serveBasicPage(func(c *Common, r *http.Request) string {
		return brandNameSubtitle(titles...)
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
	common, err := newCommon(w, r, conf.BrandName(), serveError)
	if err != nil {
		return err
	}
	if common == nil {
		return nil // request was handled
	}

	if envvar.SourcegraphDotComMode() && !actor.FromContext(r.Context()).IsAuthenticated() {
		// The user is not signed in and tried to access Sourcegraph.com.  Redirect to
		// about.sourcegraph.com so they see general info page.
		http.Redirect(w, r, (&url.URL{Scheme: aboutRedirectScheme, Host: aboutRedirectHost}).String(), http.StatusTemporaryRedirect)
		return nil
	}
	// On non-Sourcegraph.com instances, there is no separate homepage, so redirect to /search.
	r.URL.Path = "/search"
	http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
	return nil
}

func serveSignIn(w http.ResponseWriter, r *http.Request) error {
	common, err := newCommon(w, r, "", serveError)
	if err != nil {
		return err
	}
	if common == nil {
		return nil // request was handled
	}
	common.Title = brandNameSubtitle("Sign in")

	// If we are being redirected to another page after sign in, it means the
	// user attempted to access something without authorization. Reflect this
	// in the status code. This is useful when users curl / code which
	// interacts with the Sourcegraph endpoints. Specifically this is a common
	// issue facing extension developers interacting with the raw API.
	if r.URL.Query().Get("returnTo") != "" {
		w.WriteHeader(http.StatusUnauthorized)
	}

	return renderTemplate(w, "app.html", common)
}

func serveRepoOrBlob(routeName string, title func(c *Common, r *http.Request) string) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		common, err := newCommon(w, r, "", serveError)
		if err != nil {
			return err
		}
		if common == nil {
			return nil // request was handled
		}
		common.Title = title(common, r)

		q := r.URL.Query()
		_, isNewQueryUX := q["sq"] // sq URL param is only set by new query UX in SearchNavbarItem.tsx
		if search := q.Get("q"); search != "" && !isNewQueryUX {
			// Redirect old search URLs:
			//
			// 	/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7?q=ErrMethodMismatch&utm_source=chrome-extension
			// 	/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go?q=NewRouter
			//
			// To new ones:
			//
			// 	/search?q=repo:^github.com/gorilla/mux$+ErrMethodMismatch
			//
			// It does not apply the file: filter because that was not the behavior of the
			// old blob URLs with a 'q' parameter either.
			r.URL.Path = "/search"
			q.Set("sq", "repo:^"+regexp.QuoteMeta(string(common.Repo.Name))+"$")
			r.URL.RawQuery = q.Encode()
			http.Redirect(w, r, r.URL.String(), http.StatusPermanentRedirect)
			return nil
		}
		return renderTemplate(w, "app.html", common)
	}
}

// searchBadgeHandler serves the search readme badges from the search-badger service
// https://github.com/sourcegraph/search-badger
var searchBadgeHandler = &httputil.ReverseProxy{
	Director: func(r *http.Request) {
		r.URL.Scheme = "http"
		r.URL.Host = "search-badger"
		r.URL.Path = "/"
	},
	ErrorLog: log.New(env.DebugOut, "search-badger proxy: ", log.LstdFlags),
}
