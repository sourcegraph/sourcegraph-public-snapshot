package notify

import (
	"context"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

func TestSlackRateLimitNotifier(t *testing.T) {
	logger := logtest.NoOp(t)

	tests := []struct {
		name            string
		mockRedis       func(t *testing.T) redispool.KeyValue
		usagePercentage float32
		wantAlerted     bool
	}{
		{
			name:            "no alerts below lowest bucket",
			mockRedis:       func(*testing.T) redispool.KeyValue { return redispool.NewMockKeyValue() },
			usagePercentage: 0.1,
			wantAlerted:     false,
		},
		{
			name: "alert when hits 50% bucket",
			mockRedis: func(*testing.T) redispool.KeyValue {
				rs := redispool.NewMockKeyValue()
				rs.SetNxFunc.SetDefaultReturn(true, nil)
				return rs
			},
			usagePercentage: 0.5,
			wantAlerted:     true,
		},
		{
			name: "no alert when hits alerted bucket",
			mockRedis: func(*testing.T) redispool.KeyValue {
				rs := redispool.NewMockKeyValue()
				rs.SetNxFunc.SetDefaultReturn(true, nil)
				rs.GetFunc.SetDefaultReturn(redispool.NewValue(int64(50), nil))
				return rs
			},
			usagePercentage: 0.6,
			wantAlerted:     false,
		},
		{
			name: "alert when hits another bucket",
			mockRedis: func(*testing.T) redispool.KeyValue {
				rs := redispool.NewMockKeyValue()
				rs.SetNxFunc.SetDefaultReturn(true, nil)
				rs.GetFunc.SetDefaultReturn(redispool.NewValue(int64(50), nil))
				return rs
			},
			usagePercentage: 0.8,
			wantAlerted:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alerted := false
			alerter := NewSlackRateLimitNotifier(
				logger,
				test.mockRedis(t),
				"https://sourcegraph.com/",
				[]int{50, 80, 90},
				"https://hooks.slack.com",
				func(ctx context.Context, url string, msg *slack.WebhookMessage) error {
					alerted = true
					return nil
				},
			)

			alerter("alice", codygateway.ActorSourceProductSubscription, codygateway.FeatureChatCompletions, test.usagePercentage, time.Minute)
			assert.Equal(t, test.wantAlerted, alerted, "alert fired incorrectly")
		})
	}
}
