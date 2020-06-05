/**
 * Command grafana-wrapper provides a wrapper command for Grafana that
 * also handles Sourcegraph configuration changes.
 */
package main

import (
	"context"
	"os"

	"github.com/inconshreveable/log15"
)

var noConfig = os.Getenv("GRAFANA_WRAPPER_NO_CONFIG")

func main() {
	log := log15.New("cmd", "grafana-wrapper")
	ctx := context.Background()

	// controller for running/stopping grafana
	grafana := newGrafanaController(log)

	// subscribe to configuration
	if noConfig != "true" {
		log.Info("initializing configuration")
		config, err := newConfigSubscriber(ctx, log)
		if err != nil {
			log.Crit("failed to initialize configuration", "error", err)
			os.Exit(1)
		}
		config.Subscribe(grafana)
	} else {
		log.Info("configuration sync disabled")
	}

	// initial grafana startup
	if err := grafana.Restart(); err != nil {
		log.Error("failed to start grafana", "error", err)
		os.Exit(1)
	}

	// block
	select {
	case <-ctx.Done():
	}
}
