package notify

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/slack-go/slack"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

func TestThresholds(t *testing.T) {
	th := Thresholds{
		codygateway.ActorSourceDotcomUser:          []int{100},
		codygateway.ActorSourceProductSubscription: []int{100, 90},
	}
	// Explicitly configured
	autogold.Expect([]int{100}).Equal(t, th.Get(codygateway.ActorSourceDotcomUser))
	// Sorted
	autogold.Expect([]int{90, 100}).Equal(t, th.Get(codygateway.ActorSourceProductSubscription))
	// Defaults
	autogold.Expect([]int{}).Equal(t, th.Get(codygateway.ActorSource("anonymous")))
}

type mockActor struct {
	id     string
	name   string
	source codygateway.ActorSource
}

func (m *mockActor) GetID() string                      { return m.id }
func (m *mockActor) GetName() string                    { return m.name }
func (m *mockActor) GetSource() codygateway.ActorSource { return m.source }

func TestSlackRateLimitNotifier(t *testing.T) {
	logger := logtest.NoOp(t)

	tests := []struct {
		name        string
		mockRedis   func(t *testing.T) redispool.KeyValue
		usageRatio  float32
		wantAlerted bool
	}{
		{
			name:        "no alerts below lowest bucket",
			mockRedis:   func(*testing.T) redispool.KeyValue { return redispool.NewMockKeyValue() },
			usageRatio:  0.1,
			wantAlerted: false,
		},
		{
			name: "alert when hits 50% bucket",
			mockRedis: func(*testing.T) redispool.KeyValue {
				rs := redispool.NewMockKeyValue()
				rs.SetNxFunc.SetDefaultReturn(true, nil)
				return rs
			},
			usageRatio:  0.5,
			wantAlerted: true,
		},
		{
			name: "no alert when hits alerted bucket",
			mockRedis: func(*testing.T) redispool.KeyValue {
				rs := redispool.NewMockKeyValue()
				rs.SetNxFunc.SetDefaultReturn(true, nil)
				rs.GetFunc.SetDefaultReturn(redispool.NewValue(int64(50), nil))
				return rs
			},
			usageRatio:  0.6,
			wantAlerted: false,
		},
		{
			name: "alert when hits another bucket",
			mockRedis: func(*testing.T) redispool.KeyValue {
				rs := redispool.NewMockKeyValue()
				rs.SetNxFunc.SetDefaultReturn(true, nil)
				rs.GetFunc.SetDefaultReturn(redispool.NewValue(int64(50), nil))
				return rs
			},
			usageRatio:  0.8,
			wantAlerted: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alerted := false
			alerter := NewSlackRateLimitNotifier(
				logger,
				test.mockRedis(t),
				"https://sourcegraph.com/",
				Thresholds{codygateway.ActorSourceProductSubscription: []int{50, 80, 90}},
				"https://hooks.slack.com",
				func(ctx context.Context, url string, msg *slack.WebhookMessage) error {
					alerted = true
					return nil
				},
			)

			alerter(context.Background(),
				&mockActor{
					id:     "foobar",
					name:   "alice",
					source: codygateway.ActorSourceProductSubscription,
				},
				codygateway.FeatureChatCompletions,
				test.usageRatio,
				time.Minute)
			assert.Equal(t, test.wantAlerted, alerted, "alert fired incorrectly")
		})
	}
}
