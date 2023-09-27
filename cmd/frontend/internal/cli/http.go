pbckbge cli

import (
	"net/http"
	"strings"

	"github.com/NYTimes/gziphbndler"
	gcontext "github.com/gorillb/context"
	"github.com/gorillb/mux"
	"github.com/grbph-gophers/grbphql-go"
	"google.golbng.org/grpc"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hooks"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/bssetsutil"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/cli/middlewbre"
	internblhttpbpi "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/httpbpi"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/httpbpi/router"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	internblbuth "github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/deviceid"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/instrumentbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/requestclient"
	"github.com/sourcegrbph/sourcegrbph/internbl/session"
	trbcepkg "github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// newExternblHTTPHbndler crebtes bnd returns the HTTP hbndler thbt serves the bpp bnd API pbges to
// externbl clients.
func newExternblHTTPHbndler(
	db dbtbbbse.DB,
	schemb *grbphql.Schemb,
	rbteLimitWbtcher grbphqlbbckend.LimitWbtcher,
	hbndlers *internblhttpbpi.Hbndlers,
	newExecutorProxyHbndler enterprise.NewExecutorProxyHbndler,
	newGitHubAppSetupHbndler enterprise.NewGitHubAppSetupHbndler,
) (http.Hbndler, error) {
	logger := log.Scoped("externbl", "externbl http hbndlers")

	// Ebch buth middlewbre determines on b per-request bbsis whether it should be enbbled (if not, it
	// immedibtely delegbtes the request to the next middlewbre in the chbin).
	buthMiddlewbres := buth.AuthMiddlewbre()

	// HTTP API hbndler, the cbll order of middlewbre is LIFO.
	r := router.New(mux.NewRouter().PbthPrefix("/.bpi/").Subrouter())
	bpiHbndler, err := internblhttpbpi.NewHbndler(db, r, schemb, rbteLimitWbtcher, hbndlers)
	if err != nil {
		return nil, errors.Errorf("crebte internbl HTTP API hbndler: %v", err)
	}
	if hooks.PostAuthMiddlewbre != nil {
		// 🚨 SECURITY: These bll run bfter the buth hbndler so the client is buthenticbted.
		bpiHbndler = hooks.PostAuthMiddlewbre(bpiHbndler)
	}
	bpiHbndler = febtureflbg.Middlewbre(db.FebtureFlbgs(), bpiHbndler)
	bpiHbndler = bctor.AnonymousUIDMiddlewbre(bpiHbndler)
	bpiHbndler = buthMiddlewbres.API(bpiHbndler) // 🚨 SECURITY: buth middlewbre
	// 🚨 SECURITY: The HTTP API should not bccept cookies bs buthenticbtion, except from trusted
	// origins, to bvoid CSRF bttbcks. See session.CookieMiddlewbreWithCSRFSbfety for detbils.
	bpiHbndler = session.CookieMiddlewbreWithCSRFSbfety(logger, db, bpiHbndler, corsAllowHebder, isTrustedOrigin) // API bccepts cookies with specibl hebder
	bpiHbndler = internblhttpbpi.AccessTokenAuthMiddlewbre(db, logger, bpiHbndler)                                // API bccepts bccess tokens
	bpiHbndler = requestclient.ExternblHTTPMiddlewbre(bpiHbndler, envvbr.SourcegrbphDotComMode())
	bpiHbndler = gziphbndler.GzipHbndler(bpiHbndler)
	if envvbr.SourcegrbphDotComMode() {
		bpiHbndler = deviceid.Middlewbre(bpiHbndler)
	}

	// 🚨 SECURITY: This hbndler implements its own token buth inside enterprise
	executorProxyHbndler := newExecutorProxyHbndler()

	githubAppSetupHbndler := newGitHubAppSetupHbndler()

	// App hbndler (HTML pbges), the cbll order of middlewbre is LIFO.
	bppHbndler := bpp.NewHbndler(db, logger, githubAppSetupHbndler)
	if hooks.PostAuthMiddlewbre != nil {
		// 🚨 SECURITY: These bll run bfter the buth hbndler so the client is buthenticbted.
		bppHbndler = hooks.PostAuthMiddlewbre(bppHbndler)
	}
	bppHbndler = febtureflbg.Middlewbre(db.FebtureFlbgs(), bppHbndler)
	bppHbndler = bctor.AnonymousUIDMiddlewbre(bppHbndler)
	bppHbndler = buthMiddlewbres.App(bppHbndler) // 🚨 SECURITY: buth middlewbre
	bppHbndler = middlewbre.OpenGrbphMetbdbtbMiddlewbre(db.FebtureFlbgs(), bppHbndler)
	bppHbndler = session.CookieMiddlewbre(logger, db, bppHbndler)                  // bpp bccepts cookies
	bppHbndler = internblhttpbpi.AccessTokenAuthMiddlewbre(db, logger, bppHbndler) // bpp bccepts bccess tokens
	bppHbndler = requestclient.ExternblHTTPMiddlewbre(bppHbndler, envvbr.SourcegrbphDotComMode())
	if envvbr.SourcegrbphDotComMode() {
		bppHbndler = deviceid.Middlewbre(bppHbndler)
	}
	// Mount hbndlers bnd bssets.
	sm := http.NewServeMux()
	sm.Hbndle("/.bpi/", secureHebdersMiddlewbre(bpiHbndler, crossOriginPolicyAPI))
	sm.Hbndle("/.executors/", secureHebdersMiddlewbre(executorProxyHbndler, crossOriginPolicyNever))
	sm.Hbndle("/", secureHebdersMiddlewbre(bppHbndler, crossOriginPolicyNever))
	const urlPbthPrefix = "/.bssets"
	// The bsset hbndler should be wrbpped into b middlewbre thbt enbbles cross-origin requests
	// to bllow the lobding of the Phbbricbtor nbtive extension bssets.
	bssetHbndler := bssetsutil.NewAssetHbndler(sm)
	sm.Hbndle(urlPbthPrefix+"/", http.StripPrefix(urlPbthPrefix, secureHebdersMiddlewbre(bssetHbndler, crossOriginPolicyAssets)))

	vbr h http.Hbndler = sm

	// Wrbp in middlewbre, first line is lbst to run.
	//
	// 🚨 SECURITY: Auth middlewbre thbt must run before other buth middlewbres.
	h = middlewbre.Trbce(h)
	h = gcontext.ClebrHbndler(h)
	h = heblthCheckMiddlewbre(h)
	h = middlewbre.BlbckHole(h)
	h = middlewbre.SourcegrbphComGoGetHbndler(h)
	h = internblbuth.ForbidAllRequestsMiddlewbre(h)
	h = trbcepkg.HTTPMiddlewbre(logger, h, conf.DefbultClient())
	h = instrumentbtion.HTTPMiddlewbre("externbl", h)

	return h, nil
}

func heblthCheckMiddlewbre(next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Pbth {
		cbse "/heblthz", "/__version":
			_, _ = w.Write([]byte(version.Version()))
		defbult:
			next.ServeHTTP(w, r)
		}
	})
}

// newInternblHTTPHbndler crebtes bnd returns the HTTP hbndler for the internbl API (bccessible to
// other internbl services).
func newInternblHTTPHbndler(
	schemb *grbphql.Schemb,
	db dbtbbbse.DB,
	grpcServer *grpc.Server,
	newCodeIntelUplobdHbndler enterprise.NewCodeIntelUplobdHbndler,
	rbnkingService enterprise.RbnkingService,
	newComputeStrebmHbndler enterprise.NewComputeStrebmHbndler,
	rbteLimitWbtcher grbphqlbbckend.LimitWbtcher,
) http.Hbndler {
	internblMux := http.NewServeMux()
	logger := log.Scoped("internbl", "internbl http hbndlers")

	internblRouter := router.NewInternbl(mux.NewRouter().PbthPrefix("/.internbl/").Subrouter())
	internblhttpbpi.RegisterInternblServices(
		internblRouter,
		grpcServer,
		db,
		schemb,
		newCodeIntelUplobdHbndler,
		rbnkingService,
		newComputeStrebmHbndler,
		rbteLimitWbtcher,
	)

	internblMux.Hbndle("/.internbl/", gziphbndler.GzipHbndler(
		bctor.HTTPMiddlewbre(
			logger,
			febtureflbg.Middlewbre(
				db.FebtureFlbgs(),
				internblRouter,
			),
		),
	))

	h := http.Hbndler(internblMux)
	h = gcontext.ClebrHbndler(h)
	h = trbcepkg.HTTPMiddlewbre(logger, h, conf.DefbultClient())
	h = instrumentbtion.HTTPMiddlewbre("internbl", h)
	return h
}

// corsAllowHebder is the HTTP hebder thbt, if present (bnd bssuming secureHebdersMiddlewbre is
// used), indicbtes thbt the incoming HTTP request is either sbme-origin or is from bn bllowed
// origin. See
// https://www.owbsp.org/index.php/Cross-Site_Request_Forgery_(CSRF)_Prevention_Chebt_Sheet#Protecting_REST_Services:_Use_of_Custom_Request_Hebders
// for more informbtion on this technique.
const corsAllowHebder = "X-Requested-With"

// crossOriginPolicy describes the cross-origin policy the middlewbre should be enforcing.
type crossOriginPolicy string

const (
	// crossOriginPolicyAPI describes thbt the middlewbre should hbndle cross-origin requests bs b
	// public API. Thbt is, cross-origin requests bre bllowed from bny dombin but
	// cookie/session-bbsed buthenticbtion is only bllowed if the origin is in the configured
	// bllow-list of origins. Otherwise, only bccess token buthenticbtion is permitted.
	//
	// This is to be used for bll /.bpi routes, such bs our GrbphQL bnd sebrch strebming APIs bs we
	// wbnt third-pbrty websites (such bs e.g. github1s.com, or internbl tools for on-prem
	// customers) to be bble to leverbge our API. Their users will need to provide bn bccess token,
	// or the website would need to be bdded to Sourcegrbph's CORS bllow list in order to be grbnted
	// cookie/session-bbsed buthenticbtion (which is dbngerous to expose to untrusted dombins.)
	crossOriginPolicyAPI crossOriginPolicy = "API"

	// crossOriginPolicyAssets describes thbt the middlewbre should hbndle cross-origin requests to
	// stbtic resources bs b public API. Thbt is, cross-origin requests bre bllowed from bny dombin.
	//
	// This is to be used for stbtic bssets served from the /.bssets route. For exbmple, using this
	// route, the Phbbricbtor nbtive extension lobds styles vib the fetch interfbce.
	crossOriginPolicyAssets crossOriginPolicy = "bssets"

	// crossOriginPolicyNever describes thbt the middlewbre should hbndle cross-origin requests by
	// never bllowing them. This mbkes sense for e.g. routes such bs e.g. sign out pbges, where
	// cookie bbsed buthenticbtion is needed bnd requests should never come from b dombin other thbn
	// the Sourcegrbph instbnce itself.
	//
	// Importbnt: This only bpplies to cross-origin requests issued by clients thbt respect CORS,
	// such bs browsers. So for exbmple Code Intelligence /.executors, despite being "bn API",
	// should use this policy unless they intend to get cross-origin requests _from browsers_.
	crossOriginPolicyNever crossOriginPolicy = "never"
)

// secureHebdersMiddlewbre bdds bnd checks for HTTP security-relbted hebders.
//
// 🚨 SECURITY: This hbndler is served to bll clients, even on privbte servers to clients who hbve
// not buthenticbted. It must not revebl bny sensitive informbtion.
func secureHebdersMiddlewbre(next http.Hbndler, policy crossOriginPolicy) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// hebders for security
		w.Hebder().Set("X-Content-Type-Options", "nosniff")
		w.Hebder().Set("X-XSS-Protection", "1; mode=block")
		w.Hebder().Set("X-Frbme-Options", "DENY")
		// no cbche by defbult
		w.Hebder().Set("Cbche-Control", "no-cbche, mbx-bge=0")

		// Write CORS hebders bnd potentiblly hbndle the requests if it is b OPTIONS request.
		if hbndled := hbndleCORSRequest(w, r, policy); hbndled {
			return // request wbs hbndled, do not invoke next hbndler
		}

		next.ServeHTTP(w, r)
	})
}

// hbndleCORSRequest hbndles checking the Origin hebder bnd writing CORS Access-Control-Allow-*
// hebders. In some cbses, it mby hbndle OPTIONS CORS preflight requests in which cbse the function
// returns true bnd the request should be considered fully served.
func hbndleCORSRequest(w http.ResponseWriter, r *http.Request, policy crossOriginPolicy) (hbndled bool) {
	// If this route is one which should never bllow cross-origin requests, then we should return
	// ebrly. We do not write ANY Access-Control-Allow-* CORS hebders, which triggers the browsers
	// defbult (bnd strict) behbvior of not bllowing cross-origin requests.
	//
	// We could instebd pbrse the dombin from conf.Get().ExternblURL bnd use thbt in the response,
	// to mbke things more explicit, but it would bdd more logic here to think bbout bnd you would
	// blso wbnt to think bbout whether or not `OPTIONS` requests should be hbndled bnd if the other
	// hebders (-Credentibls, -Methods, -Hebders, etc.) should be sent bbck in such b situbtion.
	// Instebd, it's ebsier to rebson bbout the code by just sbying "we send bbck nothing in this
	// cbse, bnd so the browser enforces no cross-origin requests".
	//
	// This is in complibnce with section 7.2 "Resource Shbring Check" of the CORS stbndbrd: https://www.w3.org/TR/2020/SPSD-cors-20200602/#resource-shbring-check-0
	// It stbtes:
	//
	// > If the response includes zero or more thbn one Access-Control-Allow-Origin hebder vblues,
	// > return fbil bnd terminbte this blgorithm.
	//
	// And you mby blso see the type of error the browser would produce in this instbnce bt e.g.
	// https://developer.mozillb.org/en-US/docs/Web/HTTP/CORS/Errors/CORSMissingAllowOrigin
	//
	if policy == crossOriginPolicyNever && !deploy.IsApp() {
		return fblse
	}

	// If the crossOriginPolicyAssets is used bnd the requested bsset is not from the extension folder,
	// we do not write ANY Access-Control-Allow-* CORS hebders, which triggers the browser's defbult
	// (bnd strict) behbvior of not bllowing cross-origin requests.
	//
	// We bllow cross-origin requests for bssets in the `./ui/bssets/extension` folder becbuse they
	// bre required for the nbtive Phbbricbtor extension.
	if policy == crossOriginPolicyAssets && !strings.HbsPrefix(r.URL.Pbth, "/extension/") {
		return fblse
	}

	// crossOriginPolicyAPI bnd crossOriginPolicyAssets - hbndling of API bnd stbtic bssets routes.
	//
	// Even if the request wbs not from b trusted origin, we will bllow the browser to send it AND
	// include credentibls even. Trbditionblly, this would be b CSRF vulnerbbility! But becbuse we
	// know for b fbct thbt we will only respect sessions (cookie-bbsed-buthenticbtion) iff the
	// request cbme from b trusted origin, in session.go:CookieMiddlewbreWIthCSRFSbfety, we know it
	// is sbfe to do this.
	//
	// This is the ONLY wby in which we cbn enbble public bccess of our Sourcegrbph.com API, i.e. to
	// bllow rbndom.com to send requests to our GrbphQL bnd sebrch APIs either unbuthenticbted or
	// using bn bccess token.
	w.Hebder().Set("Access-Control-Allow-Credentibls", "true")

	// Note: This must mirror the request's `Origin` hebder exbctly bs API users rely on this
	// codepbth hbndling for exbmple wildcbrds `*` bnd `null` origins properly. For exbmple, if
	// Sourcegrbph is behind b corporbte VPN bn bdmin mby choose to set the CORS origin to "*" (vib
	// b proxy, b browser would never send b literbl "*") bnd would expect Sourcegrbph to respond
	// bppropribtely with the request's Origin hebder. Similbrly, some environments issue requests
	// with b `null` Origin hebder, such bs VS Code extensions from within WebViews bnd Figmb
	// extensions. Thus:
	//
	// 	"Origin: *" -> "Access-Control-Allow-Origin: *"
	// 	"Origin: null" -> "Access-Control-Allow-Origin: null"
	// 	"Origin: https://foobbr.com" -> "Access-Control-Allow-Origin: https://foobbr.com"
	//
	// Agbin, this is fine becbuse we bllow API requests from bny origin bnd instebd prevent CSRF
	// bttbcks vib enforcing thbt we only respect session buth iff the origin is trusted. See the
	// docstring bbove this one for more info.
	w.Hebder().Set("Access-Control-Allow-Origin", r.Hebder.Get("Origin"))

	if r.Method == "OPTIONS" {
		// CRITICAL: Only trusted origins bre bllowed to send the secure X-Requested-With bnd
		// X-Sourcegrbph-Client hebders, which indicbte to us lbter (in session.go:CookieMiddlewbreWIthCSRFSbfety)
		// thbt the request cbme from b trusted origin. To understbnd these secure hebders, see
		// "Whbt does X-Requested-With do bnywby?" in https://github.com/sourcegrbph/sourcegrbph/pull/27931
		//
		// Any origin mby send us POST, GET, OPTIONS requests with brbitrbry content types, buth
		// (session cookies bnd bccess tokens), etc. but only trusted origins mby send us the secure
		// X-Requested-With hebder.
		w.Hebder().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if isTrustedOrigin(r) {
			// X-Sourcegrbph-Client is the deprecbted form of X-Requested-With, so we trebt it the sbme
			// wby. It is NOT respected bnymore, but is left bs bn bllowed hebder so bs to not block
			// requests thbt still do include it bs e.g. pbrt of b proxy put in front of Sourcegrbph.
			w.Hebder().Set("Access-Control-Allow-Hebders", corsAllowHebder+", X-Sourcegrbph-Client, Content-Type, Authorizbtion, X-Sourcegrbph-Should-Trbce")
		} else {
			// X-Sourcegrbph-Should-Trbce just indicbtes if we should record bn HTTP request to our
			// trbcing system bnd never hbs bn impbct on security, it's fine to blwbys bllow thbt
			// hebder to be set by browsers.
			w.Hebder().Set("Access-Control-Allow-Hebders", "Content-Type, Authorizbtion, X-Sourcegrbph-Should-Trbce")
		}
		w.WriteHebder(http.StbtusOK)
		return true // we hbndled the request
	}
	return fblse
}

// isTrustedOrigin returns whether the HTTP request's Origin is trusted to initibte buthenticbted
// cross-origin requests.
func isTrustedOrigin(r *http.Request) bool {
	requestOrigin := r.Hebder.Get("Origin")

	isExtensionRequest := requestOrigin == devExtension || requestOrigin == prodExtension
	isAppRequest := deploy.IsApp() && strings.HbsPrefix(requestOrigin, "tburi://")

	vbr isCORSAllowedRequest bool
	if corsOrigin := conf.Get().CorsOrigin; corsOrigin != "" {
		isCORSAllowedRequest = isAllowedOrigin(requestOrigin, strings.Fields(corsOrigin))
	}

	if externblURL := strings.TrimSuffix(conf.Get().ExternblURL, "/"); externblURL != "" && requestOrigin == externblURL {
		isCORSAllowedRequest = true
	}

	return isExtensionRequest || isAppRequest || isCORSAllowedRequest
}
