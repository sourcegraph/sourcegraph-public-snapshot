/**
 * Command grafana-wrapper provides a wrapper command for Grafana that
 * also handles Sourcegraph configuration changes.
 */
package main

import (
	"context"
	"os"
	"os/exec"

	"github.com/inconshreveable/log15"
)

var noConfig = os.Getenv("GRAFANA_WRAPPER_NO_CONFIG")

func main() {
	log := log15.New("cmd", "grafana-wrapper")
	ctx := context.Background()

	if noConfig != "true" {
		log.Info("initializing configuration")
		alertsSubscriber, err := newGrafanaAlertsSubscriber(ctx, log)
		if err != nil {
			log.Crit("failed to initialize alerts", "error", err)
			os.Exit(1)
		}
		alertsSubscriber.subscribe()
	} else {
		log.Info("configuration sync disabled")
	}

	log.Info("starting grafana")
	grafanaRun := exec.Command("/run.sh")
	grafanaRun.Env = os.Environ()
	if err := grafanaRun.Run(); err != nil {
		log.Error(err.Error())
	}
}
