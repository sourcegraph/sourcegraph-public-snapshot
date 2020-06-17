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

// GrafanaChangeResult indicates output from a GrafanaChange as well as follow-up items (ie whether or not the change will require a Grafana restart)
type GrafanaChangeResult struct {
	Problems             conf.Problems
	ShouldRestartGrafana bool
}

// GrafanaChange implements a change to Grafana configuration
type GrafanaChange func(ctx context.Context, log log15.Logger, grafana *sdk.Client, newConfig *subscribedSiteConfig) (result GrafanaChangeResult)

// grafanaChangeNotifiers appliies `observability.alerts` as Grafana notifiers and attaches them to relevant alerts
func grafanaChangeNotifiers(ctx context.Context, log log15.Logger, grafana *sdk.Client, newConfig *subscribedSiteConfig) (result GrafanaChangeResult) {
	// convenience function for creating a prefixed problem
	newProblem := func(err error) *conf.Problem {
		return conf.NewSiteProblem(fmt.Sprintf("observability.alerts: %v", err))
	}

	// generate new notifiers configuration
	created, err := newGrafanaNotifiersConfig(newConfig.Alerts)
	if err != nil {
		result.Problems = append(result.Problems, newProblem(err))
		return
	}

	// get the general alerts panels in the home dashboard
	homeBoard, err := getOverviewDashboard()
	if err != nil {
		result.Problems = append(result.Problems, newProblem(fmt.Errorf("failed to generate alerts overview dashboard: %w", err)))
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
		result.Problems = append(result.Problems, newProblem(errors.New("failed to find alerts panels")))
		return
	}

	if err := resetSrcNotifiers(ctx, grafana); err != nil {
		result.Problems = append(result.Problems, newProblem(err))
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
				newProblem(fmt.Errorf("failed to create alert %q: please refer to the Grafana logs for more details", alert.UID)),
				newProblem(fmt.Errorf("grafana error: %w", err)))
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
		result.Problems = append(result.Problems, newProblem(fmt.Errorf("failed to update dashboard: %w", err)))
		return
	}

	return
}

// grafanaChangeSMTP applies SMTP server configurations to Grafana.
func grafanaChangeSMTP(ctx context.Context, log log15.Logger, grafana *sdk.Client, newConfig *subscribedSiteConfig) (result GrafanaChangeResult) {
	// convenience function for creating a prefixed problem
	newProblem := func(err error) *conf.Problem {
		return conf.NewSiteProblem(fmt.Sprintf("observability (email.smtp): %v", err))
	}

	grafanaConfig, err := getGrafanaConfig()
	if err != nil {
		result.Problems = append(result.Problems, newProblem(err))
		return
	}

	grafanaConfig.DeleteSection("smtp")
	if newConfig.Email != nil && newConfig.Email.SMTP != nil {
		smtpSection, err := grafanaConfig.NewSection("smtp")
		if err != nil {
			result.Problems = append(result.Problems, newProblem(fmt.Errorf("failed to update Grafana config: %w", err)))
			return
		}
		if err := smtpSection.ReflectFrom(newGrafanaSMTPConfig(newConfig.Email)); err != nil {
			result.Problems = append(result.Problems, newProblem(fmt.Errorf("failed to set Grafana config: %w", err)))
			return
		}
	}

	if err := grafanaConfig.SaveTo(grafanaConfigPath); err != nil {
		result.Problems = append(result.Problems, newProblem(fmt.Errorf("failed to save Grafana config: %w", err)))
		return
	}
	result.ShouldRestartGrafana = true
	return
}
