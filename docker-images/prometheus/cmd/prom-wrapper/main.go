// Command grafana-wrapper provides a wrapper command for Grafana that
// also handles Sourcegraph configuration changes and making changes to Grafana.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"

	amclient "github.com/prometheus/alertmanager/api/v2/client"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

// prom-wrapper configuration options
var (
	noConfig   = os.Getenv("DISABLE_SOURCEGRAPH_CONFIG")
	exportPort = env.Get("EXPORT_PORT", "9090", "port that should be used to reverse-proxy Prometheus and custom endpoints externally")

	prometheusPort = env.Get("PROMETHEUS_INTERNAL_PORT", "9092", "internal Prometheus port")

	alertmanagerPort       = env.Get("ALERTMANAGER_INTERNAL_PORT", "9093", "internal Alertmanager port")
	alertmanagerConfigPath = env.Get("ALERTMANAGER_CONFIG_PATH", "/sg_config_prometheus/alertmanager.yml", "alertmanager configuration")
)

func main() {
	log := log15.New("cmd", "prom-wrapper")
	ctx := context.Background()

	// spin up prometheus and alertmanager
	procErrs := make(chan error)
	go runCmd(log, procErrs, NewAlertmanagerCmd(alertmanagerConfigPath))
	var promArgs []string
	if len(os.Args) > 1 {
		promArgs = os.Args[1:] // propagate args to prometheus
	}
	go runCmd(log, procErrs, NewPrometheusCmd(promArgs, prometheusPort, exportPort))

	// router serves endpoints accessible from outside the container (defined by `exportPort`)
	// this includes any endpoints from `siteConfigSubscriber`, reverse-proxying Grafana, etc.
	router := mux.NewRouter()

	// subscribe to configuration
	if noConfig == "true" {
		log.Info("DISABLE_SOURCEGRAPH_CONFIG=true; configuration syncing is disabled")
	} else {
		log.Info("initializing configuration")
		alertmanager := amclient.NewHTTPClientWithConfig(nil, &amclient.TransportConfig{
			Host:     fmt.Sprintf("127.0.0.1:%s", alertmanagerPort),
			BasePath: "/alerts/api/v2",
			Schemes:  []string{"http"},
		})

		// limit the amount of time we spend spinning up the subscriber before erroring
		newSubscriberCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		config, err := NewSiteConfigSubscriber(newSubscriberCtx, log, alertmanager)
		if err != nil {
			log.Crit("failed to initialize configuration", "error", err)
			os.Exit(1)
		}
		cancel()

		// watch for configuration updates in the background
		go config.Subscribe(ctx)

		// serve subscriber status
		router.PathPrefix("/prom-wrapper/config-subscriber").Handler(config.Handler())
	}

	// serve alertmanager ui via reverse proxy
	router.PathPrefix("/alerts").Handler(&httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = fmt.Sprintf(":%s", alertmanagerPort)
		},
	})

	// serve prometheus by default via reverse proxy - place last so other prefixes get served first
	router.PathPrefix("/").Handler(&httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = fmt.Sprintf(":%s", prometheusPort)
		},
	})

	go func() {
		log.Debug("serving endpoints and reverse proxy")
		if err := http.ListenAndServe(fmt.Sprintf(":%s", exportPort), router); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Crit("error serving reverse proxy", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	// wait until interrupt or error
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	var exitCode int
	select {
	case sig := <-c:
		log.Info(fmt.Sprintf("stopping on signal %s", sig))
		exitCode = 2
	case err := <-procErrs:
		if err != nil {
			log.Error("subprocess exited", "error", err)
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				exitCode = exitErr.ProcessState.ExitCode()
			} else {
				exitCode = 1
			}
		} else {
			exitCode = 0
		}
	}
	os.Exit(exitCode)
}
