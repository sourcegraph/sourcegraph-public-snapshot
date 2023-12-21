package dotcomuser

import (
	"testing"
	"time"

	"github.com/aws/smithy-go/ptr"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
)

func TestNewActor(t *testing.T) {
	concurrencyConfig := codygateway.ActorConcurrencyLimitConfig{
		Percentage: 50,
		Interval:   10 * time.Second,
	}
	type args struct {
		s dotcom.DotcomUserState
	}
	tests := []struct {
		name          string
		args          args
		wantEnabled   bool
		wantChatLimit int
		wantCodeLimit int
	}{
		{
			name: "enabled with rate limits",
			args: args{
				dotcom.DotcomUserState{
					Id: string(relay.MarshalID("User", 10)),
					CodyGatewayAccess: dotcom.DotcomUserStateCodyGatewayAccess{
						CodyGatewayAccessFields: dotcom.CodyGatewayAccessFields{
							Enabled: true,
							ChatCompletionsRateLimit: &dotcom.CodyGatewayAccessFieldsChatCompletionsRateLimitCodyGatewayRateLimit{
								RateLimitFields: dotcom.RateLimitFields{
									Limit:           10,
									IntervalSeconds: 10,
								},
							},
							CodeCompletionsRateLimit: &dotcom.CodyGatewayAccessFieldsCodeCompletionsRateLimitCodyGatewayRateLimit{
								RateLimitFields: dotcom.RateLimitFields{
									Limit:           20,
									IntervalSeconds: 20,
								},
							},
						},
					},
				},
			},
			wantEnabled:   true,
			wantChatLimit: 10,
			wantCodeLimit: 20,
		},
		{
			name: "disabled with rate limits",
			args: args{
				dotcom.DotcomUserState{
					Id: string(relay.MarshalID("User", 10)),
					CodyGatewayAccess: dotcom.DotcomUserStateCodyGatewayAccess{
						CodyGatewayAccessFields: dotcom.CodyGatewayAccessFields{
							Enabled: false,
							ChatCompletionsRateLimit: &dotcom.CodyGatewayAccessFieldsChatCompletionsRateLimitCodyGatewayRateLimit{
								RateLimitFields: dotcom.RateLimitFields{
									Limit:           10,
									IntervalSeconds: 10,
								},
							},
							CodeCompletionsRateLimit: &dotcom.CodyGatewayAccessFieldsCodeCompletionsRateLimitCodyGatewayRateLimit{
								RateLimitFields: dotcom.RateLimitFields{
									Limit:           20,
									IntervalSeconds: 20,
								},
							},
						},
					},
				},
			},
			wantEnabled:   false,
			wantChatLimit: 10,
			wantCodeLimit: 20,
		},
		{
			name: "enabled no limits",
			args: args{
				dotcom.DotcomUserState{
					Id: string(relay.MarshalID("User", 10)),
					CodyGatewayAccess: dotcom.DotcomUserStateCodyGatewayAccess{
						CodyGatewayAccessFields: dotcom.CodyGatewayAccessFields{
							Enabled: true,
						},
					},
				},
			},
			wantEnabled:   true,
			wantChatLimit: 0,
			wantCodeLimit: 0,
		},
		{
			name: "empty user",
			args: args{
				dotcom.DotcomUserState{},
			},
			wantEnabled:   false,
			wantChatLimit: 0,
			wantCodeLimit: 0,
		},
		{
			name: "invalid userID",
			args: args{
				dotcom.DotcomUserState{
					Id: "NOT_A_VALID_GQL_ID",
					CodyGatewayAccess: dotcom.DotcomUserStateCodyGatewayAccess{
						CodyGatewayAccessFields: dotcom.CodyGatewayAccessFields{
							Enabled: true,
						},
					},
				},
			},
			wantEnabled:   false,
			wantChatLimit: 0,
			wantCodeLimit: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			act := newActor(nil, "", tt.args.s, concurrencyConfig)
			assert.Equal(t, act.AccessEnabled, tt.wantEnabled)
		})
	}
}

func Test_UpdateWithGracePeriod(t *testing.T) {
	now := time.Now()
	updateTime := now.Add(1 * time.Second)
	withFetchErrorNow := func(act *actor.Actor) *actor.Actor {
		res := *act
		res.LastUpdateErrorAt = ptr.Time(updateTime)
		return &res
	}

	actorWithinUpdateInterval := &actor.Actor{LastUpdated: ptr.Time(now.Add(-1 * time.Minute))}
	actorOutOfGracePeriod := &actor.Actor{LastUpdated: ptr.Time(now.Add(-2 * time.Hour))}
	fetchedActor := &actor.Actor{LastUpdated: ptr.Time(now)}

	actorThatNeedsAnUpdate := &actor.Actor{LastUpdated: ptr.Time(now.Add(-16 * time.Minute))}
	actorThatNeedsAnUpdateAndHasFailedRecently := &actor.Actor{LastUpdated: ptr.Time(now.Add(-16 * time.Minute)), LastUpdateErrorAt: ptr.Time(now.Add(-2 * time.Minute))}
	actorThatNeedsAnUpdateAndHasFailedALongTimeAgo := &actor.Actor{LastUpdated: ptr.Time(now.Add(-16 * time.Minute)), LastUpdateErrorAt: ptr.Time(now.Add(-8 * time.Minute))}

	failedFetchActor := &actor.Actor{AccessEnabled: false}
	failedFetchError := errors.New("dotcom is down")

	tests := []struct {
		name            string
		oldActor        *actor.Actor
		fetchedActor    *actor.Actor
		fetchedError    error
		wantActor       *actor.Actor
		wantFetchCalled bool
		wantErr         error
	}{
		{"fetch_success_never_updated", &actor.Actor{}, fetchedActor, nil, fetchedActor, true, nil},
		{"fetch_success_updated_recently", actorWithinUpdateInterval, fetchedActor, nil, actorWithinUpdateInterval, false, nil},
		{"fetch_success_old", actorThatNeedsAnUpdate, fetchedActor, nil, fetchedActor, true, nil},
		{"fetch_fail_never_updated_actor", &actor.Actor{}, failedFetchActor, failedFetchError, failedFetchActor, true, failedFetchError},
		{"fetch_fail_updated_recently", actorWithinUpdateInterval, failedFetchActor, failedFetchError, actorWithinUpdateInterval, false, nil},
		{"fetch_fail_old_within_grace_period_first_error", actorThatNeedsAnUpdate, failedFetchActor, failedFetchError, withFetchErrorNow(actorThatNeedsAnUpdate), true, nil},
		{"fetch_fail_old_within_grace_period_retried_recently", actorThatNeedsAnUpdateAndHasFailedRecently, failedFetchActor, failedFetchError, actorThatNeedsAnUpdateAndHasFailedRecently, false, nil},
		{"fetch_fail_old_within_grace_period_needs_retry", actorThatNeedsAnUpdateAndHasFailedALongTimeAgo, failedFetchActor, failedFetchError, withFetchErrorNow(actorThatNeedsAnUpdateAndHasFailedALongTimeAgo), true, nil},
		{"fetch_fail_old_after_grace_period", actorOutOfGracePeriod, failedFetchActor, failedFetchError, failedFetchActor, true, failedFetchError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			res, err := updateWithGracePeriod(tt.oldActor, func() (*actor.Actor, error) {
				called = true
				return tt.fetchedActor, tt.fetchedError
			}, updateTime)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantActor, res)
			assert.Equal(t, tt.wantFetchCalled, called)
		})
	}
}
