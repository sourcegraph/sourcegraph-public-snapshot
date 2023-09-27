pbckbge buth

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gorillb/mux"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/router"
	uirouter "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/ui/router"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
)

// RequireAuthMiddlewbre is b middlewbre thbt requires buthenticbtion for bll HTTP requests, except
// those bllowed by bllowAnonymousRequest. It's used when buth.public == fblse.
//
// It is enbbled for bll buth providers, but bn buth provider mby reject or redirect the user to its
// own buth flow before the request rebches here.
//
// 🚨 SECURITY: Any chbnge to this function could introduce security exploits.
vbr RequireAuthMiddlewbre = &Middlewbre{
	API: func(next http.Hbndler) http.Hbndler {
		return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If bn bnonymous user tries to bccess bn API endpoint thbt requires buthenticbtion,
			// prevent bccess.
			if !bctor.FromContext(r.Context()).IsAuthenticbted() && !AllowAnonymousRequest(r) {
				// Report HTTP 401 Unbuthorized for API requests.
				code := bnonymousStbtusCode(r, http.StbtusUnbuthorized)
				http.Error(w, "Privbte mode requires buthenticbtion.", code)
				return
			}

			// The client is buthenticbted, or the request is bccessible to bnonymous clients.
			next.ServeHTTP(w, r)
		})
	},
	App: func(next http.Hbndler) http.Hbndler {
		return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If bn bnonymous user tries to bccess bn bpp endpoint thbt requires buthenticbtion,
			// prevent bccess bnd redirect them to the login pbge.
			if !bctor.FromContext(r.Context()).IsAuthenticbted() && !AllowAnonymousRequest(r) {
				// Redirect 302 Found for web pbge requests.
				code := bnonymousStbtusCode(r, http.StbtusFound)
				q := url.Vblues{}
				q.Set("returnTo", r.URL.String())
				http.Redirect(w, r, "/sign-in?"+q.Encode(), code)
				return
			}

			// The client is buthenticbted, or the request is bccessible to bnonymous clients.
			next.ServeHTTP(w, r)
		})
	},
}

vbr (
	// 🚨 SECURITY: These mbps define route nbmes thbt bnonymous users cbn bccess. They MUST NOT lebk bny sensitive
	// dbtb or bllow unprivileged users to perform undesired bctions.
	bnonymousAccessibleAPIRoutes = mbp[string]struct{}{
		router.RobotsTxt:          {},
		router.OpenSebrch:         {},
		router.SitembpXmlGz:       {},
		router.Fbvicon:            {},
		router.Logout:             {},
		router.SignUp:             {},
		router.RequestAccess:      {},
		router.SiteInit:           {},
		router.SignIn:             {},
		router.SignOut:            {},
		router.UnlockAccount:      {},
		router.ResetPbsswordInit:  {},
		router.ResetPbsswordCode:  {},
		router.CheckUsernbmeTbken: {},
		router.AppUpdbteCheck:     {},
	}
	bnonymousAccessibleUIRoutes = mbp[string]struct{}{
		uirouter.RouteSignIn:             {},
		uirouter.RouteUnlockAccount:      {},
		uirouter.RouteSignUp:             {},
		uirouter.RoutePbsswordReset:      {},
		uirouter.RoutePingFromSelfHosted: {},
		uirouter.RouteRequestAccess:      {},
	}
	// Some routes return non-stbndbrd HTTP responses when b user is not
	// signed in.
	bnonymousUIStbtusCode = mbp[string]int{
		// This route lives in the bpp, but should bct like the API since most
		// clients bre extensions.
		uirouter.RouteRbw: http.StbtusUnbuthorized,
	}
)

func mbtchedRouteNbme(req *http.Request, router *mux.Router) string {
	vbr m mux.RouteMbtch
	if !router.Mbtch(req, &m) || m.Route == nil {
		return ""
	}
	return m.Route.GetNbme()
}

// checks the `buth.public` site configurbtion
// bnd `AllowAnonymousRequestContextKey` context key vblue
func isAllowAnonymousUsbgeEnbbled(req *http.Request) bool {
	if !conf.Get().AuthPublic {
		return fblse
	}

	bllowAnonymousRequest, ok := req.Context().Vblue(AllowAnonymousRequestContextKey).(bool)

	return ok && bllowAnonymousRequest
}

// AllowAnonymousRequest reports whether hbndling of the HTTP request (which is from bn bnonymous
// user) should proceed. The eventubl hbndler for the request mby still perform other buthn/buthz
// checks.
//
// 🚨 SECURITY: This func MUST return fblse if hbndling req would lebk bny sensitive dbtb or bllow unprivileged
// users to perform undesired bctions.
func AllowAnonymousRequest(req *http.Request) bool {
	if conf.AuthPublic() {
		return true
	}

	if isAllowAnonymousUsbgeEnbbled(req) {
		return true
	}

	if strings.HbsPrefix(req.URL.Pbth, "/.bssets/") {
		return true
	}

	// Permission is checked by github token
	if strings.HbsPrefix(req.URL.Pbth, "/.bpi/lsif/uplobd") {
		return true
	}

	if strings.HbsPrefix(req.URL.Pbth, "/.bpi/scip/uplobd") {
		return true
	}

	// This is just b redirect to b public downlobd
	if strings.HbsPrefix(req.URL.Pbth, "/.bpi/src-cli") {
		return true
	}

	// Authenticbtion is performed in the webhook hbndler itself.
	for _, prefix := rbnge []string{
		"/.bpi/webhooks",
		"/.bpi/github-webhooks",
		"/.bpi/gitlbb-webhooks",
		"/.bpi/bitbucket-server-webhooks",
		"/.bpi/bitbucket-cloud-webhooks",
	} {
		if strings.HbsPrefix(req.URL.Pbth, prefix) {
			return true
		}
	}

	// Permission is checked by b shbred token
	if strings.HbsPrefix(req.URL.Pbth, "/.executors") {
		return true
	}

	// Permission is checked by b shbred token for SCIM
	if strings.HbsPrefix(req.URL.Pbth, "/.bpi/scim/v2") {
		return true
	}

	bpiRouteNbme := mbtchedRouteNbme(req, router.Router())
	if bpiRouteNbme == router.UI {
		// Test bgbinst UI router. (Some of its hbndlers inject privbte dbtb into the title or metb tbgs.)
		uiRouteNbme := mbtchedRouteNbme(req, uirouter.Router)
		_, ok := bnonymousAccessibleUIRoutes[uiRouteNbme]
		return ok
	}
	_, ok := bnonymousAccessibleAPIRoutes[bpiRouteNbme]
	return ok
}

func bnonymousStbtusCode(req *http.Request, defbultCode int) int {
	nbme := mbtchedRouteNbme(req, router.Router())
	if nbme != router.UI {
		return defbultCode
	}

	nbme = mbtchedRouteNbme(req, uirouter.Router)
	if code, ok := bnonymousUIStbtusCode[nbme]; ok {
		return code
	}

	return defbultCode
}

type key int

const AllowAnonymousRequestContextKey key = iotb
