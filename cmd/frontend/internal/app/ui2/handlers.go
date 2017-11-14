package ui2

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/invite"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/jscontext"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth0"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/graphqlbackend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

var (
	injectedHTMLHeadTop    = env.Get("HTML_HEAD_TOP", "", "HTML to inject at the top of the <head> element on each page")
	injectedHTMLHeadBottom = env.Get("HTML_HEAD_BOTTOM", "", "HTML to inject at the bottom of the <head> element on each page")
	injectedHTMLBodyTop    = env.Get("HTML_BODY_TOP", "", "HTML to inject at the top of the <body> element on each page")
	injectedHTMLBodyBottom = env.Get("HTML_BODY_BOTTOM", "", "HTML to inject at the bottom of the <body> element on each page")
)

type InjectedHTML struct {
	HeadTop    template.HTML
	HeadBottom template.HTML
	BodyTop    template.HTML
	BodyBottom template.HTML
}

type Common struct {
	Injected InjectedHTML
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
// In the case of a repository that is cloning, a Common data structure is
// returned but it has an incomplete RevSpec.
func newCommon(w http.ResponseWriter, r *http.Request, title string, serveError func(w http.ResponseWriter, r *http.Request, err error, statusCode int)) (*Common, error) {
	common := &Common{
		Injected: InjectedHTML{
			HeadTop:    template.HTML(injectedHTMLHeadTop),
			HeadBottom: template.HTML(injectedHTMLHeadBottom),
			BodyTop:    template.HTML(injectedHTMLBodyTop),
			BodyBottom: template.HTML(injectedHTMLBodyBottom),
		},
		// InjectedHTMLHeadTop: template.HTML(injectedHTMLHeadTop),
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
			if e, ok := err.(vcs.RepoNotExistError); ok {
				if e.CloneInProgress {
					// Repo is cloning.
					return common, nil
				}
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

func serveBasicPageWithEmailVerification(title func(c *Common, r *http.Request) string) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		actor := actor.FromContext(r.Context())
		if actor != nil && actor.UID != "" {
			inviteToken := r.URL.Query().Get("token")
			// Verify the user email if they follow an invite link.
			if inviteToken != "" {
				// 🚨 SECURITY: verify that the token is valid before verifying email
				payload, err := invite.ParseToken(inviteToken)
				if err != nil {
					return err
				}
				// 🚨 SECURITY: verify the current actor's email iff it's the same as the email in the token
				// and the actor came from native authentication (i.e., not an external SSO provider)
				if payload.Email == actor.Email && actor.Provider == "" {
					err = auth0.VerifyEmail(r.Context(), actor.UID)
					if err != nil {
						return err
					}
				}
			}
		}

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

func serveEditorAuthWithEditorBetaRegistration(w http.ResponseWriter, r *http.Request) error {
	// Add editor beta tag for users who sign in or sign up from the editor.
	// This logic is executed when they are redirected to the editor-auth page
	// with the referrer=editor query string.
	user, err := localstore.Users.GetByCurrentAuthUser(r.Context())
	if err != nil {
		log15.Debug("no current auth user", "error", err)
	}
	if user != nil {
		referrer := r.URL.Query().Get("referrer")
		if referrer == "editor" {
			_, err := localstore.UserTags.CreateIfNotExists(r.Context(), user.ID, "editor-beta")
			if err != nil {
				return err
			}
		}
	}

	common, err := newCommon(w, r, "Authenticate editor - Sourcegraph", serveError)
	if err != nil {
		return err
	}
	if common == nil {
		return nil // request was handled
	}
	return renderTemplate(w, "app.html", common)
}

func serveHome(w http.ResponseWriter, r *http.Request) error {
	common, err := newCommon(w, r, "Sourcegraph", serveError)
	if err != nil {
		return err
	}
	if common == nil {
		return nil // request was handled
	}

	if (r.Host == "sourcegraph.com" || r.Host == "www.sourcegraph.com") && !actor.FromContext(r.Context()).IsAuthenticated() {
		// The user is not signed in and tried to access our main site at sourcegraph.com.
		// Redirect to about.sourcegraph.com so they see general info.
		u, err := url.Parse(aboutRedirectScheme + "://" + aboutRedirectHost)
		if err != nil {
			return err
		}
		http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
		return nil
	}

	// sourcegraph.com (not about) homepage. There is none, redirect them to /search.
	r.URL.Path = "/search"
	http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
	return nil
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
			// 	/search?q=ErrMethodMismatch&sq=repo:^github.com/gorilla/mux$
			//
			// It does not apply the file: filter because that was not the behavior of the
			// old blob URLs with a 'q' parameter either.
			r.URL.Path = "/search"
			q.Set("sq", "repo:^"+regexp.QuoteMeta(common.Repo.URI)+"$")
			r.URL.RawQuery = q.Encode()
			http.Redirect(w, r, r.URL.String(), http.StatusPermanentRedirect)
			return nil
		}
		return renderTemplate(w, "app.html", common)
	}
}

func serveComment(w http.ResponseWriter, r *http.Request) error {
	common, err := newCommon(w, r, "", serveError)
	if err != nil {
		return err
	}
	if common == nil {
		return nil // request was handled
	}

	// Locate the shared item.
	sharedItem, err := localstore.SharedItems.Get(r.Context(), mux.Vars(r)["ULID"])
	if err != nil {
		if _, ok := err.(localstore.ErrSharedItemNotFound); ok {
			// shared item does not exist.
			serveError(w, r, err, http.StatusNotFound)
			return nil
		}
		return errors.Wrap(err, "SharedItems.Get")
	}

	// Based on the shared item, determine the title and the thread ID.
	var (
		title    string
		threadID int32
	)
	switch {
	case sharedItem.ThreadID != nil:
		threadID = *sharedItem.ThreadID
		// TODO(slimsag): future: fetching all for thread just for first one's
		// title is not that optimal
		//
		// TODO(slimsag): future: If comment or thread was deleted, return 404
		// instead of 500.
		comments, err := localstore.Comments.GetAllForThread(r.Context(), *sharedItem.ThreadID)
		if err != nil {
			return errors.Wrap(err, "Comments.GetAllForThread")
		}
		if len(comments) > 0 {
			title = graphqlbackend.TitleFromContents(comments[0].Contents)
		}
	case sharedItem.CommentID != nil:
		// TODO(slimsag): future: If comment or thread was deleted, return 404
		// instead of 500.
		comment, err := localstore.Comments.GetByID(r.Context(), *sharedItem.CommentID)
		if err != nil {
			return errors.Wrap(err, "Comments.GetByID")
		}
		threadID = comment.ThreadID
		title = graphqlbackend.TitleFromContents(comment.Contents)
	}

	thread, err := localstore.Threads.Get(r.Context(), threadID)
	if err != nil {
		return errors.Wrap(err, "Threads.Get")
	}
	orgRepo, err := localstore.OrgRepos.GetByID(r.Context(), thread.OrgRepoID)
	if err != nil {
		return errors.Wrap(err, "OrgRepos.GetByID")
	}

	if !sharedItem.Public {
		actor := actor.FromContext(r.Context())
		if !actor.IsAuthenticated() {
			u := &url.URL{
				Path: "/sign-in",
			}
			q := u.Query()
			q.Set("returnTo", r.URL.String())
			u.RawQuery = q.Encode()
			http.Redirect(w, r, u.String(), http.StatusSeeOther)
			return nil
		}

		// 🚨 SECURITY: verify that the current user is in the org.
		_, err = localstore.OrgMembers.GetByOrgIDAndUserID(r.Context(), orgRepo.OrgID, actor.UID)
		if err != nil {
			// User is not in the org. We don't want to produce a 500, because we
			// want to render a nice error page on the frontend. But it's important
			// that we do not leak information about the shared item (e.g. through
			// the page title, see below).
			common.Title = "Sourcegraph"
			return renderTemplate(w, "app.html", common)
		}
	}

	if title != "" {
		common.Title = fmt.Sprintf("%s - Sourcegraph", title)
	} else {
		common.Title = fmt.Sprintf("%s - Sourcegraph", thread.File)
	}
	return renderTemplate(w, "app.html", common)
}
