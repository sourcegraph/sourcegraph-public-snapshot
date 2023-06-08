package shared

import (
	"context"
	"testing"
	"time"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/slack-go/slack"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/actor/productsubscription"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

func TestRateLimitAlerter(t *testing.T) {
	logger := logtest.NoOp(t)
	source := &productsubscription.Source{}

	tests := []struct {
		name            string
		mockRedis       func(t *testing.T) redispool.KeyValue
		usagePercentage float32
		wantAlerted     bool
	}{
		{
			name:            "no alerts below threshold",
			mockRedis:       func(*testing.T) redispool.KeyValue { return redispool.NewMockKeyValue() },
			usagePercentage: 0.1,
			wantAlerted:     false,
		},
		{
			name: "alert above threshold",
			mockRedis: func(*testing.T) redispool.KeyValue {
				rs := redispool.NewMockKeyValue()
				rs.SetNxFunc.SetDefaultReturn(true, nil)
				return rs
			},
			usagePercentage: 0.8,
			wantAlerted:     true,
		},
		{
			name: "no alert during cooldown",
			mockRedis: func(t *testing.T) redispool.KeyValue {
				rs := redispool.NewMockKeyValue()
				rs.SetNxFunc.SetDefaultReturn(true, nil)
				rs.GetFunc.SetDefaultReturn(redispool.NewValue(time.Now().Add(time.Minute).Format(time.RFC3339), nil))

				t.Cleanup(func() {
					mockrequire.Called(t, rs.GetFunc)
				})
				return rs
			},
			usagePercentage: 0.8,
			wantAlerted:     false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alerted := false
			alerter := newRateLimitAlerter(
				logger,
				test.mockRedis(t),
				"https://sourcegraph.com/.api/graphql",
				codygateway.ActorRateLimitAlertConfig{
					Threshold:       0.8,
					Interval:        time.Minute,
					SlackWebhookURL: "https://hooks.slack.com",
				},
				func(ctx context.Context, url string, msg *slack.WebhookMessage) error {
					alerted = true
					return nil
				},
			)

			alerter(&actor.Actor{ID: "alice", Source: source}, codygateway.FeatureChatCompletions, test.usagePercentage)
			assert.Equal(t, test.wantAlerted, alerted, "alert fired incorrectly")
		})
	}
}

func TestTryAcquireRedisLockOnce(t *testing.T) {
	// todo
}
