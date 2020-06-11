package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/grafana-tools/sdk"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// generateSiteConfigChecksum generates a checksum of relevant parts of site configuration
func generateSiteConfigChecksum(config *schema.SiteConfiguration) []byte {
	if config.ObservabilityAlerts == nil {
		return nil
	}
	b, err := json.Marshal(config.ObservabilityAlerts)
	if err != nil {
		return nil
	}
	sum := sha256.Sum256(b)
	return sum[:]
}

type siteConfigSubscriber struct {
	log     log15.Logger
	grafana *sdk.Client

	mux       sync.RWMutex
	alerts    []*schema.ObservabilityAlerts // can be any conf.GrafanaNotifierX
	alertsSum []byte
}

func newSiteConfigSubscriber(ctx context.Context, logger log15.Logger, grafana *sdk.Client) (*siteConfigSubscriber, error) {
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
	siteConfig := conf.Get().SiteConfiguration
	sum := generateSiteConfigChecksum(&siteConfig)

	// Set up overview dashboard if it does not exist. We attach alerts to a copy of the
	// default home dashboard, because dashboards provisioned from disk cannot be edited.
	if _, _, err := grafana.GetDashboardByUID(ctx, overviewDashboardUID); err != nil {
		homeBoard, err := getOverviewDashboard()
		if err != nil {
			return nil, fmt.Errorf("failed to generate alerts overview dashboard: %w", err)
		}
		if _, err := grafana.SetDashboard(ctx, *homeBoard, sdk.SetDashboardParams{}); err != nil {
			return nil, fmt.Errorf("failed to set up alerts overview dashboard: %w", err)
		}
	}

	subscriber := &siteConfigSubscriber{log: log, grafana: grafana}
	return subscriber, subscriber.updateGrafanaConfig(ctx, siteConfig.ObservabilityAlerts, sum)
}

func (c *siteConfigSubscriber) Subscribe(ctx context.Context) {
	conf.Watch(func() {
		c.mux.RLock()
		newSiteConfig := conf.Get().SiteConfiguration
		newSum := generateSiteConfigChecksum(&newSiteConfig)
		isUnchanged := bytes.Equal(c.alertsSum, newSum)
		c.mux.RUnlock()

		// ignore irrelevant changes
		if isUnchanged {
			c.log.Debug("config updated contained no relevant changes - ignoring")
			return
		}

		// update configuration
		if err := c.updateGrafanaConfig(ctx, newSiteConfig.ObservabilityAlerts, newSum); err != nil {
			c.log.Error("failed to apply config changes - ignoring update", "error", err)
		}
	})
}

func (c *siteConfigSubscriber) resetSrcAlerts(ctx context.Context) error {
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
func (c *siteConfigSubscriber) updateGrafanaConfig(ctx context.Context, newAlerts []*schema.ObservabilityAlerts, newSum []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.log.Debug("updating grafana configuration")

	// generate new notifiers configuration
	created, err := generateNotifiersConfig(c.alerts, newAlerts)
	if err != nil {
		return err
	}

	// get the general alerts panels in the home dashboard
	homeBoard, err := getOverviewDashboard()
	if err != nil {
		return fmt.Errorf("failed to generate alerts overview dashboard: %w", err)
	}
	var criticalPanel, warningPanel *sdk.Panel
	for _, panel := range homeBoard.Panels {
		var level string
		switch panel.CommonPanel.Title {
		case "Critical alerts by service":
			level, criticalPanel = "critical", panel
		case "Warning alerts by service":
			level, warningPanel = "warning", panel
		default:
			continue
		}
		panel.Alert = newDefaultAlertsPanelAlert(level)
	}
	if criticalPanel == nil || warningPanel == nil {
		return errors.New("failed to find alerts panels")
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
		// register alert in corresponding panel
		var panel *sdk.Panel
		if strings.Contains(alert.UID, "critical") {
			panel = criticalPanel
		} else {
			panel = warningPanel
		}
		panel.Alert.Notifications = append(panel.Alert.Notifications, alert)
	}

	// update board
	_, err = c.grafana.SetDashboard(ctx, *homeBoard, sdk.SetDashboardParams{Overwrite: true})
	if err != nil {
		return fmt.Errorf("failed to update dashboard: %w", err)
	}

	// update state
	c.alerts = newAlerts
	c.alertsSum = newSum
	c.log.Debug("updated grafana configuration")
	return nil
}
