package ui2

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/invite"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/jscontext"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth0"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/eventlogger"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/graphqlbackend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

var (
	injectedHTMLHeadTop    = conf.Get().HTMLHeadTop
	injectedHTMLHeadBottom = conf.Get().HTMLHeadTop
	injectedHTMLBodyTop    = conf.Get().HTMLBodyTop
	injectedHTMLBodyBottom = conf.Get().HTMLBodyTop
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
}

type Common struct {
	Injected InjectedHTML
	Metadata *Metadata
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
			if e, ok := err.(localstore.ErrRepoSeeOther); ok {
				// Repo does not exist here, redirect to the reccomended location.
				http.Redirect(w, r, e.RedirectURL, http.StatusSeeOther)
				return nil, nil
			}
			if legacyerr.ErrCode(err) == legacyerr.NotFound {
				// Repo does not exist.
				serveError(w, r, err, http.StatusNotFound)
				return nil, nil
			}
			if errors.Cause(err) == vcs.ErrRevisionNotFound {
				// Revision does not exist.
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
	// Add editor beta tag for users who sign in or sign up from the editor, and to
	// all of their orgs.
	// This logic is executed when they are redirected to the editor-auth page
	// with the referrer=editor query string.
	user, err := localstore.Users.GetByCurrentAuthUser(r.Context())
	if err != nil {
		log15.Debug("no current auth user", "error", err)
	}
	if user != nil {
		referrer := r.URL.Query().Get("referrer")
		if referrer == "editor" {
			const editorBetaTag = "editor-beta"
			// Add tag to user.
			_, err := localstore.UserTags.CreateIfNotExists(r.Context(), user.ID, editorBetaTag)
			if err != nil {
				return err
			}

			// Add tag to all orgs.
			orgs, err := localstore.Orgs.GetByUserID(r.Context(), user.Auth0ID)
			if err != nil {
				return err
			}
			for _, org := range orgs {
				if _, err := localstore.OrgTags.CreateIfNotExists(r.Context(), org.ID, editorBetaTag); err != nil {
					return err
				}
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

	// At this point, it's a public ('secret URL') shared item.
	//
	// Generate metadata for the page.
	snippet := false
	if title == "" {
		snippet = true
		title = fmt.Sprintf("%s (Snippet)", thread.File)
	}

	var rev string
	if thread.Branch != nil {
		rev = "@" + *thread.Branch
	}
	var description string
	if snippet {
		description = fmt.Sprintf("Snippet from %s:%d (%s%s) ", thread.File, thread.StartLine, orgRepo.CanonicalRemoteID, rev)
	} else {
		description = fmt.Sprintf("Discussion at %s:%d (%s%s) ", thread.File, thread.StartLine, orgRepo.CanonicalRemoteID, rev)
	}

	metadata := &Metadata{}
	ua := r.Header.Get("User-Agent")
	service := ""
	switch {
	case strings.Contains(ua, "Slackbot"):
		service = "Slack"
		// Note the HTML escape here is not for security -- but rather for
		// Slack's quite strange behavior which requires double escaping to get
		// proper rendering of e.g. &lt; and &gt; brackets.
		//
		// To test this unfurl a link to a comment with the text:
		//
		// 	"<button> below the `<form` that's right!"
		//
		metadata.Title = strings.Replace(title, "<", "&lt;", -1)
		metadata.Title = strings.Replace(metadata.Title, ">", "&gt;", -1)
		metadata.Description = description

	case strings.Contains(ua, "Twitterbot"):
		service = "Twitter"
		// Try it here: https://cards-dev.twitter.com/validator
		fallthrough

	case strings.Contains(ua, "facebook"):
		service = "Facebook"
		// Try it here: https://developers.facebook.com/tools/debug/sharing/
		//
		// Note: ngrok often blocks Facebook's crawlers for some reason (https://developers.facebook.com/bugs/824028317765435/).
		// Try localtunnel instead: https://localtunnel.github.io/www/#quickstart
		fallthrough

	default:
		metadata.Title = title
		metadata.Description = description
	}
	common.Metadata = metadata

	if service != "" {
		// Link unfurled by some service in specific (i.e. not just some user
		// visiting this link in their browser)
		err := eventlogger.EventLogger.LogEvent(nil, "CommentUnfurled", map[string]string{"unfurl_service": service})
		if err != nil {
			log15.Warn("failed to log link unfurl event", "error", err)
		}
	}

	common.Title = fmt.Sprintf("%s - Sourcegraph", title)
	return renderTemplate(w, "app.html", common)
}

func serveOpen(w http.ResponseWriter, r *http.Request) error {
	common, err := newCommon(w, r, "Open in Sourcegraph", serveError)
	if err != nil {
		return err
	}
	if common == nil {
		return nil // request was handled
	}

	q := r.URL.Query()
	repo := q.Get("repo") // always present; e.g. "ssh://git@github.com/sourcegraph/sourcegraph.git"

	// Guess that the repo name is the last repo clone URL path component.
	repoSplit := strings.Split(repo, "/")
	repoName := strings.TrimSuffix(repoSplit[len(repoSplit)-1], ".git")
	repoName = strings.Title(repoName)

	// Generally only present for diff links:
	// e.g. https://sourcegraph.com/open?repo=git%40github.com%3Asourcegraph%2Fsourcegraph.git&revision=vo%2Flight_theme&baseRevision=master&vcs=git
	revision := q.Get("revision")         // e.g. "sg/featurebranch"
	baseRevision := q.Get("baseRevision") // e.g. "master"

	// Generally only present for links to exact file/line number:
	// e.g. https://sourcegraph.com/open?repo=ssh%3A%2F%2Fgit%40github.com%2Fsourcegraph%2Fsourcegraph.git&vcs=git&path=web%2Fsrc%2Fcomments%2FCommentsPage.tsx&selection=177&thread=1300
	pathStr := q.Get("path")           // e.g. "web/src/comments/CommentsPage.tsx"
	lineNumber := q.Get("selection")   // e.g. "177"
	_, fileName := path.Split(pathStr) // "CommentsPage.tsx"

	// Generate metadata for the page.
	metadata := &Metadata{}
	ua := r.Header.Get("User-Agent")
	service := ""
	switch {
	case strings.Contains(ua, "Twitterbot"):
		service = "Twitter"
		// Try it here: https://cards-dev.twitter.com/validator
		if baseRevision == "" {
			metadata.Title = fmt.Sprintf("%s:%s", ellipsisPath(pathStr, 2), lineNumber)
			metadata.Description = fmt.Sprintf("Open %s:%s (%s) in Sourcegraph Editor", fileName, lineNumber, repoName)
		} else {
			metadata.Title = fmt.Sprintf("%s...%s", baseRevision, revision)
			metadata.Description = fmt.Sprintf("Open Git diff %s...%s (%s) in Sourcegraph Editor", baseRevision, revision, repoName)
		}

	case strings.Contains(ua, "Slackbot"):
		service = "Slack"
		fallthrough

	case strings.Contains(ua, "facebook"):
		service = "Facebook"
		// Try it here: https://developers.facebook.com/tools/debug/sharing/
		//
		// Note: ngrok often blocks Facebook's crawlers for some reason (https://developers.facebook.com/bugs/824028317765435/).
		// Try localtunnel instead: https://localtunnel.github.io/www/#quickstart
		fallthrough

	default:
		if baseRevision == "" {
			metadata.Title = fmt.Sprintf("%s:%s - %s", pathStr, lineNumber, repoName)
			metadata.Description = fmt.Sprintf("Open %s:%s (%s) in Sourcegraph Editor", fileName, lineNumber, repoName)
		} else {
			metadata.Title = fmt.Sprintf("%s...%s - %s", baseRevision, revision, repoName)
			metadata.Description = fmt.Sprintf("Open Git diff %s...%s (%s) in Sourcegraph Editor", baseRevision, revision, repoName)
		}
	}
	common.Metadata = metadata

	if service != "" {
		// Link unfurled by some service in specific (i.e. not just some user
		// visiting this link in their browser)
		err := eventlogger.EventLogger.LogEvent(nil, "DeepLinkUnfurled", map[string]string{"unfurl_service": service})
		if err != nil {
			log15.Warn("failed to log link unfurl event", "error", err)
		}
	}

	return renderTemplate(w, "app.html", common)
}

// ellipsisPath returns the given path with at max 2 path components from the
// end, and an ellipsis (…) at the front when neccessary.
func ellipsisPath(pathStr string, n int) string {
	split := strings.Split(pathStr, "/")
	if len(split) < n {
		return pathStr
	}
	return path.Join(append([]string{"…"}, split[len(split)-n:]...)...)
}
