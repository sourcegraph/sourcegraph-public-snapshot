package app

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

	"github.com/gorilla/mux"
	"go.uber.org/atomic"

	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/debugproxies"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/otlpadapter"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/otlpenv"
	srcprometheus "github.com/sourcegraph/sourcegraph/internal/src-prometheus"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	grafanaURLFromEnv = env.Get("GRAFANA_SERVER_URL", "", "URL at which Grafana can be reached")
	jaegerURLFromEnv  = env.Get("JAEGER_SERVER_URL", "", "URL at which Jaeger UI can be reached")
)

func init() {
	conf.ContributeWarning(newPrometheusValidator(srcprometheus.NewClient(srcprometheus.PrometheusURL)))
}

func addNoK8sClientHandler(r *mux.Router, db database.DB) {
	noHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `Cluster information not available`)
		fmt.Fprintf(w, `<br><br><a href="headers">headers</a><br>`)
	})
	r.Handle("/", debugproxies.AdminOnly(db, noHandler))
}

// addDebugHandlers registers the reverse proxies to each services debug
// endpoints.
func addDebugHandlers(r *mux.Router, db database.DB) {
	addGrafana(r, db)
	addJaeger(r, db)
	addSentry(r)
	addOpenTelemetryProtocolAdapter(r)

	var rph debugproxies.ReverseProxyHandler

	if len(debugserver.Services) > 0 {
		peps := make([]debugproxies.Endpoint, 0, len(debugserver.Services))
		for _, s := range debugserver.Services {
			peps = append(peps, debugproxies.Endpoint{
				Service: s.Name,
				Addr:    s.Host,
			})
		}
		rph.Populate(db, peps)
	} else if deploy.IsDeployTypeKubernetes(deploy.Type()) {
		err := debugproxies.StartClusterScanner(func(endpoints []debugproxies.Endpoint) {
			rph.Populate(db, endpoints)
		})
		if err != nil {
			// we ended up here because cluster is not a k8s cluster
			addNoK8sClientHandler(r, db)
			return
		}
	} else {
		addNoK8sClientHandler(r, db)
	}

	rph.AddToRouter(r, db) // todo
}

// PreMountGrafanaHook (if set) is invoked as a hook prior to mounting a
// the Grafana endpoint to the debug router.
var PreMountGrafanaHook func() error

// This error is returned if the current license does not support monitoring.
const errMonitoringNotLicensed = `The feature "monitoring" is not activated in your Sourcegraph license. Upgrade your Sourcegraph subscription to use this feature.`

func addNoGrafanaHandler(r *mux.Router, db database.DB) {
	noGrafana := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `Grafana endpoint proxying: Please set env var GRAFANA_SERVER_URL`)
	})
	r.Handle("/grafana", debugproxies.AdminOnly(db, noGrafana))
}

func addGrafanaNotLicensedHandler(r *mux.Router, db database.DB) {
	notLicensed := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, errMonitoringNotLicensed, http.StatusUnauthorized)
	})
	r.Handle("/grafana", debugproxies.AdminOnly(db, notLicensed))
}

// addReverseProxyForService registers a reverse proxy for the specified service.
func addGrafana(r *mux.Router, db database.DB) {
	if PreMountGrafanaHook != nil {
		if err := PreMountGrafanaHook(); err != nil {
			addGrafanaNotLicensedHandler(r, db)
			return
		}
	}
	if len(grafanaURLFromEnv) > 0 {
		grafanaURL, err := url.Parse(grafanaURLFromEnv)
		if err != nil {
			log.Printf("failed to parse GRAFANA_SERVER_URL=%s: %v",
				grafanaURLFromEnv, err)
			addNoGrafanaHandler(r, db)
		} else {
			prefix := "/grafana"
			// 🚨 SECURITY: Only admins have access to Grafana dashboard
			r.PathPrefix(prefix).Handler(debugproxies.AdminOnly(db, &httputil.ReverseProxy{
				Director: func(req *http.Request) {
					// if set, grafana will fail with an authentication error, so don't allow passthrough
					req.Header.Del("Authorization")
					req.URL.Scheme = "http"
					req.URL.Host = grafanaURL.Host
					if i := strings.Index(req.URL.Path, prefix); i >= 0 {
						req.URL.Path = req.URL.Path[i+len(prefix):]
					}
				},
			}))
		}
	} else {
		addNoGrafanaHandler(r, db)
	}
}

// addSentry declares a route for handling tunneled sentry events from the client.
// See https://docs.sentry.io/platforms/javascript/troubleshooting/#dealing-with-ad-blockers.
//
// The route only forwards known project ids, so a DSN must be defined in siteconfig.Log.Sentry.Dsn
// to allow events to be forwarded. Sentry responses are ignored.
func addSentry(r *mux.Router) {
	logger := sglog.Scoped("sentryTunnel")

	// Helper to fetch Sentry configuration from siteConfig.
	getConfig := func() (string, string, error) {
		var sentryDSN string
		siteConfig := conf.Get().SiteConfiguration
		if siteConfig.Log != nil && siteConfig.Log.Sentry != nil && siteConfig.Log.Sentry.Dsn != "" {
			sentryDSN = siteConfig.Log.Sentry.Dsn
		}
		if sentryDSN == "" {
			return "", "", errors.New("no sentry config available in siteconfig")
		}
		u, err := url.Parse(sentryDSN)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("%s://%s", u.Scheme, u.Host), strings.TrimPrefix(u.Path, "/"), nil
	}

	r.HandleFunc("/sentry_tunnel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Read the envelope.
		b, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Warn("failed to read request body", sglog.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Extract the DSN and ProjectID
		n := bytes.IndexByte(b, '\n')
		if n < 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		h := struct {
			DSN string `json:"dsn"`
		}{}
		err = json.Unmarshal(b[0:n], &h)
		if err != nil {
			logger.Warn("failed to parse request body", sglog.Error(err))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		u, err := url.Parse(h.DSN)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		pID := strings.TrimPrefix(u.Path, "/")
		if pID == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		sentryHost, configProjectID, err := getConfig()
		if err != nil {
			logger.Warn("failed to read sentryDSN from siteconfig", sglog.Error(err))
			w.WriteHeader(http.StatusForbidden)
			return
		}
		// hardcoded in client/browser/src/shared/sentry/index.ts
		hardcodedSentryProjectID := "1334031"
		if !(pID == configProjectID || pID == hardcodedSentryProjectID) {
			// not our projects, just discard the request.
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		client := http.Client{
			// We want to keep this short, the default client settings are not strict enough.
			Timeout: 3 * time.Second,
		}
		apiUrl := fmt.Sprintf("%s/api/%s/envelope/", sentryHost, pID)

		// Asynchronously forward to Sentry, there's no need to keep holding this connection
		// opened any longer.
		go func() {
			resp, err := client.Post(apiUrl, "text/plain;charset=UTF-8", bytes.NewReader(b))
			if err != nil || resp.StatusCode >= 400 {
				logger.Warn("failed to forward", sglog.Error(err), sglog.Int("statusCode", resp.StatusCode))
				return
			}
			resp.Body.Close()
		}()

		w.WriteHeader(http.StatusOK)
	})
}

func addNoJaegerHandler(r *mux.Router, db database.DB) {
	noJaeger := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `Jaeger endpoint proxying: Please set env var JAEGER_SERVER_URL`)
	})
	r.Handle("/jaeger", debugproxies.AdminOnly(db, noJaeger))
}

func addJaeger(r *mux.Router, db database.DB) {
	if len(jaegerURLFromEnv) > 0 {
		jaegerURL, err := url.Parse(jaegerURLFromEnv)
		if err != nil {
			log.Printf("failed to parse JAEGER_SERVER_URL=%s: %v", jaegerURLFromEnv, err)
			addNoJaegerHandler(r, db)
		} else {
			prefix := "/jaeger"
			// 🚨 SECURITY: Only admins have access to Jaeger dashboard
			r.PathPrefix(prefix).Handler(debugproxies.AdminOnly(db, &httputil.ReverseProxy{
				Director: func(req *http.Request) {
					req.URL.Scheme = "http"
					req.URL.Host = jaegerURL.Host
				},
			}))
		}

	} else {
		addNoJaegerHandler(r, db)
	}
}

func clientOtelEnabled(s schema.SiteConfiguration) bool {
	if s.ObservabilityClient == nil {
		return false
	}
	if s.ObservabilityClient.OpenTelemetry == nil {
		return false
	}
	return s.ObservabilityClient.OpenTelemetry.Endpoint != ""
}

// addOpenTelemetryProtocolAdapter registers handlers that forward OpenTelemetry protocol
// (OTLP) requests in the http/json format to the configured backend.
func addOpenTelemetryProtocolAdapter(r *mux.Router) {
	var (
		ctx      = context.Background()
		endpoint = otlpenv.GetEndpoint()
		protocol = otlpenv.GetProtocol()
		logger   = sglog.Scoped("otlpAdapter").
				With(sglog.String("endpoint", endpoint), sglog.String("protocol", string(protocol)))
	)

	// Clients can take a while to receive new site configuration - since this debug
	// tunnel should only be receiving OpenTelemetry from clients, if client OTEL is
	// disabled this tunnel should no-op.
	clientEnabled := atomic.NewBool(clientOtelEnabled(conf.SiteConfig()))
	conf.Watch(func() {
		clientEnabled.Store(clientOtelEnabled(conf.SiteConfig()))
	})

	// If no endpoint is configured, we export a no-op handler
	if endpoint == "" {
		logger.Info("no OTLP endpoint configured, data received at /-/debug/otlp will not be exported")

		r.PathPrefix("/otlp").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `OpenTelemetry protocol tunnel: please configure an exporter endpoint with OTEL_EXPORTER_OTLP_ENDPOINT`)
			w.WriteHeader(http.StatusNotFound)
		})
		return
	}

	// Register adapter endpoints
	otlpadapter.Register(ctx, logger, protocol, endpoint, r, clientEnabled)
}

// newPrometheusValidator renders problems with the Prometheus deployment and relevant site configuration
// as reported by `prom-wrapper` inside the `sourcegraph/prometheus` container if Prometheus is enabled.
//
// It also accepts the error from creating `srcprometheus.Client` as an parameter, to validate
// Prometheus configuration.
func newPrometheusValidator(prom srcprometheus.Client, promErr error) conf.Validator {
	return func(c conftypes.SiteConfigQuerier) conf.Problems {
		// surface new prometheus client error if it was unexpected
		prometheusUnavailable := errors.Is(promErr, srcprometheus.ErrPrometheusUnavailable)
		if promErr != nil && !prometheusUnavailable {
			return conf.NewSiteProblems(fmt.Sprintf("Prometheus (`PROMETHEUS_URL`) might be misconfigured: %v", promErr))
		}

		// no need to validate prometheus config if no `observability.*` settings are configured
		observabilityNotConfigured := len(c.SiteConfig().ObservabilityAlerts) == 0 && len(c.SiteConfig().ObservabilitySilenceAlerts) == 0
		if observabilityNotConfigured {
			// no observability configuration, no checks to make
			return nil
		} else if prometheusUnavailable {
			// no prometheus, but observability is configured
			return conf.NewSiteProblems("`observability.alerts` or `observability.silenceAlerts` are configured, but Prometheus is not available")
		}

		// use a short timeout to avoid having this block problems from loading
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		// get reported problems
		status, err := prom.GetConfigStatus(ctx)
		if err != nil {
			return conf.NewSiteProblems(fmt.Sprintf("`observability`: failed to fetch alerting configuration status: %v", err))
		}
		return status.Problems
	}
}
