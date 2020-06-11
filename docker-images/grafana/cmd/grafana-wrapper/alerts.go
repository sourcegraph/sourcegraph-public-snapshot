package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/grafana-tools/sdk"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// #nosec G306  grafana runs as UID 472
const fileMode = 0666

func generateAlertsChecksum(alerts []*schema.ObservabilityAlerts) []byte {
	b, err := json.Marshal(alerts)
	if err != nil {
		return nil
	}
	sum := sha256.Sum256(b)
	return sum[:]
}

type configSubscriber struct {
	log     log15.Logger
	grafana *sdk.Client

	mux       sync.RWMutex
	alerts    []*schema.ObservabilityAlerts // can be any conf.GrafanaNotifierX
	alertsSum []byte
}

func newConfigSubscriber(ctx context.Context, logger log15.Logger, grafana *sdk.Client) (*configSubscriber, error) {
	log := logger.New("logger", "config-subscriber")

	// Syncing relies on access to frontend, so wait until it is ready
	log.Info("waiting for frontend", "url", api.InternalClient.URL)
	if err := api.InternalClient.WaitForFrontend(ctx); err != nil {
		return nil, fmt.Errorf("sourcegraph-frontend not reachable: %w", err)
	}
	log.Debug("detected frontend ready")

	// Need grafana to be ready to intialize alerts
	log.Info("waiting for grafana")
	if err := waitForGrafana(ctx, grafana); err != nil {
		return nil, fmt.Errorf("grafana not reachable: %w", err)
	}
	log.Debug("detected grafana ready")

	// Load initial alerts configuration
	alerts := conf.Get().ObservabilityAlerts
	sum := generateAlertsChecksum(alerts)

	subscriber := &configSubscriber{log: log, grafana: grafana}
	return subscriber, subscriber.updateGrafanaConfig(ctx, alerts, sum)
}

func (c *configSubscriber) Subscribe(ctx context.Context) {
	conf.Watch(func() {
		c.mux.RLock()
		newAlerts := conf.Get().ObservabilityAlerts
		newAlertsSum := generateAlertsChecksum(newAlerts)
		isUnchanged := bytes.Equal(c.alertsSum, newAlertsSum)
		c.mux.RUnlock()

		// ignore irrelevant changes
		if isUnchanged {
			c.log.Debug("config updated contained no relevant changes - ignoring")
			return
		}

		// update configuration
		if err := c.updateGrafanaConfig(ctx, newAlerts, newAlertsSum); err != nil {
			c.log.Error("failed to apply config changes - ignoring update", "error", err)
		}
	})
}

func (c *configSubscriber) resetSrcAlerts(ctx context.Context) error {
	alerts, err := c.grafana.GetAllAlertNotifications(ctx)
	if err != nil {
		return err
	}
	for _, alert := range alerts {
		if strings.HasPrefix(alert.UID, "src-") {
			if err := c.grafana.DeleteAlertNotificationUID(ctx, alert.UID); err != nil {
				return fmt.Errorf("failed to delete alert %q: %w", alert.UID, err)
			}
		}
	}
	return nil
}

// updateGrafanaConfig updates grafanaAlertsSubscriber state and writes it to disk
func (c *configSubscriber) updateGrafanaConfig(ctx context.Context, newAlerts []*schema.ObservabilityAlerts, newSum []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.log.Debug("updating grafana configuration")

	// generate new configuration
	created, err := generateNotifiersConfig(c.alerts, newAlerts)
	if err != nil {
		return err
	}

	// TODO: error handling and rollback. how do we propagage warnings back to the frontend?
	// could maybe revert back to c.alerts
	if err := c.resetSrcAlerts(ctx); err != nil {
		return err
	}
	for _, alert := range created {
		_, err = c.grafana.CreateAlertNotification(ctx, alert)
		if err != nil {
			return fmt.Errorf("failed to create alert %q: %w", alert.UID, err)
		}
	}

	// update state
	c.alerts = newAlerts
	c.alertsSum = newSum
	c.log.Debug("updated grafana configuration")
	return nil
}

func newAlertUID(alertType string, alert *schema.ObservabilityAlerts) string {
	return fmt.Sprintf("src-%s-%v-%s", alert.Level, alertType, alert.Id)
}

func generateNotifiersConfig(current []*schema.ObservabilityAlerts, newAlerts []*schema.ObservabilityAlerts) ([]sdk.AlertNotification, error) {
	// generate grafana notifiers
	var newGrafanaAlerts []sdk.AlertNotification
	for _, alert := range newAlerts {
		alertType, fields, err := structToNotifierSettings(alert.Notifier)
		if err != nil {
			return nil, fmt.Errorf("new notifier '%s' is invalid: %w", alert.Id, err)
		}
		uid := newAlertUID(alertType, alert)
		newGrafanaAlerts = append(newGrafanaAlerts, sdk.AlertNotification{
			UID:      uid,
			Name:     alert.Id,
			Type:     alertType,
			Settings: fields,
		})
	}
	return newGrafanaAlerts, nil
}

// structToNotifierSettings marshals the provided notifier and unmarshals it into a map
// that corresponds with Grafana's notifier settings
func structToNotifierSettings(n *schema.Notifier) (string, map[string]interface{}, error) {
	b, err := n.MarshalJSON()
	if err != nil {
		return "", nil, fmt.Errorf("invalid notifier: %w", err)
	}
	var fields map[string]interface{}
	if err := json.Unmarshal(b, &fields); err != nil {
		return "", nil, fmt.Errorf("could not parse notifier fields: %w", err)
	}

	// the notifiers field maps exactly to grafana notifier settings, except for the additional type field
	alertType := fields["type"].(string)
	delete(fields, "type")

	return alertType, fields, nil
}
