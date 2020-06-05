package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

var grafanaConfigPath = os.Getenv("GF_PATHS_CONFIG")

func generateAlertsChecksum(alerts []interface{}) []byte {
	b, err := json.Marshal(alerts)
	if err != nil {
		return nil
	}
	sum := sha256.Sum256(b)
	return sum[:]
}

type configSubscriber struct {
	log log15.Logger
	mux sync.RWMutex

	alerts    []interface{} // can be any conf.GrafanaNotifierX
	alertsSum []byte
}

func newConfigSubscriber(ctx context.Context, logger log15.Logger) (*configSubscriber, error) {
	log := logger.New("logger", "config-subscriber")
	// Syncing relies on access to frontend, so wait until it is ready
	log.Info("waiting for frontend", "url", api.InternalClient.URL)
	if err := api.InternalClient.WaitForFrontend(ctx); err != nil {
		return nil, fmt.Errorf("sourcegraph-frontend not reachable: %w", err)
	}
	log.Debug("detected frontend ready")

	// Load initial alerts configuration
	alerts := conf.Get().ObservabilityAlerts
	sum := generateAlertsChecksum(alerts)
	subscriber := &configSubscriber{log: log}
	return subscriber, subscriber.updateGrafanaConfig(alerts, sum)
}

func (c *configSubscriber) Subscribe(grafana *grafanaController) {
	conf.Watch(func() {
		// check if update is worth acting on
		c.mux.RLock()
		newAlerts := conf.Get().ObservabilityAlerts
		newAlertsSum := generateAlertsChecksum(newAlerts)
		isUnchanged := bytes.Equal(c.alertsSum, newAlertsSum)
		c.mux.RUnlock()
		if isUnchanged {
			c.log.Debug("config updated contained no relevant changes - ignoring")
			return
		}

		// update configuration
		if err := c.updateGrafanaConfig(newAlerts, newAlertsSum); err != nil {
			c.log.Error("failed to apply config changes - ignoring update", "error", err)
		} else {
			// grafana must be restarted for config changes to apply
			if err := grafana.Restart(); err != nil {
				c.log.Crit("error occured while restarting grafana", "error", err)
			}
		}
	})
}

// updateGrafanaConfig updates grafanaAlertsSubscriber state and writes it to disk
func (c *configSubscriber) updateGrafanaConfig(newAlerts []interface{}, newSum []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.alerts = newAlerts
	c.alertsSum = newSum

	// TODO - write config to disk
	c.log.Info("got updated alerts")

	return nil
}
