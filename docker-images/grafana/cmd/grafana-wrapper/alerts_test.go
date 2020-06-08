package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGenerateNotifiersConfig(t *testing.T) {
	type args struct {
		current   *notificationsAsConfig
		newAlerts []*schema.ObservabilityAlerts
	}
	tests := []struct {
		name    string
		args    args
		want    *notificationsAsConfig
		wantErr bool
	}{
		{
			name: "current alerts should be deleted",
			args: args{
				current: &notificationsAsConfig{
					Notifications: []*notificationFromConfig{
						{Uid: "test-alert"},
					},
				},
				newAlerts: nil,
			},
			want: &notificationsAsConfig{
				DeleteNotifications: []*deleteNotificationConfig{
					{Uid: "test-alert"},
				},
			},
		}, {
			name: "should convert Sourcegraph observability alert to Grafana notifier",
			args: args{
				current: nil,
				newAlerts: []*schema.ObservabilityAlerts{
					{
						Id:    "test-alert",
						Level: "warning",
						Notifier: &schema.Notifier{
							Email: &schema.GrafanaNotifierEmail{
								Type:        "email",
								SingleEmail: "robert@bobheadxi.dev",
							},
						},
					},
				},
			},
			want: &notificationsAsConfig{
				Notifications: []*notificationFromConfig{
					{
						Uid:  "src-warning-email-test-alert",
						Name: "test-alert",
						Type: "email",
						Settings: map[string]interface{}{
							"singleEmail": "robert@bobheadxi.dev",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateNotifiersConfig(tt.args.current, tt.args.newAlerts)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateNotifiersConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("generateNotifiersConfig(): %s", diff)
			}
		})
	}
}
