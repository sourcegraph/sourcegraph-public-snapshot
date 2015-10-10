// Package router contains the URL router for the frontend app.
package router

import (
	"log"
	"net/url"
	"os"

	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/go-sourcegraph/routevar"
	"sourcegraph.com/sourcegraph/go-sourcegraph/spec"
	"src.sourcegraph.com/sourcegraph/app/appconf"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	gitrouter "src.sourcegraph.com/sourcegraph/gitserver/router"
)

const (
	About = "about"

	BlogIndex     = "blog"
	BlogIndexAtom = "blog[.atom]"
	BlogPost      = "blog.post"

	Liveblog = "liveblog"

	Builds = "builds"

	Home       = "home"
	BetaSignup = "beta-signup"

	RegisterClient = "register-client"

	DownloadInstall = "download.install"
	Download        = "download"

	SitemapIndex = "sitemap-index"
	RepoSitemap  = "repo.sitemap"

	User                           = "person"
	UserSettingsEmails             = "person.settings.emails"
	UserSettingsProfile            = "person.settings.profile"
	UserSettingsAuth               = "person.settings.auth"
	UserSettingsIntegrations       = "person.settings.integrations"
	UserSettingsIntegrationsUpdate = "person.settings.integrations.update"
	UserOrgs                       = "person.orgs"
	OrgMembers                     = "org.members"

	Repo               = "repo"
	RepoBadges         = "repo.badges"
	RepoCounters       = "repo.counters"
	RepoBuilds         = "repo.builds"
	RepoBuild          = "repo.build"
	RepoBuildUpdate    = "repo.build.update"
	RepoBuildLog       = "repo.build.log"
	RepoBuildTaskLog   = "repo.build.task.log"
	RepoBuildsCreate   = "repo.builds.create"
	RepoSearch         = "repo.search"
	RepoRefresh        = "repo.refresh"
	RepoTree           = "repo.tree"
	RepoTreeShare      = "repo.tree.share"
	RepoCompare        = "repo.compare"
	RepoCompareAll     = "repo.compare.all"
	RepoDiscussion     = "repo.discussion"
	RepoDiscussionList = "repo.discussions"

	RepoEnable = "repo.enable"

	Changeset            = "repo.changeset"
	ChangesetList        = "repo.changeset.list"
	ChangesetFiles       = "repo.changeset.files"
	ChangesetFilesFilter = "repo.changeset.files.filter"

	RepoRevCommits = "repo.rev.commits"
	RepoCommit     = "repo.commit"
	RepoTags       = "repo.tags"
	RepoBranches   = "repo.branches"

	SearchForm    = "search.form"
	SearchResults = "search.results"

	SourceboxFile = "sourcebox.file"
	SourceboxDef  = "sourcebox.def"

	LogIn          = "log-in"
	LogOut         = "log-out"
	SignUp         = "sign-up"
	ForgotPassword = "forgot-password"
	ResetPassword  = "reset-password"

	OAuth2ServerAuthorize = "oauth-provider.authorize"
	OAuth2ServerToken     = "oauth-provider.token"

	OAuth2ClientReceive = "oauth-client.receive"

	Def         = "def"
	DefExamples = "def.examples"
	DefPopover  = "def.popover"
	DefShare    = "def.share"

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

	base.PathPrefix("/blog/live").Name(Liveblog)

	base.Path(`/blog`).Methods("GET").Name(BlogIndex)
	base.Path(`/blog{Format:\.atom}`).Methods("GET").Name(BlogIndexAtom)
	base.Path("/blog/{Slug:.*}").Methods("GET").Name(BlogPost)

	base.Path("/.builds").Methods("GET").Name(Builds)
	base.Path("/.beta-signup").Methods("POST").Name(BetaSignup)

	base.Path("/.download/install.sh").Methods("GET").Name(DownloadInstall)
	base.Path("/.download/{Suffix:.*}").Methods("GET").Name(Download)

	base.Path("/search").Methods("GET").Queries("q", "").Name(SearchResults)
	base.Path("/search").Methods("GET").Name(SearchForm)

	base.Path("/login").Methods("GET", "POST").Name(LogIn)
	base.Path("/join").Methods("GET", "POST").Name(SignUp)
	base.Path("/logout").Methods("POST").Name(LogOut)
	base.Path("/forgot").Methods("GET", "POST").Name(ForgotPassword)
	base.Path("/reset").Methods("GET", "POST").Name(ResetPassword)

	base.Path("/login/oauth/authorize").Methods("GET").Name(OAuth2ServerAuthorize)
	base.Path("/login/oauth/token").Methods("POST").Name(OAuth2ServerToken)

	base.Path("/login/oauth/receive").Methods("GET").Name(OAuth2ClientReceive)

	base.Path("/sitemap-index.xml").Methods("GET").Name(SitemapIndex)

	// User routes begin with tilde (~).
	userPath := `/~` + routevar.User
	user := base.PathPrefix(userPath).Subrouter()
	user.Path("/.settings/profile").Methods("GET", "POST").Name(UserSettingsProfile)
	user.Path("/.settings/emails").Methods("GET").Name(UserSettingsEmails)
	if !appconf.Flags.DisableIntegrations {
		integrationsPath := "/.settings/integrations"
		integrations := user.PathPrefix(integrationsPath).Subrouter()
		user.Path(integrationsPath).Methods("GET").Name(UserSettingsIntegrations)
		integrations.Path("/{Integration}").Methods("POST").Name(UserSettingsIntegrationsUpdate)
	}
	if !authutil.ActiveFlags.DisableUserProfiles {
		base.Path(userPath).Methods("GET").Name(User)
		user.Path("/.orgs").Methods("GET").Name(UserOrgs)
		user.Path("/.members").Methods("GET").Name(OrgMembers)
		user.Path("/.settings/auth").Methods("GET").Name(UserSettingsAuth)
	}

	// attach git transport endpoints
	gitrouter.New(base)

	repo := base.PathPrefix(`/` + routevar.Repo).Subrouter()

	repoRevPath := `/` + routevar.RepoRev
	base.Path(repoRevPath).Methods("GET").PostMatchFunc(routevar.FixRepoRevVars).BuildVarsFunc(routevar.PrepareRepoRevRouteVars).Name(Repo)
	repoRev := base.PathPrefix(repoRevPath).PostMatchFunc(routevar.FixRepoRevVars).BuildVarsFunc(routevar.PrepareRepoRevRouteVars).Subrouter()

	// See router_util/def_route.go for an explanation of how we match def
	// routes.
	defPath := "/" + routevar.Def
	repoRev.Path(defPath).Methods("GET").PostMatchFunc(routevar.FixDefUnitVars).BuildVarsFunc(routevar.PrepareDefRouteVars).Name(Def)
	def := repoRev.PathPrefix(defPath).PostMatchFunc(routevar.FixDefUnitVars).BuildVarsFunc(routevar.PrepareDefRouteVars).Subrouter()
	def.Path("/.examples").Methods("GET").Name(DefExamples)
	def.Path("/.popover").Methods("GET").Name(DefPopover)
	def.Path("/.sourcebox.{Format}").Methods("GET").Name(SourceboxDef)
	def.Path("/.share").Methods("GET").Name(DefShare)
	// TODO(x): def history route

	// See router_util/tree_route.go for an explanation of how we match tree
	// entry routes.
	repoTreePath := "/.tree" + routevar.TreeEntryPath
	repoRev.Path(repoTreePath + "/.sourcebox.{Format}").PostMatchFunc(routevar.FixTreeEntryVars).BuildVarsFunc(routevar.PrepareTreeEntryRouteVars).Name(SourceboxFile)
	repoRev.Path(repoTreePath + "/.share").PostMatchFunc(routevar.FixTreeEntryVars).BuildVarsFunc(routevar.PrepareTreeEntryRouteVars).Name(RepoTreeShare)
	repoRev.Path(repoTreePath).Methods("GET").PostMatchFunc(routevar.FixTreeEntryVars).BuildVarsFunc(routevar.PrepareTreeEntryRouteVars).Name(RepoTree)

	repoRev.Path("/.refresh").Methods("POST", "PUT").Name(RepoRefresh)
	repoRev.Path("/.badges").Methods("GET").Name(RepoBadges)
	repoRev.Path("/.search").Methods("GET").Name(RepoSearch)
	repoRev.Path("/.counters").Methods("GET").Name(RepoCounters)
	repoRev.Path("/.commits").Methods("GET").Name(RepoRevCommits)
	repoRev.Path(`/.discussion/{ID:\d+}`).Methods("GET", "POST").Name(RepoDiscussion)
	repoRev.Path("/.discussions").Methods("GET").Name(RepoDiscussionList)

	repoRev.Path(`/.changesets/{ID:\d+}`).Methods("GET").Name(Changeset)
	repoRev.Path(`/.changesets`).Methods("GET").Name(ChangesetList)
	repoRev.Path(`/.changesets/{ID:\d+}/files`).Methods("GET").Name(ChangesetFiles)
	repoRev.Path(`/.changesets/{ID:\d+}/files/{Filter:.+}`).Methods("GET").Name(ChangesetFilesFilter)

	headVar := "{Head:" + routevar.NamedToNonCapturingGroups(spec.RevPattern) + "}"
	repoRev.Path("/.compare/" + headVar).Methods("GET").Name(RepoCompare)
	repoRev.Path("/.compare/" + headVar + "/.all").Methods("GET").Name(RepoCompareAll)

	repo.Path("/.enable").Methods("GET", "POST", "DELETE").Name(RepoEnable)

	repo.Path("/.commits/{Rev:" + spec.PathNoLeadingDotComponentPattern + "}").Methods("GET").Name(RepoCommit)
	repo.Path("/.branches").Methods("GET").Name(RepoBranches)
	repo.Path("/.tags").Methods("GET").Name(RepoTags)
	repo.Path("/.sitemap.xml").Methods("GET").Name(RepoSitemap)

	repo.Path("/.builds").Methods("GET").Name(RepoBuilds)
	repo.Path("/.builds").Methods("POST").Name(RepoBuildsCreate)
	repoBuildPath := `/.builds/{CommitID}/{Attempt:\d+}`
	repo.Path(repoBuildPath).Methods("GET").Name(RepoBuild)
	repo.Path(repoBuildPath).Methods("POST").Name(RepoBuildUpdate)
	repoBuild := repo.PathPrefix(repoBuildPath).Subrouter()
	repoBuild.Path("/log").Methods("GET").Name(RepoBuildLog)
	repoBuild.Path("/tasks/{TaskID}/log").Methods("GET").Name(RepoBuildTaskLog)

	// This route should be AFTER all other repo/repoRev routes;
	// otherwise it will match every subroute.
	//
	// App is the app ID (e.g., "issues"), and AppPath is an opaque
	// path that Sourcegraph passes directly to the app. The empty
	// AppPath is the app's homepage, and it manages its own subpaths.
	repoRev.PathPrefix(`/.{App}{AppPath:(?:/.*)?}`).Name(RepoAppFrame)

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
