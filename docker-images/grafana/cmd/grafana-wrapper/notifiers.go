package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/grafana-tools/sdk"
	"github.com/sourcegraph/sourcegraph/schema"
)

func newAlertUID(alertType string, alert *schema.ObservabilityAlerts) string {
	return fmt.Sprintf("src-%s-%v-%s", alert.Level, alertType, alert.Id)
}

func newGrafanaNotifiersConfig(newAlerts []*schema.ObservabilityAlerts) ([]sdk.AlertNotification, error) {
	// generate grafana notifiers
	var newGrafanaAlerts []sdk.AlertNotification
	for _, alert := range newAlerts {
		alertType, fields, err := notifierToGrafanaNotifierSettings(&alert.Notifier)
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

// notifierToGrafanaNotifierSettings marshals the provided notifier and unmarshals it into a map
// that corresponds with Grafana's notifier settings
func notifierToGrafanaNotifierSettings(n *schema.Notifier) (string, map[string]interface{}, error) {
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

// resetSrcNotifiers deletes all alert notifiers in Grafana's DB starting with the UID `"src-"`
func resetSrcNotifiers(ctx context.Context, grafana *sdk.Client) error {
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
