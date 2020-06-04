package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func generateAlertsChecksum(alerts []interface{}) []byte {
	b, err := json.Marshal(alerts)
	if err != nil {
		return nil
	}
	sum := sha256.Sum256(b)
	return sum[:]
}

type grafanaAlertsSubscriber struct {
	log       log15.Logger
	mux       sync.RWMutex
	alerts    []interface{} // can be any conf.GrafanaNotifierX
	alertsSum []byte
}

func newGrafanaAlertsSubscriber(ctx context.Context, log log15.Logger) (*grafanaAlertsSubscriber, error) {
	// Syncing relies on access to frontend, so wait until it is ready
	log.Info("waiting for frontend", "url", api.InternalClient.URL)
	if err := api.InternalClient.WaitForFrontend(ctx); err != nil {
		return nil, fmt.Errorf("sourcegraph-frontend not reachable: %w", err)
	}
	log.Debug("detected frontend ready")

	// Load initial alerts configuration
	alerts := conf.Get().ObservabilityAlerts
	sum := generateAlertsChecksum(alerts)
	subscriber := &grafanaAlertsSubscriber{log: log}
	return subscriber, subscriber.updateGrafana(alerts, sum)
}

func (a *grafanaAlertsSubscriber) subscribe() {
	conf.Watch(func() {
		a.mux.RLock()
		newAlerts := conf.Get().ObservabilityAlerts
		newAlertsSum := generateAlertsChecksum(newAlerts)
		isUnchanged := bytes.Equal(a.alertsSum, newAlertsSum)
		a.mux.RUnlock()
		if isUnchanged {
			return
		}
		if err := a.updateGrafana(newAlerts, newAlertsSum); err != nil {
			a.log.Error("grafana update failed", "error", err)
		}
	})
}

func (a *grafanaAlertsSubscriber) updateGrafana(newAlerts []interface{}, newSum []byte) error {
	a.mux.Lock()
	defer a.mux.Unlock()

	a.alerts = newAlerts
	a.alertsSum = newSum

	// TODO
	a.log.Info("got updated alerts")

	return nil
}
