package dotcomuser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/dotcom"
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
					Username: "user",
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
					Username: "user",
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
					Username: "user",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			act := newActor(nil, "", tt.args.s, concurrencyConfig)
			assert.Equal(t, act.AccessEnabled, tt.wantEnabled)
		})
	}
}
