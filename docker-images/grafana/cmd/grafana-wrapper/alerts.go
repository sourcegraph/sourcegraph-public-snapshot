package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	"gopkg.in/yaml.v2"
)

var grafanaProvisioningPath = os.Getenv("GF_PATHS_PROVISIONING")
var grafanaProvisioningNotifiersPath = path.Join(grafanaProvisioningPath, "notifiers")
var srcNotifiersPath = path.Join(grafanaProvisioningNotifiersPath, "sourcegraph.yaml")

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
	log log15.Logger
	mux sync.RWMutex

	alerts    []*schema.ObservabilityAlerts // can be any conf.GrafanaNotifierX
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

	// initialize directory for config, in case it does not exist
	if err := os.MkdirAll(grafanaProvisioningNotifiersPath, fileMode); err != nil {
		return nil, fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// Load initial alerts configuration
	alerts := conf.Get().ObservabilityAlerts
	sum := generateAlertsChecksum(alerts)
	subscriber := &configSubscriber{log: log}
	return subscriber, subscriber.updateGrafanaConfig(alerts, sum)
}

func (c *configSubscriber) Subscribe(grafana *grafanaController) {
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
func (c *configSubscriber) updateGrafanaConfig(newAlerts []*schema.ObservabilityAlerts, newSum []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.log.Debug("updating grafana configuration")

	// generate new configuration
	var current notificationsAsConfig
	b, err := ioutil.ReadFile(srcNotifiersPath)
	if err == nil {
		if err := yaml.Unmarshal(b, &current); err != nil {
			return fmt.Errorf("failed to read existing configuration: %w", err)
		}
	}
	notifiersConfig, err := generateNotifiersConfig(&current, newAlerts)
	if err != nil {
		return err
	}
	c.log.Debug("new configuration prepared",
		"notifiers", len(notifiersConfig.Notifications),
		"notifiers_removed", len(notifiersConfig.DeleteNotifications))

	// replace existing configuration
	b, err = yaml.Marshal(&notifiersConfig)
	if err != nil {
		return fmt.Errorf("failed write new configuration: %w", err)
	}
	os.Remove(srcNotifiersPath)
	if err := ioutil.WriteFile(srcNotifiersPath, b, fileMode); err != nil {
		return fmt.Errorf("failed write new configuration: %w", err)
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

func generateNotifiersConfig(current *notificationsAsConfig, newAlerts []*schema.ObservabilityAlerts) (*notificationsAsConfig, error) {
	var notifiersConfig notificationsAsConfig

	// mark existing alerts as deleted
	if current != nil {
		for _, n := range current.Notifications {
			notifiersConfig.DeleteNotifications = append(notifiersConfig.DeleteNotifications, &deleteNotificationConfig{
				Uid:  n.Uid,
				Name: n.Name,
			})
		}
	}

	// generate grafana notifiers
	for _, alert := range newAlerts {
		alertType, fields, err := structToNotifierSettings(alert.Notifier)
		if err != nil {
			return nil, fmt.Errorf("notifier '%s' is invalid: %w", alert.Id, err)
		}
		notifiersConfig.Notifications = append(notifiersConfig.Notifications, &notificationFromConfig{
			Uid:      newAlertUID(alertType, alert),
			Name:     alert.Id,
			Type:     alertType,
			Settings: fields,
		})
	}

	return &notifiersConfig, nil
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
