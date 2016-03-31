// Package router contains the URL router for the frontend app.
package router

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/mux"
	gitrouter "sourcegraph.com/sourcegraph/sourcegraph/app/internal/gitserver/router"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/spec"
)

const (
	Builds = "builds"

	Home = "home"

	RegisterClient = "register-client"

	RobotsTxt = "robots-txt"
	Favicon   = "favicon"

	SitemapIndex = "sitemap-index"
	RepoSitemap  = "repo.sitemap"

	User                = "person"
	UserSettingsProfile = "person.settings.profile"

	Repo             = "repo"
	RepoBuilds       = "repo.builds"
	RepoBuild        = "repo.build"
	RepoBuildUpdate  = "repo.build.update"
	RepoBuildTaskLog = "repo.build.task.log"
	RepoBuildsCreate = "repo.builds.create"
	RepoTree         = "repo.tree"

	RepoRevCommits = "repo.rev.commits"
	RepoCommit     = "repo.commit"
	RepoTags       = "repo.tags"
	RepoBranches   = "repo.branches"

	LogIn          = "log-in"
	LogOut         = "log-out"
	SignUp         = "sign-up"
	ForgotPassword = "forgot-password"
	ResetPassword  = "reset-password"

	OAuth2ServerAuthorize = "oauth-provider.authorize"
	OAuth2ServerToken     = "oauth-provider.token"

	GitHubOAuth2Initiate = "github-oauth2.initiate"
	GitHubOAuth2Receive  = "github-oauth2.receive"

	Def     = "def"
	DefRefs = "def.refs"

	UserContent = "usercontent"

	OldDefRedirect = "old-def-redirect"

	OldTreeRedirect = "old-tree-redirect"

	// Platform routes
	RepoAppFrame = "repo.appframe"
)

// Router is an app URL router.
type Router struct{ mux.Router }

// New creates a new app router with route URL pattern definitions but
// no handlers attached to the routes.
//
// It is in a separate package from app so that other packages may use it to
// generate URLs without resulting in Go import cycles (and so we can release
// the router as open-source to support our client library).
func New(base *mux.Router) *Router {
	if base == nil {
		base = mux.NewRouter()
	}

	base.StrictSlash(true)

	base.Path("/").Methods("GET").Name(Home)
	base.Path("/register-client").Methods("GET", "POST").Name(RegisterClient)

	base.Path("/.builds").Methods("GET").Name(Builds)

	base.Path("/login").Methods("GET", "POST").Name(LogIn)
	base.Path("/join").Methods("GET", "POST").Name(SignUp)
	base.Path("/logout").Methods("POST").Name(LogOut)
	base.Path("/forgot").Methods("GET", "POST").Name(ForgotPassword)
	base.Path("/reset").Methods("GET", "POST").Name(ResetPassword)

	base.Path("/login/oauth/authorize").Methods("GET").Name(OAuth2ServerAuthorize)
	base.Path("/login/oauth/token").Methods("POST").Name(OAuth2ServerToken)

	base.Path("/robots.txt").Methods("GET").Name(RobotsTxt)
	base.Path("/favicon.ico").Methods("GET").Name(Favicon)

	base.Path("/sitemap.xml").Methods("GET").Name(SitemapIndex)

	base.Path("/github-oauth/initiate").Methods("GET").Name(GitHubOAuth2Initiate)
	base.Path("/github-oauth/receive").Methods("GET", "POST").Name(GitHubOAuth2Receive)

	base.Path("/usercontent/{Name}").Methods("GET").Name(UserContent)

	// User routes begin with tilde (~).
	userPath := `/~` + routevar.User
	user := base.PathPrefix(userPath).Subrouter()
	user.Path("/.settings/profile").Methods("GET", "POST").Name(UserSettingsProfile)

	addOldDefRedirectRoute(&Router{*base}, base)
	addOldTreeRedirectRoute(&Router{*base}, base)

	// attach git transport endpoints
	gitrouter.New(base)

	repoPath := `/` + routevar.Repo
	base.Path(repoPath + routevar.RepoRevSuffix).Methods("GET").Name(Repo)
	repo := base.PathPrefix(repoPath + "/" + spec.RepoPathDelim + "/").Subrouter()
	repoRev := base.PathPrefix(repoPath + routevar.RepoRevSuffix + "/" + spec.RepoPathDelim + "/").Subrouter()

	defPath := "/def/" + routevar.Def
	def := repoRev.PathPrefix(defPath + "/-/").Subrouter()
	def.Path("/refs").Methods("GET").Name(DefRefs)
	repoRev.Path(defPath).Methods("GET").Name(Def)

	// See router_util/tree_route.go for an explanation of how we match tree
	// entry routes.
	repoTreePath := "/tree{Path:.*}"
	repoRev.Path(repoTreePath + "/.sourcebox.{Format}").HandlerFunc(gone)
	repoRev.Path(repoTreePath).Methods("GET").Name(RepoTree)

	repoRev.Path("/commits").Methods("GET").Name(RepoRevCommits)

	repoRev.Path("/commit").Methods("GET").Name(RepoCommit)
	repo.Path("/branches").Methods("GET").Name(RepoBranches)
	repo.Path("/tags").Methods("GET").Name(RepoTags)
	repo.Path("/sitemap.xml").Methods("GET").Name(RepoSitemap)

	repo.Path("/builds").Methods("GET").Name(RepoBuilds)
	repo.Path("/builds").Methods("POST").Name(RepoBuildsCreate)
	repoBuildPath := `/builds/{Build:\d+}`
	repo.Path(repoBuildPath).Methods("GET").Name(RepoBuild)
	repo.Path(repoBuildPath).Methods("POST").Name(RepoBuildUpdate)
	repoBuild := repo.PathPrefix(repoBuildPath).Subrouter()
	repoBuild.Path(`/tasks/{Task:\d+}/log`).Methods("GET").Name(RepoBuildTaskLog)

	// This route should be AFTER all other repo/repoRev routes;
	// otherwise it will match every subroute.
	//
	// App is the app ID (e.g., "issues"), and AppPath is an opaque
	// path that Sourcegraph passes directly to the app. The empty
	// AppPath is the app's homepage, and it manages its own subpaths.
	repoRev.PathPrefix(`/app/{App}{AppPath:(?:/.*)?}`).Name(RepoAppFrame)

	return &Router{*base}
}

func (r *Router) URLToOrError(routeName string, params ...string) (*url.URL, error) {
	route := r.Get(routeName)
	if route == nil {
		log.Panicf("no such route: %q (params: %v)", routeName, params)
	}
	u, err := route.URL(params...)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Router) URLTo(routeName string, params ...string) *url.URL {
	u, err := r.URLToOrError(routeName, params...)
	if err != nil {
		if os.Getenv("STRICT_URL_GEN") != "" && *u == (url.URL{}) {
			log.Panicf("Failed to generate route. See log message above.")
		}
		log.Printf("Route error: failed to make URL for route %q (params: %v): %s", routeName, params, err)
		return &url.URL{}
	}
	return u
}

var Rel = New(nil)

// gone returns HTTP 410 Gone, which indicates that this is an
// "expected 404" and suppresses it from our 404 logs.
func gone(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusGone)
}
