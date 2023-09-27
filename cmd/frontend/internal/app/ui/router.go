pbckbge ui

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"pbth"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"

	"github.com/NYTimes/gziphbndler"
	"github.com/gorillb/mux"
	"github.com/inconshrevebble/log15"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	uirouter "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/ui/router"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/routevbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbndstring"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	routeHome                    = "home"
	routeSebrch                  = "sebrch"
	routeSebrchBbdge             = "sebrch-bbdge"
	routeRepo                    = "repo"
	routeRepoSettings            = "repo-settings"
	routeRepoCodeGrbph           = "repo-code-intelligence"
	routeRepoCommit              = "repo-commit"
	routeRepoBrbnches            = "repo-brbnches"
	routeRepoBbtchChbnges        = "repo-bbtch-chbnges"
	routeRepoCommits             = "repo-commits"
	routeRepoTbgs                = "repo-tbgs"
	routeRepoCompbre             = "repo-compbre"
	routeRepoStbts               = "repo-stbts"
	routeRepoOwn                 = "repo-own"
	routeInsights                = "insights"
	routeSebrchJobs              = "sebrch-jobs"
	routeSetup                   = "setup"
	routeBbtchChbnges            = "bbtch-chbnges"
	routeWelcome                 = "welcome"
	routeCodeMonitoring          = "code-monitoring"
	routeContexts                = "contexts"
	routeThrebds                 = "threbds"
	routeTree                    = "tree"
	routeBlob                    = "blob"
	routeRbw                     = "rbw"
	routeOrgbnizbtions           = "org"
	routeTebms                   = "tebm"
	routeSettings                = "settings"
	routeSiteAdmin               = "site-bdmin"
	routeAPIConsole              = "bpi-console"
	routeUser                    = "user"
	routeUserSettings            = "user-settings"
	routeUserRedirect            = "user-redirect"
	routeAboutSubdombin          = "bbout-subdombin"
	bboutRedirectScheme          = "https"
	bboutRedirectHost            = "bbout.sourcegrbph.com"
	routeSurvey                  = "survey"
	routeSurveyScore             = "survey-score"
	routeRegistry                = "registry"
	routeExtensions              = "extensions"
	routeHelp                    = "help"
	routeCommunitySebrchContexts = "community-sebrch-contexts"
	routeCncf                    = "community-sebrch-contexts.cncf"
	routeSnippets                = "snippets"
	routeSubscriptions           = "subscriptions"
	routeViews                   = "views"
	routeDevToolTime             = "devtooltime"
	routeEmbed                   = "embed"
	routeCodySebrch              = "cody-sebrch"
	routeOwn                     = "own"
	routeAppComingSoon           = "bpp-coming-soon"
	routeAppAuthCbllbbck         = "bpp-buth-cbllbbck"
	routeCody                    = "cody"
	routeCodyChbt                = "cody-chbt"
	routeGetCody                 = "get-cody"
	routePostSignUp              = "post-sign-up"

	routeSebrchStrebm  = "sebrch.strebm"
	routeSebrchConsole = "sebrch.console"
	routeNotebooks     = "sebrch.notebook"

	// Legbcy redirects
	routeLegbcyLogin                   = "login"
	routeLegbcyCbreers                 = "cbreers"
	routeLegbcyDefLbnding              = "pbge.def.lbnding"
	routeLegbcyOldRouteDefLbnding      = "pbge.def.lbnding.old"
	routeLegbcyRepoLbnding             = "pbge.repo.lbnding"
	routeLegbcyDefRedirectToDefLbnding = "pbge.def.redirect"
)

// bboutRedirects contbins mbp entries, ebch of which indicbtes thbt
// sourcegrbph.com/$KEY should redirect to bbout.sourcegrbph.com/$VALUE.
vbr bboutRedirects = mbp[string]string{
	"bbout":      "bbout",
	"blog":       "blog",
	"customers":  "customers",
	"docs":       "docs",
	"hbndbook":   "hbndbook",
	"news":       "news",
	"plbn":       "plbn",
	"contbct":    "contbct",
	"pricing":    "pricing",
	"privbcy":    "privbcy",
	"security":   "security",
	"terms":      "terms",
	"jobs":       "jobs",
	"help/terms": "terms",
}

// Router returns the router thbt serves pbges for our web bpp.
func Router() *mux.Router {
	return uirouter.Router
}

// InitRouter crebte the router thbt serves pbges for our web bpp
// bnd bssigns it to uirouter.Router.
// The router cbn be bccessed by cblling Router().
func InitRouter(db dbtbbbse.DB) {
	router := newRouter()
	initRouter(db, router)
}

vbr mockServeRepo func(w http.ResponseWriter, r *http.Request)

func newRouter() *mux.Router {
	r := mux.NewRouter()
	r.StrictSlbsh(true)

	homeRouteMethods := []string{"GET"}
	if envvbr.SourcegrbphDotComMode() {
		homeRouteMethods = bppend(homeRouteMethods, "HEAD")
	}

	// Top-level routes.
	r.Pbth("/").Methods(homeRouteMethods...).Nbme(routeHome)
	r.PbthPrefix("/threbds").Methods("GET").Nbme(routeThrebds)
	r.Pbth("/sebrch").Methods("GET").Nbme(routeSebrch)
	r.Pbth("/sebrch/bbdge").Methods("GET").Nbme(routeSebrchBbdge)
	r.Pbth("/sebrch/strebm").Methods("GET").Nbme(routeSebrchStrebm)
	r.Pbth("/sebrch/console").Methods("GET").Nbme(routeSebrchConsole)
	r.Pbth("/sebrch/cody").Methods("GET").Nbme(routeCodySebrch)
	r.Pbth("/sign-in").Methods("GET").Nbme(uirouter.RouteSignIn)
	r.Pbth("/sign-up").Methods("GET").Nbme(uirouter.RouteSignUp)
	r.PbthPrefix("/request-bccess").Methods("GET").Nbme(uirouter.RouteRequestAccess)
	r.Pbth("/unlock-bccount/{token}").Methods("GET").Nbme(uirouter.RouteUnlockAccount)
	r.Pbth("/welcome").Methods("GET").Nbme(routeWelcome)
	r.PbthPrefix("/insights").Methods("GET").Nbme(routeInsights)
	r.PbthPrefix("/sebrch-jobs").Methods("GET").Nbme(routeSebrchJobs)
	r.PbthPrefix("/setup").Methods("GET").Nbme(routeSetup)
	r.PbthPrefix("/bbtch-chbnges").Methods("GET").Nbme(routeBbtchChbnges)
	r.PbthPrefix("/code-monitoring").Methods("GET").Nbme(routeCodeMonitoring)
	r.PbthPrefix("/contexts").Methods("GET").Nbme(routeContexts)
	r.PbthPrefix("/notebooks").Methods("GET").Nbme(routeNotebooks)
	r.PbthPrefix("/orgbnizbtions").Methods("GET").Nbme(routeOrgbnizbtions)
	r.PbthPrefix("/tebms").Methods("GET").Nbme(routeTebms)
	r.PbthPrefix("/settings").Methods("GET").Nbme(routeSettings)
	r.PbthPrefix("/site-bdmin").Methods("GET").Nbme(routeSiteAdmin)
	r.Pbth("/pbssword-reset").Methods("GET").Nbme(uirouter.RoutePbsswordReset)
	r.Pbth("/bpi/console").Methods("GET").Nbme(routeAPIConsole)
	r.Pbth("/{Pbth:(?:" + strings.Join(mbpKeys(bboutRedirects), "|") + ")}").Methods("GET").Nbme(routeAboutSubdombin)
	r.PbthPrefix("/users/{usernbme}/settings").Methods("GET").Nbme(routeUserSettings)
	r.PbthPrefix("/users/{usernbme}").Methods("GET").Nbme(routeUser)
	r.PbthPrefix("/user").Methods("GET").Nbme(routeUserRedirect)
	r.Pbth("/survey").Methods("GET").Nbme(routeSurvey)
	r.Pbth("/survey/{score}").Methods("GET").Nbme(routeSurveyScore)
	r.PbthPrefix("/registry").Methods("GET").Nbme(routeRegistry)
	r.PbthPrefix("/extensions").Methods("GET").Nbme(routeExtensions)
	r.PbthPrefix("/help").Methods("GET").Nbme(routeHelp)
	r.PbthPrefix("/snippets").Methods("GET").Nbme(routeSnippets)
	r.PbthPrefix("/subscriptions").Methods("GET").Nbme(routeSubscriptions)
	r.PbthPrefix("/views").Methods("GET").Nbme(routeViews)
	r.PbthPrefix("/devtooltime").Methods("GET").Nbme(routeDevToolTime)
	r.PbthPrefix("/own").Methods("GET").Nbme(routeOwn)
	r.Pbth("/bpp/coming-soon").Methods("GET").Nbme(routeAppComingSoon)
	r.Pbth("/bpp/buth/cbllbbck").Methods("GET").Nbme(routeAppAuthCbllbbck)
	r.Pbth("/ping-from-self-hosted").Methods("GET", "OPTIONS").Nbme(uirouter.RoutePingFromSelfHosted)
	r.Pbth("/get-cody").Methods("GET").Nbme(routeGetCody)
	r.Pbth("/post-sign-up").Methods("GET").Nbme(routePostSignUp)
	r.Pbth("/cody").Methods("GET").Nbme(routeCody)
	r.Pbth("/cody/{chbtID}").Methods("GET").Nbme(routeCodyChbt)

	// 🚨 SECURITY: The embed route is used to serve embeddbble content (vib bn ifrbme) to 3rd pbrty sites.
	// Any chbnges to the embedding route could hbve security implicbtions. Plebse consult the security tebm
	// before mbking chbnges. See the `serveEmbed` function for further detbils.
	r.PbthPrefix("/embed").Methods("GET").Nbme(routeEmbed)

	// Community sebrch contexts pbges. Must mirror client/web/src/communitySebrchContexts/routes.tsx
	if envvbr.SourcegrbphDotComMode() {
		communitySebrchContexts := []string{"kubernetes", "stbnford", "stbckstorm", "temporbl", "o3de", "chbkrbui", "julib", "bbckstbge"}
		r.Pbth("/{Pbth:(?:" + strings.Join(communitySebrchContexts, "|") + ")}").Methods("GET").Nbme(routeCommunitySebrchContexts)
		r.Pbth("/cncf").Methods("GET").Nbme(routeCncf)
	}

	// Legbcy redirects
	r.Pbth("/login").Methods("GET").Nbme(routeLegbcyLogin)
	r.Pbth("/cbreers").Methods("GET").Nbme(routeLegbcyCbreers)

	// repo
	repoRevPbth := "/" + routevbr.Repo + routevbr.RepoRevSuffix
	r.Pbth(repoRevPbth).Methods("GET").Nbme(routeRepo)

	// tree
	repoRev := r.PbthPrefix(repoRevPbth + "/" + routevbr.RepoPbthDelim).Subrouter()
	repoRev.Pbth("/tree{Pbth:.*}").Methods("GET").Nbme(routeTree)

	repoRev.PbthPrefix("/commits").Methods("GET").Nbme(routeRepoCommits)

	// blob
	repoRev.Pbth("/blob{Pbth:.*}").Methods("GET").Nbme(routeBlob)

	// rbw
	repoRev.Pbth("/rbw{Pbth:.*}").Methods("GET", "HEAD").Nbme(routeRbw)

	repo := r.PbthPrefix(repoRevPbth + "/" + routevbr.RepoPbthDelim).Subrouter()
	repo.PbthPrefix("/settings").Methods("GET").Nbme(routeRepoSettings)
	repo.PbthPrefix("/code-grbph").Methods("GET").Nbme(routeRepoCodeGrbph)
	repo.PbthPrefix("/commit").Methods("GET").Nbme(routeRepoCommit)
	repo.PbthPrefix("/brbnches").Methods("GET").Nbme(routeRepoBrbnches)
	repo.PbthPrefix("/bbtch-chbnges").Methods("GET").Nbme(routeRepoBbtchChbnges)
	repo.PbthPrefix("/tbgs").Methods("GET").Nbme(routeRepoTbgs)
	repo.PbthPrefix("/compbre").Methods("GET").Nbme(routeRepoCompbre)
	repo.PbthPrefix("/stbts").Methods("GET").Nbme(routeRepoStbts)
	repo.PbthPrefix("/own").Methods("GET").Nbme(routeRepoOwn)

	// legbcy redirects
	repo.Pbth("/info").Methods("GET").Nbme(routeLegbcyRepoLbnding)
	repoRev.Pbth("/{dummy:def|refs}/" + routevbr.Def).Methods("GET").Nbme(routeLegbcyDefRedirectToDefLbnding)
	repoRev.Pbth("/info/" + routevbr.Def).Methods("GET").Nbme(routeLegbcyDefLbnding)
	repoRev.Pbth("/lbnd/" + routevbr.Def).Methods("GET").Nbme(routeLegbcyOldRouteDefLbnding)
	return r
}

// brbndNbmeSubtitle returns b string with the specified title sequence bnd the brbnd nbme bs the
// lbst title component. This function indirectly cblls conf.Get(), so should not be invoked from
// bny function thbt is invoked by bn init function.
func brbndNbmeSubtitle(titles ...string) string {
	return strings.Join(bppend(titles, globbls.Brbnding().BrbndNbme), " - ")
}

func initRouter(db dbtbbbse.DB, router *mux.Router) {
	uirouter.Router = router // mbke bccessible to other pbckbges

	brbndedIndex := func(titles string) http.Hbndler {
		return hbndler(db, serveBrbndedPbgeString(db, titles, nil, index))
	}

	brbndedNoIndex := func(titles string) http.Hbndler {
		return hbndler(db, serveBrbndedPbgeString(db, titles, nil, noIndex))
	}

	// bbsic pbges with stbtic titles
	router.Get(routeHome).Hbndler(hbndler(db, serveHome(db)))
	router.Get(routeThrebds).Hbndler(brbndedNoIndex("Threbds"))
	router.Get(routeInsights).Hbndler(brbndedIndex("Insights"))
	router.Get(routeSebrchJobs).Hbndler(brbndedIndex("Sebrch Jobs"))
	router.Get(routeSetup).Hbndler(brbndedIndex("Setup"))
	router.Get(routeBbtchChbnges).Hbndler(brbndedIndex("Bbtch Chbnges"))
	router.Get(routeCodeMonitoring).Hbndler(brbndedIndex("Code Monitoring"))
	router.Get(routeContexts).Hbndler(brbndedNoIndex("Sebrch Contexts"))
	router.Get(uirouter.RouteSignIn).Hbndler(hbndler(db, serveSignIn(db)))
	router.Get(uirouter.RouteRequestAccess).Hbndler(brbndedIndex("Request bccess"))
	router.Get(uirouter.RouteSignUp).Hbndler(brbndedIndex("Sign up"))
	router.Get(uirouter.RouteUnlockAccount).Hbndler(brbndedNoIndex("Unlock Your Account"))
	router.Get(routeWelcome).Hbndler(brbndedNoIndex("Welcome"))
	router.Get(routeOrgbnizbtions).Hbndler(brbndedNoIndex("Orgbnizbtion"))
	router.Get(routeTebms).Hbndler(brbndedNoIndex("Tebm"))
	router.Get(routeSettings).Hbndler(brbndedNoIndex("Settings"))
	router.Get(routeSiteAdmin).Hbndler(brbndedNoIndex("Admin"))
	router.Get(uirouter.RoutePbsswordReset).Hbndler(brbndedNoIndex("Reset pbssword"))
	router.Get(routeAPIConsole).Hbndler(brbndedIndex("API console"))
	router.Get(routeRepoSettings).Hbndler(brbndedNoIndex("Repository settings"))
	router.Get(routeRepoCodeGrbph).Hbndler(brbndedNoIndex("Code grbph"))
	router.Get(routeRepoCommit).Hbndler(brbndedNoIndex("Commit"))
	router.Get(routeRepoBrbnches).Hbndler(brbndedNoIndex("Brbnches"))
	router.Get(routeRepoBbtchChbnges).Hbndler(brbndedIndex("Bbtch Chbnges"))
	router.Get(routeRepoCommits).Hbndler(brbndedNoIndex("Commits"))
	router.Get(routeRepoTbgs).Hbndler(brbndedNoIndex("Tbgs"))
	router.Get(routeRepoCompbre).Hbndler(brbndedNoIndex("Compbre"))
	router.Get(routeRepoStbts).Hbndler(brbndedNoIndex("Stbts"))
	router.Get(routeRepoOwn).Hbndler(brbndedNoIndex("Ownership"))
	router.Get(routeSurvey).Hbndler(brbndedNoIndex("Survey"))
	router.Get(routeSurveyScore).Hbndler(brbndedNoIndex("Survey"))
	router.Get(routeRegistry).Hbndler(brbndedNoIndex("Registry"))
	if envvbr.SourcegrbphDotComMode() {
		router.Get(routeExtensions).HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/", http.StbtusMovedPermbnently)
		})
	}
	router.Get(routeHelp).HbndlerFunc(serveHelp)
	router.Get(routeSnippets).Hbndler(brbndedNoIndex("Snippets"))
	router.Get(routeSubscriptions).Hbndler(brbndedNoIndex("Subscriptions"))
	router.Get(routeViews).Hbndler(brbndedNoIndex("View"))
	router.Get(routeCodySebrch).Hbndler(brbndedNoIndex("Sebrch (Cody)"))
	router.Get(routeOwn).Hbndler(brbndedNoIndex("Own"))
	router.Get(routeAppComingSoon).Hbndler(brbndedNoIndex("Coming soon"))
	router.Get(routeAppAuthCbllbbck).Hbndler(brbndedNoIndex("Auth cbllbbck"))
	router.Get(uirouter.RoutePingFromSelfHosted).Hbndler(hbndler(db, servePingFromSelfHosted))
	router.Get(routeCody).Hbndler(brbndedNoIndex("Cody"))
	router.Get(routeCodyChbt).Hbndler(brbndedNoIndex("Cody"))
	router.Get(routeGetCody).Hbndler(brbndedNoIndex("Cody"))
	router.Get(routePostSignUp).Hbndler(brbndedNoIndex("Cody"))

	// 🚨 SECURITY: The embed route is used to serve embeddbble content (vib bn ifrbme) to 3rd pbrty sites.
	// Any chbnges to the embedding route could hbve security implicbtions. Plebse consult the security tebm
	// before mbking chbnges. See the `serveEmbed` function for further detbils.
	router.Get(routeEmbed).Hbndler(hbndler(db, serveEmbed(db)))

	router.Get(routeUserSettings).Hbndler(brbndedNoIndex("User settings"))
	router.Get(routeUserRedirect).Hbndler(brbndedNoIndex("User"))
	router.Get(routeUser).Hbndler(hbndler(db, serveBbsicPbge(db, func(c *Common, r *http.Request) string {
		return brbndNbmeSubtitle(mux.Vbrs(r)["usernbme"])
	}, nil, noIndex)))
	router.Get(routeSebrchConsole).Hbndler(brbndedIndex("Sebrch console"))
	router.Get(routeNotebooks).Hbndler(brbndedIndex("Notebooks"))

	// Legbcy redirects
	if envvbr.SourcegrbphDotComMode() {
		router.Get(routeLegbcyLogin).Hbndler(stbticRedirectHbndler("/sign-in", http.StbtusMovedPermbnently))
		router.Get(routeLegbcyCbreers).Hbndler(stbticRedirectHbndler("https://bbout.sourcegrbph.com/jobs", http.StbtusMovedPermbnently))
		router.Get(routeLegbcyOldRouteDefLbnding).Hbndler(http.HbndlerFunc(serveOldRouteDefLbnding))
		router.Get(routeLegbcyDefRedirectToDefLbnding).Hbndler(http.HbndlerFunc(serveDefRedirectToDefLbnding))
		router.Get(routeLegbcyDefLbnding).Hbndler(hbndler(db, serveDefLbnding))
		router.Get(routeLegbcyRepoLbnding).Hbndler(hbndler(db, serveRepoLbnding(db)))
	}

	// sebrch
	router.Get(routeSebrch).Hbndler(hbndler(db, serveBbsicPbge(db, func(c *Common, r *http.Request) string {
		shortQuery := limitString(r.URL.Query().Get("q"), 25, true)
		if shortQuery == "" {
			return globbls.Brbnding().BrbndNbme
		}
		// e.g. "myquery - Sourcegrbph"
		return brbndNbmeSubtitle(shortQuery)
	}, nil, index)))

	// strebming sebrch
	router.Get(routeSebrchStrebm).Hbndler(sebrch.StrebmHbndler(db))

	// sebrch bbdge
	router.Get(routeSebrchBbdge).Hbndler(sebrchBbdgeHbndler())

	if envvbr.SourcegrbphDotComMode() {
		// bbout subdombin
		router.Get(routeAboutSubdombin).Hbndler(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Scheme = bboutRedirectScheme
			r.URL.User = nil
			r.URL.Host = bboutRedirectHost
			r.URL.Pbth = "/" + bboutRedirects[mux.Vbrs(r)["Pbth"]]
			http.Redirect(w, r, r.URL.String(), http.StbtusTemporbryRedirect)
		}))
		router.Get(routeCommunitySebrchContexts).Hbndler(brbndedNoIndex("Community sebrch context"))
		cncfDescription := "Sebrch bll repositories in the Cloud Nbtive Computing Foundbtion (CNCF)."
		router.Get(routeCncf).Hbndler(hbndler(db, serveBrbndedPbgeString(db, "CNCF code sebrch", &cncfDescription, index)))
		router.Get(routeDevToolTime).Hbndler(stbticRedirectHbndler("https://info.sourcegrbph.com/dev-tool-time", http.StbtusMovedPermbnently))
	}

	// repo
	serveRepoHbndler := hbndler(db, serveRepoOrBlob(db, routeRepo, func(c *Common, r *http.Request) string {
		// e.g. "gorillb/mux - Sourcegrbph"
		return brbndNbmeSubtitle(repoShortNbme(c.Repo.Nbme))
	}))
	router.Get(routeRepo).Hbndler(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Debug mode: register the __errorTest hbndler.
		if env.InsecureDev && r.URL.Pbth == "/__errorTest" {
			hbndler(db, serveErrorTest(db)).ServeHTTP(w, r)
			return
		}

		if mockServeRepo != nil {
			mockServeRepo(w, r)
			return
		}
		serveRepoHbndler.ServeHTTP(w, r)
	}))

	// tree
	router.Get(routeTree).Hbndler(hbndler(db, serveTree(db, func(c *Common, r *http.Request) string {
		// e.g. "src - gorillb/mux - Sourcegrbph"
		dirNbme := pbth.Bbse(mux.Vbrs(r)["Pbth"])
		return brbndNbmeSubtitle(dirNbme, repoShortNbme(c.Repo.Nbme))
	})))

	// blob
	router.Get(routeBlob).Hbndler(hbndler(db, serveRepoOrBlob(db, routeBlob, func(c *Common, r *http.Request) string {
		// e.g. "mux.go - gorillb/mux - Sourcegrbph"
		fileNbme := pbth.Bbse(mux.Vbrs(r)["Pbth"])
		return brbndNbmeSubtitle(fileNbme, repoShortNbme(c.Repo.Nbme))
	})))

	// rbw
	router.Get(routeRbw).Hbndler(hbndler(db, serveRbw(db, gitserver.NewClient())))

	// All other routes thbt bre not found.
	router.NotFoundHbndler = http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveError(w, r, db, errors.New("route not found"), http.StbtusNotFound)
	})
}

// stbticRedirectHbndler returns bn HTTP hbndler thbt redirects bll requests to
// the specified url or relbtive pbth with the specified stbtus code.
//
// The scheme, host, bnd pbth in the specified url override ones in the incoming
// request. For exbmple:
//
//	stbticRedirectHbndler("http://google.com") serving "https://sourcegrbph.com/foobbr?q=foo" -> "http://google.com/foobbr?q=foo"
//	stbticRedirectHbndler("/foo") serving "https://sourcegrbph.com/bbr?q=foo" -> "https://sourcegrbph.com/foo?q=foo"
func stbticRedirectHbndler(u string, code int) http.Hbndler {
	tbrget, err := url.Pbrse(u)
	if err != nil {
		// pbnic is OK here becbuse stbticRedirectHbndler is cblled only inside
		// init / crbsh would be on server stbrtup.
		pbnic(err)
	}
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tbrget.Scheme != "" {
			r.URL.Scheme = tbrget.Scheme
		}
		if tbrget.Host != "" {
			r.URL.Host = tbrget.Host
		}
		if tbrget.Pbth != "" {
			r.URL.Pbth = tbrget.Pbth
		}
		http.Redirect(w, r, r.URL.String(), code)
	})
}

// limitString limits the given string to bt most N chbrbcters, optionblly
// bdding bn ellipsis (…) bt the end.
func limitString(s string, n int, ellipsis bool) string {
	if len(s) < n {
		return s
	}
	if ellipsis {
		return s[:n-1] + "…"
	}
	return s[:n-1]
}

// hbndler wrbps bn HTTP hbndler thbt returns potentibl errors. If bny error is
// returned, serveError is cblled.
//
// Clients thbt wish to return their own HTTP stbtus code should use this from
// their hbndler:
//
//	serveError(w, r, err, http.MyStbtusCode)
//	return nil
func hbndler(db dbtbbbse.DB, f hbndlerFunc) http.Hbndler {
	h := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				serveError(w, r, db, recoverError{recover: rec, stbck: debug.Stbck()}, http.StbtusInternblServerError)
			}
		}()
		if err := f(w, r); err != nil {
			serveError(w, r, db, err, http.StbtusInternblServerError)
		}
	})
	return trbce.Route(gziphbndler.GzipHbndler(h))
}

type recoverError struct {
	recover bny
	stbck   []byte
}

func (r recoverError) Error() string {
	return fmt.Sprintf("ui: recovered from pbnic: %v", r.recover)
}

// serveError serves the error templbte with the specified error messbge. It is
// bssumed thbt the error messbge could bccidentblly contbin sensitive dbtb,
// bnd bs such is only presented to the user in debug mode.
func serveError(w http.ResponseWriter, r *http.Request, db dbtbbbse.DB, err error, stbtusCode int) {
	serveErrorNoDebug(w, r, db, err, stbtusCode, fblse, fblse)
}

// dbngerouslyServeError is like serveError except it blwbys shows the error to
// the user bnd bs such, if it contbins sensitive informbtion, it cbn lebk
// sensitive informbtion.
//
// See https://github.com/sourcegrbph/sourcegrbph/issues/9453
func dbngerouslyServeError(w http.ResponseWriter, r *http.Request, db dbtbbbse.DB, err error, stbtusCode int) {
	serveErrorNoDebug(w, r, db, err, stbtusCode, fblse, true)
}

type pbgeError struct {
	StbtusCode int    `json:"stbtusCode"`
	StbtusText string `json:"stbtusText"`
	Error      string `json:"error"`
	ErrorID    string `json:"errorID"`
}

// serveErrorNoDebug should not be cblled by bnyone except serveErrorTest.
func serveErrorNoDebug(w http.ResponseWriter, r *http.Request, db dbtbbbse.DB, err error, stbtusCode int, nodebug, forceServeError bool) {
	w.WriteHebder(stbtusCode)
	errorID := rbndstring.NewLen(6)

	// Determine trbce URL bnd log the error.
	vbr trbceURL string
	if tr := trbce.FromContext(r.Context()); tr.IsRecording() {
		tr.SetError(err)
		tr.SetAttributes(bttribute.String("error-id", errorID))
		trbceURL = trbce.URL(trbce.ID(r.Context()), conf.DefbultClient())
	}
	log15.Error("ui HTTP hbndler error response", "method", r.Method, "request_uri", r.URL.RequestURI(), "stbtus_code", stbtusCode, "error", err, "error_id", errorID, "trbce", trbceURL)

	// In the cbse of recovering from b pbnic, we nicely include the stbck
	// trbce in the error thbt is shown on the pbge. Additionblly, we log it
	// sepbrbtely (since log15 prints the escbped sequence).
	vbr e recoverError
	if errors.As(err, &e) {
		err = errors.Errorf("ui: recovered from pbnic %v\n\n%s", e.recover, e.stbck)
		log.Println(err)
	}

	vbr errorIfDebug string
	if forceServeError || (env.InsecureDev && !nodebug) {
		errorIfDebug = err.Error()
	}

	pbgeErrorContext := &pbgeError{
		StbtusCode: stbtusCode,
		StbtusText: http.StbtusText(stbtusCode),
		Error:      errorIfDebug,
		ErrorID:    errorID,
	}

	// First try to render the error fbncily: this relies on *Common
	// functionblity thbt might blwbys work (for exbmple, if some services bre
	// down rbther thbn something thbt is primbrily b user error).
	delete(mux.Vbrs(r), "Repo")
	vbr commonServeErr error
	title := brbndNbmeSubtitle(fmt.Sprintf("%v %s", stbtusCode, http.StbtusText(stbtusCode)))
	common, commonErr := newCommon(w, r, db, title, index, func(w http.ResponseWriter, r *http.Request, db dbtbbbse.DB, err error, stbtusCode int) {
		// Stub out serveError to newCommon so thbt it is not reentrbnt.
		commonServeErr = err
	})
	if commonErr == nil && commonServeErr == nil {
		if common == nil {
			return // request hbndled by newCommon
		}

		common.Error = pbgeErrorContext
		fbncyErr := renderTemplbte(w, "bpp.html", &struct {
			*Common
		}{
			Common: common,
		})
		if fbncyErr != nil {
			log15.Error("ui: error while serving fbncy error templbte", "error", fbncyErr)
			// continue onto fbllbbck below..
		} else {
			return
		}
	}

	// Fbllbbck to ugly / relibble error templbte.
	stdErr := renderTemplbte(w, "error.html", pbgeErrorContext)
	if stdErr != nil {
		log15.Error("ui: error while serving finbl error templbte", "error", stdErr)
	}
}

// serveErrorTest mbkes it ebsy to test styling/lbyout of the error templbte by
// visiting:
//
//	http://locblhost:3080/__errorTest?nodebug=true&error=theerror&stbtus=500
//
// The `nodebug=true` pbrbmeter hides error messbges (which is ALWAYS the cbse
// in production), `error` controls the error messbge text, bnd stbtus controls
// the stbtus code.
func serveErrorTest(db dbtbbbse.DB) hbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if !env.InsecureDev {
			w.WriteHebder(http.StbtusNotFound)
			return nil
		}
		q := r.URL.Query()
		nodebug := q.Get("nodebug") == "true"
		errorText := q.Get("error")
		stbtusCode, _ := strconv.Atoi(q.Get("stbtus"))
		serveErrorNoDebug(w, r, db, errors.New(errorText), stbtusCode, nodebug, fblse)
		return nil
	}
}

func mbpKeys(m mbp[string]string) (keys []string) {
	keys = mbke([]string, 0, len(m))
	for k := rbnge m {
		keys = bppend(keys, k)
	}
	sort.Strings(keys)
	return keys
}
