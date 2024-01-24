package sentryalert

import (
	"encoding/json"
	"testing"

	"github.com/hexops/autogold/v2"
)

func TestActionMarshal(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config Action
		want   autogold.Value
	}{
		{
			name: "Slack Action",
			config: Action{
				ID: SlackNotifyServiceAction,
				ActionParameters: map[string]any{
					"workspace":  12345,
					"channel":    "test-channel",
					"channel_id": 67890,
					"tags":       "test-service",
				},
			},
			want: autogold.Expect(`{"channel":"test-channel","channel_id":67890,"id":"sentry.integrations.slack.notify_action.SlackNotifyServiceAction","tags":"test-service","workspace":12345}`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := json.Marshal(tc.config)
			tc.want.Equal(t, string(got))
		})
	}
}

func TestFilterMarshal(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config Filter
		want   autogold.Value
	}{
		{
			name: "Age Filter",
			config: Filter{
				ID: AgeComparisonFilter,
				FilterParameters: map[string]any{
					"comparison_type": "older",
					"value":           3,
					"time":            "week",
				},
			},
			want: autogold.Expect(`{"comparison_type":"older","id":"sentry.rules.filters.age_comparison.AgeComparisonFilter","time":"week","value":3}`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := json.Marshal(tc.config)
			tc.want.Equal(t, string(got))
		})
	}
}
