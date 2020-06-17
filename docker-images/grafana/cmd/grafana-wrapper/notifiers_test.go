package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana-tools/sdk"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestNewGrafanaNotifiersConfig(t *testing.T) {
	type args struct {
		current   []*schema.ObservabilityAlerts
		newAlerts []*schema.ObservabilityAlerts
	}
	tests := []struct {
		name        string
		args        args
		wantCreated []sdk.AlertNotification
		wantErr     bool
	}{
		{
			name: "should convert Sourcegraph observability alert to Grafana notifier",
			args: args{
				newAlerts: []*schema.ObservabilityAlerts{
					{
						Id:    "test-alert",
						Level: "warning",
						Notifier: schema.Notifier{
							Slack: &schema.GrafanaNotifierSlack{
								Type: "slack",
								Url:  "https://soucegraph.com",
							},
						},
					},
				},
			},
			wantCreated: []sdk.AlertNotification{{
				UID:  "src-warning-slack-test-alert",
				Name: "test-alert",
				Type: "slack",
				Settings: map[string]interface{}{
					"url": "https://soucegraph.com",
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created, err := newGrafanaNotifiersConfig(tt.args.newAlerts)
			if (err != nil) != tt.wantErr {
				t.Errorf("newGrafanaNotifiersConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.wantCreated, created); diff != "" {
				t.Errorf("newGrafanaNotifiersConfig() created: %s", diff)
			}
		})
	}
}
