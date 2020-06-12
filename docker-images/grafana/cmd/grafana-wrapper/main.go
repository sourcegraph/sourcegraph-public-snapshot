/**
 * Command grafana-wrapper provides a wrapper command for Grafana that
 * also handles Sourcegraph configuration changes.
 */
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/grafana-tools/sdk"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var noConfig = os.Getenv("GRAFANA_WRAPPER_NO_CONFIG")
var grafanaPort = env.Get("GRAFANA_PORT", "3370", "grafana port")
var grafanaCredentials = env.Get("GRAFANA_CREDENTIALS", "admin:admin", "credentials for accessing the grafana server")

func main() {
	log := log15.New("cmd", "grafana-wrapper")
	ctx := context.Background()

	// spin up grafana
	grafanaErrs := make(chan error)
	go func() {
		grafanaErrs <- newGrafanaRunCmd().Run()
	}()

	// subscribe to configuration
	if noConfig == "true" {
		log.Info("configuration sync disabled")
	} else {
		log.Info("initializing configuration")
		grafanaClient := sdk.NewClient(fmt.Sprintf("http://localhost:%s", grafanaPort), grafanaCredentials, http.DefaultClient)

		// limit the amount of time we spend spinning up the subscriber before erroring
		newSubscriberCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		config, err := newSiteConfigSubscriber(newSubscriberCtx, log, grafanaClient)
		if err != nil {
			log.Crit("failed to initialize configuration", "error", err)
			os.Exit(1)
		}
		cancel()

		// watch for configuration updates in the background
		config.Subscribe(ctx)
	}

	err := <-grafanaErrs
	if err != nil {
		log.Crit("grafana exited", "error", err)
		os.Exit(1)
	} else {
		log.Info("grafana exited")
		os.Exit(0)
	}
}
