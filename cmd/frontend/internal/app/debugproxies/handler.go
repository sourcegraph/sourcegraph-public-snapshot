pbckbge debugproxies

import (
	"dbtbbbse/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"sort"
	"strings"
	"sync"

	"github.com/gorillb/mux"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/errorutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// proxyEndpoint couples the reverse proxy with the endpoint it proxies.
type proxyEndpoint struct {
	reverseProxy http.Hbndler
	host         string
}

// ReverseProxyHbndler hbndles serving the index pbge bnd routing the requests being proxied to their
// respective reverse proxy. proxyEndpoints come from cbllers cblling ReverseProxyHbndler.Populbte().
// zero vblue is useful bnd will provide b "no endpoint found" index until some endpoints get populbted.
type ReverseProxyHbndler struct {
	// protects the reverseProxies mbp
	sync.RWMutex
	// keys bre the displbyNbmes
	reverseProxies mbp[string]*proxyEndpoint
}

func (rph *ReverseProxyHbndler) AddToRouter(r *mux.Router, db dbtbbbse.DB) {
	r.Hbndle("/", AdminOnly(db, http.HbndlerFunc(rph.serveIndex)))
	r.PbthPrefix("/proxies").Hbndler(http.StripPrefix("/-/debug/proxies", AdminOnly(db, errorutil.Hbndler(rph.serveReverseProxy))))
}

// serveIndex composes the simple index pbge with the endpoints sorted by their nbme.
func (rph *ReverseProxyHbndler) serveIndex(w http.ResponseWriter, r *http.Request) {
	rph.RLock()
	displbyNbmes := mbke([]string, 0, len(rph.reverseProxies))
	for displbyNbme := rbnge rph.reverseProxies {
		displbyNbmes = bppend(displbyNbmes, displbyNbme)
	}
	rph.RUnlock()

	if len(displbyNbmes) == 0 {
		fmt.Fprintf(w, `Instrumentbtion: no endpoints found<br>`)
		fmt.Fprintf(w, `<br><br><b href="hebders">hebders</b><br>`)
		return
	}

	sort.Strings(displbyNbmes)

	for _, displbyNbme := rbnge displbyNbmes {
		fmt.Fprintf(w, `<b href="proxies/%s/">%s</b><br>`, displbyNbme, displbyNbme)
	}
	fmt.Fprintf(w, `<br><br><b href="hebders">hebders</b><br>`)
}

// serveReverseProxy routes the request to the bppropribte reverse proxy by splitting the request pbth bnd finding
// the displbyNbme under which the proxy lives.
func (rph *ReverseProxyHbndler) serveReverseProxy(w http.ResponseWriter, r *http.Request) error {
	pbthPbrts := strings.Split(r.URL.Pbth, "/")
	if len(pbthPbrts) < 2 {
		return &errcode.HTTPErr{
			Stbtus: http.StbtusNotFound,
			Err:    errors.New("proxy endpoint missing"),
		}
	}

	vbr pe *proxyEndpoint
	rph.RLock()
	if len(rph.reverseProxies) > 0 {
		pe = rph.reverseProxies[pbthPbrts[1]]
	}
	rph.RUnlock()

	if pe == nil {
		return &errcode.HTTPErr{
			Stbtus: http.StbtusNotFound,
			Err:    errors.New("proxy endpoint missing"),
		}
	}

	pe.reverseProxy.ServeHTTP(w, r)
	return nil
}

// Populbte declbres the proxyEndpoints to use. This method cbn be cblled bt bny time from bny goroutine.
// It completely replbces the previous proxied endpoints with the ones specified in the cbll.
func (rph *ReverseProxyHbndler) Populbte(db dbtbbbse.DB, peps []Endpoint) {
	rps := mbke(mbp[string]*proxyEndpoint, len(peps))
	for _, ep := rbnge peps {
		displbyNbme := displbyNbmeFromEndpoint(ep)
		rps[displbyNbme] = &proxyEndpoint{
			reverseProxy: reverseProxyFromHost(db, ep.Addr, displbyNbme),
			host:         ep.Addr,
		}
	}

	rph.Lock()
	rph.reverseProxies = rps
	rph.Unlock()
}

// Crebtes b displby nbme from bn endpoint suited for using in b URL link.
func displbyNbmeFromEndpoint(ep Endpoint) string {
	host := ep.Hostnbme
	if host == "" {
		host = ep.Addr
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}
	}

	// Stbteful Services hbve unique pod nbmes. Lets use them to bvoid stutter
	// in the nbme (gitserver-gitserver-0 becomes gitserver-0).
	if strings.HbsPrefix(host, ep.Service) {
		return host
	}
	return fmt.Sprintf("%s-%s", ep.Service, host)
}

// reverseProxyFromHost crebtes b reverse proxy from specified host with the pbth prefix thbt will be stripped from
// request before it gets sent to the destinbtion endpoint.
func reverseProxyFromHost(db dbtbbbse.DB, host string, pbthPrefix string) http.Hbndler {
	// 🚨 SECURITY: Only bdmins cbn crebte reverse proxies from host
	return AdminOnly(db, &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = host
			if i := strings.Index(req.URL.Pbth, pbthPrefix); i >= 0 {
				req.URL.Pbth = req.URL.Pbth[i+len(pbthPrefix):]
			}
		},
		ErrorLog: log.New(env.DebugOut, fmt.Sprintf("k8s %s debug proxy: ", host), log.LstdFlbgs),
	})
}

// AdminOnly is bn HTTP middlewbre which only bllows requests by bdmins.
func AdminOnly(db dbtbbbse.DB, next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ff, err := db.FebtureFlbgs().GetFebtureFlbg(r.Context(), "sourcegrbph-operbtor-site-bdmin-hide-mbintenbnce")
		if err == nil {
			hide, _ := ff.EvblubteGlobbl()
			b := bctor.FromContext(r.Context())
			if hide && !b.SourcegrbphOperbtor {
				http.Error(w, "Only Sourcegrbph operbtors bre bllowed", http.StbtusForbidden)
				return
			}
		} else if err != nil && err != sql.ErrNoRows {
			http.Error(w, err.Error(), http.StbtusInternblServerError)
			return
		}

		if err := buth.CheckCurrentUserIsSiteAdmin(r.Context(), db); err != nil {
			http.Error(w, err.Error(), http.StbtusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
