package shared

import (
	"context"
	"fmt"
	"testing"
	"time"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/slack-go/slack"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			alerter := newRateLimitAlerter(
				logger,
				test.mockRedis(t),
				"https://sourcegraph.com/.api/graphql",
				codygateway.ActorRateLimitAlertConfig{
					Thresholds:      []int{50, 80, 90},
					SlackWebhookURL: "https://hooks.slack.com",
				},
				func(ctx context.Context, url string, msg *slack.WebhookMessage) error {
					alerted = true
					return nil
				},
			)

			alerter(&actor.Actor{ID: "alice", Source: source}, codygateway.FeatureChatCompletions, test.usagePercentage, time.Minute)
			assert.Equal(t, test.wantAlerted, alerted, "alert fired incorrectly")
		})
	}
}

func TestTryAcquireRedisLockOnce(t *testing.T) {
	t.Run("acquire and release", func(t *testing.T) {
		rs := redispool.NewMockKeyValue()
		rs.SetNxFunc.SetDefaultReturn(true, nil)
		acquired, release, err := tryAcquireRedisLockOnce(rs, "chicken-dinner", time.Minute)
		require.NoError(t, err)
		require.NotNil(t, release)
		assert.True(t, acquired)
		release()
	})

	t.Run("acquire and give up", func(t *testing.T) {
		rs := redispool.NewMockKeyValue()
		rs.SetNxFunc.PushReturn(true, nil)
		aliceAcquired, aliceRelease, err := tryAcquireRedisLockOnce(rs, "chicken-dinner", time.Minute)
		require.NoError(t, err)
		require.NotNil(t, aliceRelease)
		assert.True(t, aliceAcquired)
		defer aliceRelease()

		rs.SetNxFunc.PushReturn(false, nil)
		rs.GetFunc.PushReturn(redispool.NewValue(fmt.Sprintf("%d,8527", time.Now().Add(time.Minute).UnixNano()), nil))
		bobAcquired, _, err := tryAcquireRedisLockOnce(rs, "chicken-dinner", time.Minute)
		require.NoError(t, err)
		assert.False(t, bobAcquired)
		mockrequire.CalledN(t, rs.SetNxFunc, 2)
		mockrequire.Called(t, rs.GetFunc)
	})

	t.Run("acquire an expired lock", func(t *testing.T) {
		rs := redispool.NewMockKeyValue()
		rs.SetNxFunc.PushReturn(true, nil)
		aliceAcquired, aliceRelease, err := tryAcquireRedisLockOnce(rs, "chicken-dinner", time.Minute)
		require.NoError(t, err)
		require.NotNil(t, aliceRelease)
		assert.True(t, aliceAcquired)
		defer aliceRelease()

		mockCurrentLockToken := fmt.Sprintf("%d,8527", time.Now().Add(-time.Minute).UnixNano())
		rs.SetNxFunc.PushReturn(false, nil)
		rs.GetFunc.PushReturn(redispool.NewValue(mockCurrentLockToken, nil))
		rs.GetSetFunc.PushReturn(redispool.NewValue(mockCurrentLockToken, nil))
		bobAcquired, bobRelease, err := tryAcquireRedisLockOnce(rs, "chicken-dinner", time.Minute)
		require.NoError(t, err)
		require.NotNil(t, bobRelease)
		assert.True(t, bobAcquired)
		mockrequire.CalledN(t, rs.SetNxFunc, 2)
		mockrequire.Called(t, rs.GetFunc)
		mockrequire.Called(t, rs.GetSetFunc)
		bobRelease()
	})

	t.Run("acquire an expired lock but act too slow", func(t *testing.T) {
		rs := redispool.NewMockKeyValue()
		rs.SetNxFunc.PushReturn(true, nil)
		aliceAcquired, aliceRelease, err := tryAcquireRedisLockOnce(rs, "chicken-dinner", time.Minute)
		require.NoError(t, err)
		require.NotNil(t, aliceRelease)
		assert.True(t, aliceAcquired)
		defer aliceRelease()

		mockCurrentLockToken := fmt.Sprintf("%d,8527", time.Now().Add(-time.Minute).UnixNano())
		rs.SetNxFunc.PushReturn(false, nil)
		rs.GetFunc.PushReturn(redispool.NewValue(mockCurrentLockToken, nil))
		rs.GetSetFunc.PushHook(func(_ string, value any) redispool.Value {
			return redispool.NewValue(value, nil) // Return anything that's not mockCurrentLockToken
		})
		bobAcquired, _, err := tryAcquireRedisLockOnce(rs, "chicken-dinner", time.Minute)
		require.NoError(t, err)
		assert.False(t, bobAcquired)
		mockrequire.CalledN(t, rs.SetNxFunc, 2)
		mockrequire.Called(t, rs.GetFunc)
		mockrequire.Called(t, rs.GetSetFunc)
	})
}
