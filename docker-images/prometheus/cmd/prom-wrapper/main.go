// Command prom-wrapper provides a wrapper command for Prometheus that
// also handles Sourcegraph configuration changes and making changes to Prometheus.
//
// See https://docs-legacy.sourcegraph.com/dev/background-information/observability/prometheus
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	amclient "github.com/prometheus/alertmanager/api/v2/client"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	srcprometheus "github.com/sourcegraph/sourcegraph/internal/src-prometheus"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// prom-wrapper configuration options
var (
	noConfig       = os.Getenv("DISABLE_SOURCEGRAPH_CONFIG")
	noAlertmanager = os.Getenv("DISABLE_ALERTMANAGER")
	exportPort     = env.Get("EXPORT_PORT", "9090", "port that should be used to reverse-proxy Prometheus and custom endpoints externally")

	prometheusPort = env.Get("PROMETHEUS_INTERNAL_PORT", "9092", "internal Prometheus port")

	alertmanagerPort          = env.Get("ALERTMANAGER_INTERNAL_PORT", "9093", "internal Alertmanager port")
	alertmanagerConfigPath    = env.Get("ALERTMANAGER_CONFIG_PATH", "/sg_config_prometheus/alertmanager.yml", "path to alertmanager configuration")
	alertmanagerEnableCluster = env.Get("ALERTMANAGER_ENABLE_CLUSTER", "false", "enable alertmanager clustering")

	opsGenieAPIKey = os.Getenv("OPSGENIE_API_KEY")
)

func main() {
	liblog := log.Init(log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	})
	defer liblog.Sync()

	logger := log.Scoped("prom-wrapper")
	ctx := context.Background()

	disableAlertmanager := noAlertmanager == "true"
	disableSourcegraphConfig := noConfig == "true"
	logger.Info("starting prom-wrapper",
		log.Bool("disableAlertmanager", disableAlertmanager),
		log.Bool("disableSourcegraphConfig", disableSourcegraphConfig))

	// spin up prometheus and alertmanager
	procErrs := make(chan error)
	var promArgs []string
	if len(os.Args) > 1 {
		promArgs = os.Args[1:] // propagate args to prometheus
	}
	go runCmd(log.Scoped("prometheus"), procErrs, NewPrometheusCmd(promArgs, prometheusPort))

	// router serves endpoints accessible from outside the container (defined by `exportPort`)
	// this includes any endpoints from `siteConfigSubscriber`, reverse-proxying services, etc.
	router := mux.NewRouter()

	// alertmanager client
	alertmanager := amclient.NewHTTPClientWithConfig(nil, &amclient.TransportConfig{
		Host:     fmt.Sprintf("127.0.0.1:%s", alertmanagerPort),
		BasePath: fmt.Sprintf("/%s/api/v2", alertmanagerPathPrefix),
		Schemes:  []string{"http"},
	})

	// disable all components that depend on Alertmanager if DISABLE_ALERTMANAGER=true
	if disableAlertmanager {
		logger.Warn("DISABLE_ALERTMANAGER=true; Alertmanager is disabled")
	} else {
		// start alertmanager
		go runCmd(log.Scoped("alertmanager"), procErrs, NewAlertmanagerCmd(alertmanagerConfigPath))

		// wait for alertmanager to become available
		logger.Info("waiting for alertmanager")
		alertmanagerWaitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		if err := waitForAlertmanager(alertmanagerWaitCtx, alertmanager); err != nil {
			logger.Fatal("unable to reach Alertmanager within deadline", log.Error(err))
		}
		cancel()
		logger.Debug("detected alertmanager ready")

		// subscribe to configuration
		if disableSourcegraphConfig {
			logger.Info("DISABLE_SOURCEGRAPH_CONFIG=true; configuration syncing is disabled")
		} else {
			logger.Info("initializing configuration")
			subscriber := NewSiteConfigSubscriber(logger.Scoped("siteconfig"), alertmanager)

			// watch for configuration updates in the background
			go subscriber.Subscribe(ctx)

			// serve subscriber status
			router.PathPrefix(srcprometheus.EndpointConfigSubscriber).Handler(subscriber.Handler())
		}

		// serve alertmanager via reverse proxy
		router.PathPrefix(fmt.Sprintf("/%s", alertmanagerPathPrefix)).Handler(&httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "http"
				req.URL.Host = fmt.Sprintf(":%s", alertmanagerPort)
			},
		})
	}

	// serve alerts summary status
	alertsReporter := NewAlertsStatusReporter(logger, alertmanager)
	router.PathPrefix(srcprometheus.EndpointAlertsStatus).Handler(alertsReporter.Handler())

	// serve prometheus by default via reverse proxy - place last so other prefixes get served first
	router.PathPrefix("/").Handler(&httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = fmt.Sprintf(":%s", prometheusPort)
		},
	})

	go func() {
		logger.Debug("serving endpoints and reverse proxy")
		if err := http.ListenAndServe(fmt.Sprintf(":%s", exportPort), router); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("error serving reverse proxy", log.Error(err))
			os.Exit(1)
		}
		os.Exit(0)
	}()

	// wait until interrupt or error
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	var exitCode int
	select {
	case sig := <-c:
		logger.Info("stopping on signal", log.String("signal", sig.String()))
		exitCode = 2
	case err := <-procErrs:
		if err != nil {
			var e *exec.ExitError
			if errors.As(err, &e) {
				exitCode = e.ProcessState.ExitCode()
			} else {
				exitCode = 1
			}
		} else {
			exitCode = 0
		}
	}
	os.Exit(exitCode)
}
