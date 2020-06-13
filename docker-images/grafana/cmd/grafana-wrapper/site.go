package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	alerts    []*schema.ObservabilityAlerts
	alertsSum []byte
	problems  conf.Problems // exported by handler
}

func newSiteConfigSubscriber(ctx context.Context, logger log15.Logger, grafana *sdk.Client) (*siteConfigSubscriber, error) {
	log := logger.New("logger", "config-subscriber")

	// Syncing relies on access to frontend, so wait until it is ready
	log.Info("waiting for frontend", "url", api.InternalClient.URL)
	if err := api.InternalClient.WaitForFrontend(ctx); err != nil {
		return nil, err
	}
	log.Debug("detected frontend ready")

	// Need grafana to be ready to intialize alerts
	log.Info("waiting for grafana")
	if err := waitForGrafana(ctx, grafana); err != nil {
		return nil, err
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
	subscriber.updateGrafanaConfig(ctx, siteConfig.ObservabilityAlerts, sum)
	return subscriber, nil
}

func (c *siteConfigSubscriber) Handler() http.Handler {
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		c.mux.RLock()
		defer c.mux.RUnlock()

		b, err := json.Marshal(map[string]interface{}{
			"problems": c.problems,
		})
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		_, _ = w.Write(b)
	})
	return handler
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
		c.updateGrafanaConfig(ctx, newSiteConfig.ObservabilityAlerts, newSum)
	})
}

// resetSrcNotifiers deletes all alert notifiers in Grafana's DB starting with the UID `"src-"`
func (c *siteConfigSubscriber) resetSrcNotifiers(ctx context.Context) error {
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

// updateGrafanaConfig updates grafanaAlertsSubscriber state and writes it to disk. It never returns an error,
// instead all errors are reported as problems
func (c *siteConfigSubscriber) updateGrafanaConfig(ctx context.Context, newAlerts []*schema.ObservabilityAlerts, newSum []byte) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.log.Debug("updating grafana configuration")

	var problems conf.Problems

	// generate new notifiers configuration
	created, err := generateNotifiersConfig(c.alerts, newAlerts)
	if err != nil {
		problems = append(problems, newObservabilityAlertsProblem(err))
		return
	}

	// get the general alerts panels in the home dashboard
	homeBoard, err := getOverviewDashboard()
	if err != nil {
		problems = append(problems, newObservabilityAlertsProblem(fmt.Errorf("failed to generate alerts overview dashboard: %w", err)))
		return
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
		problems = append(problems, newObservabilityAlertsProblem(errors.New("failed to find alerts panels")))
		return
	}

	if err := c.resetSrcNotifiers(ctx); err != nil {
		problems = append(problems, newObservabilityAlertsProblem(err))
		// silently try to recreate alerts, in case any were deleted
		c.log.Warn("failed to reset notifiers - attempting to recreate")
		for _, alert := range created {
			if _, err := c.grafana.CreateAlertNotification(ctx, alert); err != nil {
				c.log.Warn(fmt.Sprintf("failed to recreate notifier %q", alert.UID), "error", err)
			}
		}
		return
	}
	for _, alert := range created {
		_, err = c.grafana.CreateAlertNotification(ctx, alert)
		if err != nil {
			c.log.Error(fmt.Sprintf("failed to create notifier %q", alert.UID), "error", err)
			problems = append(problems, newObservabilityAlertsProblem(fmt.Errorf("failed to create alert %q: please refer to the Grafana logs for more details", alert.UID)))
			problems = append(problems, newObservabilityAlertsProblem(fmt.Errorf("grafana error:", err)))
			continue
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
		problems = append(problems, newObservabilityAlertsProblem(fmt.Errorf("failed to update dashboard: %w", err)))
		return
	}

	// update state
	c.alerts = newAlerts
	c.alertsSum = newSum
	c.problems = problems
	c.log.Debug("updated grafana configuration")
}

func newObservabilityAlertsProblem(err error) *conf.Problem {
	return conf.NewSiteProblem(fmt.Sprintf("observability.alerts: %v", err))
}
