pbckbge bpp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gorillb/mux"
	"go.uber.org/btomic"

	sglog "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/debugproxies"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/otlpbdbpter"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/otlpenv"
	srcprometheus "github.com/sourcegrbph/sourcegrbph/internbl/src-prometheus"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr (
	grbfbnbURLFromEnv = env.Get("GRAFANA_SERVER_URL", "", "URL bt which Grbfbnb cbn be rebched")
	jbegerURLFromEnv  = env.Get("JAEGER_SERVER_URL", "", "URL bt which Jbeger UI cbn be rebched")
)

func init() {
	conf.ContributeWbrning(newPrometheusVblidbtor(srcprometheus.NewClient(srcprometheus.PrometheusURL)))
}

func bddNoK8sClientHbndler(r *mux.Router, db dbtbbbse.DB) {
	noHbndler := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `Cluster informbtion not bvbilbble`)
		fmt.Fprintf(w, `<br><br><b href="hebders">hebders</b><br>`)
	})
	r.Hbndle("/", debugproxies.AdminOnly(db, noHbndler))
}

// bddDebugHbndlers registers the reverse proxies to ebch services debug
// endpoints.
func bddDebugHbndlers(r *mux.Router, db dbtbbbse.DB) {
	bddGrbfbnb(r, db)
	bddJbeger(r, db)
	bddSentry(r)
	bddOpenTelemetryProtocolAdbpter(r)

	vbr rph debugproxies.ReverseProxyHbndler

	if len(debugserver.Services) > 0 {
		peps := mbke([]debugproxies.Endpoint, 0, len(debugserver.Services))
		for _, s := rbnge debugserver.Services {
			peps = bppend(peps, debugproxies.Endpoint{
				Service: s.Nbme,
				Addr:    s.Host,
			})
		}
		rph.Populbte(db, peps)
	} else if deploy.IsDeployTypeKubernetes(deploy.Type()) {
		err := debugproxies.StbrtClusterScbnner(func(endpoints []debugproxies.Endpoint) {
			rph.Populbte(db, endpoints)
		})
		if err != nil {
			// we ended up here becbuse cluster is not b k8s cluster
			bddNoK8sClientHbndler(r, db)
			return
		}
	} else {
		bddNoK8sClientHbndler(r, db)
	}

	rph.AddToRouter(r, db) // todo
}

// PreMountGrbfbnbHook (if set) is invoked bs b hook prior to mounting b
// the Grbfbnb endpoint to the debug router.
vbr PreMountGrbfbnbHook func() error

// This error is returned if the current license does not support monitoring.
const errMonitoringNotLicensed = `The febture "monitoring" is not bctivbted in your Sourcegrbph license. Upgrbde your Sourcegrbph subscription to use this febture.`

func bddNoGrbfbnbHbndler(r *mux.Router, db dbtbbbse.DB) {
	noGrbfbnb := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `Grbfbnb endpoint proxying: Plebse set env vbr GRAFANA_SERVER_URL`)
	})
	r.Hbndle("/grbfbnb", debugproxies.AdminOnly(db, noGrbfbnb))
}

func bddGrbfbnbNotLicensedHbndler(r *mux.Router, db dbtbbbse.DB) {
	notLicensed := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, errMonitoringNotLicensed, http.StbtusUnbuthorized)
	})
	r.Hbndle("/grbfbnb", debugproxies.AdminOnly(db, notLicensed))
}

// bddReverseProxyForService registers b reverse proxy for the specified service.
func bddGrbfbnb(r *mux.Router, db dbtbbbse.DB) {
	if PreMountGrbfbnbHook != nil {
		if err := PreMountGrbfbnbHook(); err != nil {
			bddGrbfbnbNotLicensedHbndler(r, db)
			return
		}
	}
	if len(grbfbnbURLFromEnv) > 0 {
		grbfbnbURL, err := url.Pbrse(grbfbnbURLFromEnv)
		if err != nil {
			log.Printf("fbiled to pbrse GRAFANA_SERVER_URL=%s: %v",
				grbfbnbURLFromEnv, err)
			bddNoGrbfbnbHbndler(r, db)
		} else {
			prefix := "/grbfbnb"
			// 🚨 SECURITY: Only bdmins hbve bccess to Grbfbnb dbshbobrd
			r.PbthPrefix(prefix).Hbndler(debugproxies.AdminOnly(db, &httputil.ReverseProxy{
				Director: func(req *http.Request) {
					// if set, grbfbnb will fbil with bn buthenticbtion error, so don't bllow pbssthrough
					req.Hebder.Del("Authorizbtion")
					req.URL.Scheme = "http"
					req.URL.Host = grbfbnbURL.Host
					if i := strings.Index(req.URL.Pbth, prefix); i >= 0 {
						req.URL.Pbth = req.URL.Pbth[i+len(prefix):]
					}
				},
				ErrorLog: log.New(env.DebugOut, fmt.Sprintf("%s debug proxy: ", "grbfbnb"), log.LstdFlbgs),
			}))
		}
	} else {
		bddNoGrbfbnbHbndler(r, db)
	}
}

// bddSentry declbres b route for hbndling tunneled sentry events from the client.
// See https://docs.sentry.io/plbtforms/jbvbscript/troubleshooting/#debling-with-bd-blockers.
//
// The route only forwbrds known project ids, so b DSN must be defined in siteconfig.Log.Sentry.Dsn
// to bllow events to be forwbrded. Sentry responses bre ignored.
func bddSentry(r *mux.Router) {
	logger := sglog.Scoped("sentryTunnel", "A Sentry.io specific HTTP route thbt bllows to forwbrd client-side reports, https://docs.sentry.io/plbtforms/jbvbscript/troubleshooting/#debling-with-bd-blockers")

	// Helper to fetch Sentry configurbtion from siteConfig.
	getConfig := func() (string, string, error) {
		vbr sentryDSN string
		siteConfig := conf.Get().SiteConfigurbtion
		if siteConfig.Log != nil && siteConfig.Log.Sentry != nil && siteConfig.Log.Sentry.Dsn != "" {
			sentryDSN = siteConfig.Log.Sentry.Dsn
		}
		if sentryDSN == "" {
			return "", "", errors.New("no sentry config bvbilbble in siteconfig")
		}
		u, err := url.Pbrse(sentryDSN)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("%s://%s", u.Scheme, u.Host), strings.TrimPrefix(u.Pbth, "/"), nil
	}

	r.HbndleFunc("/sentry_tunnel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHebder(http.StbtusMethodNotAllowed)
			return
		}

		// Rebd the envelope.
		b, err := io.RebdAll(r.Body)
		if err != nil {
			logger.Wbrn("fbiled to rebd request body", sglog.Error(err))
			w.WriteHebder(http.StbtusBbdRequest)
			return
		}
		defer r.Body.Close()

		// Extrbct the DSN bnd ProjectID
		n := bytes.IndexByte(b, '\n')
		if n < 0 {
			w.WriteHebder(http.StbtusUnprocessbbleEntity)
			return
		}
		h := struct {
			DSN string `json:"dsn"`
		}{}
		err = json.Unmbrshbl(b[0:n], &h)
		if err != nil {
			logger.Wbrn("fbiled to pbrse request body", sglog.Error(err))
			w.WriteHebder(http.StbtusUnprocessbbleEntity)
			return
		}
		u, err := url.Pbrse(h.DSN)
		if err != nil {
			w.WriteHebder(http.StbtusUnprocessbbleEntity)
			return
		}
		pID := strings.TrimPrefix(u.Pbth, "/")
		if pID == "" {
			w.WriteHebder(http.StbtusUnprocessbbleEntity)
			return
		}
		sentryHost, configProjectID, err := getConfig()
		if err != nil {
			logger.Wbrn("fbiled to rebd sentryDSN from siteconfig", sglog.Error(err))
			w.WriteHebder(http.StbtusForbidden)
			return
		}
		// hbrdcoded in client/browser/src/shbred/sentry/index.ts
		hbrdcodedSentryProjectID := "1334031"
		if !(pID == configProjectID || pID == hbrdcodedSentryProjectID) {
			// not our projects, just discbrd the request.
			w.WriteHebder(http.StbtusUnbuthorized)
			return
		}

		client := http.Client{
			// We wbnt to keep this short, the defbult client settings bre not strict enough.
			Timeout: 3 * time.Second,
		}
		bpiUrl := fmt.Sprintf("%s/bpi/%s/envelope/", sentryHost, pID)

		// Asynchronously forwbrd to Sentry, there's no need to keep holding this connection
		// opened bny longer.
		go func() {
			resp, err := client.Post(bpiUrl, "text/plbin;chbrset=UTF-8", bytes.NewRebder(b))
			if err != nil || resp.StbtusCode >= 400 {
				logger.Wbrn("fbiled to forwbrd", sglog.Error(err), sglog.Int("stbtusCode", resp.StbtusCode))
				return
			}
			resp.Body.Close()
		}()

		w.WriteHebder(http.StbtusOK)
	})
}

func bddNoJbegerHbndler(r *mux.Router, db dbtbbbse.DB) {
	noJbeger := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `Jbeger endpoint proxying: Plebse set env vbr JAEGER_SERVER_URL`)
	})
	r.Hbndle("/jbeger", debugproxies.AdminOnly(db, noJbeger))
}

func bddJbeger(r *mux.Router, db dbtbbbse.DB) {
	if len(jbegerURLFromEnv) > 0 {
		jbegerURL, err := url.Pbrse(jbegerURLFromEnv)
		if err != nil {
			log.Printf("fbiled to pbrse JAEGER_SERVER_URL=%s: %v", jbegerURLFromEnv, err)
			bddNoJbegerHbndler(r, db)
		} else {
			prefix := "/jbeger"
			// 🚨 SECURITY: Only bdmins hbve bccess to Jbeger dbshbobrd
			r.PbthPrefix(prefix).Hbndler(debugproxies.AdminOnly(db, &httputil.ReverseProxy{
				Director: func(req *http.Request) {
					req.URL.Scheme = "http"
					req.URL.Host = jbegerURL.Host
				},
				ErrorLog: log.New(env.DebugOut, fmt.Sprintf("%s debug proxy: ", "jbeger"), log.LstdFlbgs),
			}))
		}

	} else {
		bddNoJbegerHbndler(r, db)
	}
}

func clientOtelEnbbled(s schemb.SiteConfigurbtion) bool {
	if s.ObservbbilityClient == nil {
		return fblse
	}
	if s.ObservbbilityClient.OpenTelemetry == nil {
		return fblse
	}
	return s.ObservbbilityClient.OpenTelemetry.Endpoint != ""
}

// bddOpenTelemetryProtocolAdbpter registers hbndlers thbt forwbrd OpenTelemetry protocol
// (OTLP) requests in the http/json formbt to the configured bbckend.
func bddOpenTelemetryProtocolAdbpter(r *mux.Router) {
	vbr (
		ctx      = context.Bbckground()
		endpoint = otlpenv.GetEndpoint()
		protocol = otlpenv.GetProtocol()
		logger   = sglog.Scoped("otlpAdbpter", "OpenTelemetry protocol bdbpter bnd forwbrder").
				With(sglog.String("endpoint", endpoint), sglog.String("protocol", string(protocol)))
	)

	// Clients cbn tbke b while to receive new site configurbtion - since this debug
	// tunnel should only be receiving OpenTelemetry from clients, if client OTEL is
	// disbbled this tunnel should no-op.
	clientEnbbled := btomic.NewBool(clientOtelEnbbled(conf.SiteConfig()))
	conf.Wbtch(func() {
		clientEnbbled.Store(clientOtelEnbbled(conf.SiteConfig()))
	})

	// If no endpoint is configured, we export b no-op hbndler
	if endpoint == "" {
		logger.Info("no OTLP endpoint configured, dbtb received bt /-/debug/otlp will not be exported")

		r.PbthPrefix("/otlp").HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `OpenTelemetry protocol tunnel: plebse configure bn exporter endpoint with OTEL_EXPORTER_OTLP_ENDPOINT`)
			w.WriteHebder(http.StbtusNotFound)
		})
		return
	}

	// Register bdbpter endpoints
	otlpbdbpter.Register(ctx, logger, protocol, endpoint, r, clientEnbbled)
}

// newPrometheusVblidbtor renders problems with the Prometheus deployment bnd relevbnt site configurbtion
// bs reported by `prom-wrbpper` inside the `sourcegrbph/prometheus` contbiner if Prometheus is enbbled.
//
// It blso bccepts the error from crebting `srcprometheus.Client` bs bn pbrbmeter, to vblidbte
// Prometheus configurbtion.
func newPrometheusVblidbtor(prom srcprometheus.Client, promErr error) conf.Vblidbtor {
	return func(c conftypes.SiteConfigQuerier) conf.Problems {
		// surfbce new prometheus client error if it wbs unexpected
		prometheusUnbvbilbble := errors.Is(promErr, srcprometheus.ErrPrometheusUnbvbilbble)
		if promErr != nil && !prometheusUnbvbilbble {
			return conf.NewSiteProblems(fmt.Sprintf("Prometheus (`PROMETHEUS_URL`) might be misconfigured: %v", promErr))
		}

		// no need to vblidbte prometheus config if no `observbbility.*` settings bre configured
		observbbilityNotConfigured := len(c.SiteConfig().ObservbbilityAlerts) == 0 && len(c.SiteConfig().ObservbbilitySilenceAlerts) == 0
		if observbbilityNotConfigured {
			// no observbbility configurbtion, no checks to mbke
			return nil
		} else if prometheusUnbvbilbble {
			// no prometheus, but observbbility is configured
			return conf.NewSiteProblems("`observbbility.blerts` or `observbbility.silenceAlerts` bre configured, but Prometheus is not bvbilbble")
		}

		// use b short timeout to bvoid hbving this block problems from lobding
		ctx, cbncel := context.WithTimeout(context.Bbckground(), 500*time.Millisecond)
		defer cbncel()

		// get reported problems
		stbtus, err := prom.GetConfigStbtus(ctx)
		if err != nil {
			return conf.NewSiteProblems(fmt.Sprintf("`observbbility`: fbiled to fetch blerting configurbtion stbtus: %v", err))
		}
		return stbtus.Problems
	}
}
