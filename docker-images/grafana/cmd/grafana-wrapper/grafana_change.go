package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana-tools/sdk"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

type GrafanaChangeResult struct {
	Problems             conf.Problems
	ShouldRestartGrafana bool
}

type GrafanaChange func(ctx context.Context, log log15.Logger, grafana *sdk.Client, oldConfig, newConfig *subscribedSiteConfig) (result GrafanaChangeResult)

func grafanaChangeNotifiers(ctx context.Context, log log15.Logger, grafana *sdk.Client, oldConfig, newConfig *subscribedSiteConfig) (result GrafanaChangeResult) {
	// resetSrcNotifiers deletes all alert notifiers in Grafana's DB starting with the UID `"src-"`
	resetSrcNotifiers := func(ctx context.Context, grafana *sdk.Client) error {
		alerts, err := grafana.GetAllAlertNotifications(ctx)
		if err != nil {
			return err
		}
		for _, alert := range alerts {
			if strings.HasPrefix(alert.UID, "src-") {
				if err := grafana.DeleteAlertNotificationUID(ctx, alert.UID); err != nil {
					return fmt.Errorf("failed to delete alert %q: %w", alert.UID, err)
				}
			}
		}
		return nil
	}

	newObservabilityAlertsProblem := func(err error) *conf.Problem {
		return conf.NewSiteProblem(fmt.Sprintf("observability.alerts: %v", err))
	}

	// generate new notifiers configuration
	created, err := generateNotifiersConfig(newConfig.Alerts)
	if err != nil {
		result.Problems = append(result.Problems, newObservabilityAlertsProblem(err))
		return
	}

	// get the general alerts panels in the home dashboard
	homeBoard, err := getOverviewDashboard()
	if err != nil {
		result.Problems = append(result.Problems, newObservabilityAlertsProblem(fmt.Errorf("failed to generate alerts overview dashboard: %w", err)))
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
		result.Problems = append(result.Problems, newObservabilityAlertsProblem(errors.New("failed to find alerts panels")))
		return
	}

	if err := resetSrcNotifiers(ctx, grafana); err != nil {
		result.Problems = append(result.Problems, newObservabilityAlertsProblem(err))
		// silently try to recreate alerts, in case any were deleted
		log.Warn("failed to reset notifiers - attempting to recreate")
		for _, alert := range created {
			if _, err := grafana.CreateAlertNotification(ctx, alert); err != nil {
				log.Warn(fmt.Sprintf("failed to recreate notifier %q", alert.UID), "error", err)
			}
		}
		return
	}
	for _, alert := range created {
		_, err = grafana.CreateAlertNotification(ctx, alert)
		if err != nil {
			log.Error(fmt.Sprintf("failed to create notifier %q", alert.UID), "error", err)
			result.Problems = append(result.Problems,
				newObservabilityAlertsProblem(fmt.Errorf("failed to create alert %q: please refer to the Grafana logs for more details", alert.UID)),
				newObservabilityAlertsProblem(fmt.Errorf("grafana error: %w", err)))
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
	_, err = grafana.SetDashboard(ctx, *homeBoard, sdk.SetDashboardParams{Overwrite: true})
	if err != nil {
		result.Problems = append(result.Problems, newObservabilityAlertsProblem(fmt.Errorf("failed to update dashboard: %w", err)))
		return
	}

	return
}

func grafanaChangeSMTP(ctx context.Context, log log15.Logger, grafana *sdk.Client, oldConfig, newConfig *subscribedSiteConfig) (result GrafanaChangeResult) {
	// TODO update SMTP from config
	return
}
