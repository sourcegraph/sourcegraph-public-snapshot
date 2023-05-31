package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Khan/genqlient/graphql"
	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/actor/anonymous"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/actor/productsubscription"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
)

// TODO(@bobheadxi): Try to rewrite this as a table-driven test for less copy-pasta.
func TestAuthenticatorMiddleware(t *testing.T) {
	logger := logtest.Scoped(t)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	t.Run("unauthenticated and allow anonymous", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		(&Authenticator{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     actor.Sources{anonymous.NewSource(true)},
		}).Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthenticated but disallow anonymous", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		(&Authenticator{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     actor.Sources{anonymous.NewSource(false)},
		}).Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("authenticated without cache hit", func(t *testing.T) {
		cache := NewMockCache()
		client := NewMockClient()
		client.MakeRequestFunc.SetDefaultHook(func(_ context.Context, _ *graphql.Request, resp *graphql.Response) error {
			resp.Data.(*dotcom.CheckAccessTokenResponse).Dotcom = dotcom.CheckAccessTokenDotcomDotcomQuery{
				ProductSubscriptionByAccessToken: dotcom.CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription{
					ProductSubscriptionState: dotcom.ProductSubscriptionState{
						Id:         "UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==",
						Uuid:       "6452a8fc-e650-45a7-a0a2-357f776b3b46",
						IsArchived: false,
						CodyGatewayAccess: dotcom.ProductSubscriptionStateCodyGatewayAccess{
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
										Limit:           10,
										IntervalSeconds: 10,
									},
								},
							},
						},
					},
				},
			}
			return nil
		})
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NotNil(t, actor.FromContext(r.Context()))
			w.WriteHeader(http.StatusOK)
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer sgs_abc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250a4a861f37")
		(&Authenticator{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     actor.Sources{productsubscription.NewSource(logger, cache, client, false)},
		}).Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		mockrequire.Called(t, client.MakeRequestFunc)
	})

	t.Run("authenticated with cache hit", func(t *testing.T) {
		cache := NewMockCache()
		cache.GetFunc.SetDefaultReturn(
			[]byte(`{"id":"UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==","accessEnabled":true,"rateLimit":null}`),
			true,
		)
		client := NewMockClient()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NotNil(t, actor.FromContext(r.Context()))
			w.WriteHeader(http.StatusOK)
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer sgs_abc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250a4a861f37")
		(&Authenticator{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     actor.Sources{productsubscription.NewSource(logger, cache, client, false)},
		}).Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		mockrequire.NotCalled(t, client.MakeRequestFunc)
	})

	t.Run("authenticated but not enabled", func(t *testing.T) {
		cache := NewMockCache()
		cache.GetFunc.SetDefaultReturn(
			[]byte(`{"id":"UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==","accessEnabled":false,"rateLimit":null}`),
			true,
		)
		client := NewMockClient()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer sgs_abc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250a4a861f37")
		(&Authenticator{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     actor.Sources{productsubscription.NewSource(logger, cache, client, false)},
		}).Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("internal mode, authenticated but not dev license", func(t *testing.T) {
		cache := NewMockCache()
		client := NewMockClient()
		client.MakeRequestFunc.SetDefaultHook(func(_ context.Context, _ *graphql.Request, resp *graphql.Response) error {
			resp.Data.(*dotcom.CheckAccessTokenResponse).Dotcom = dotcom.CheckAccessTokenDotcomDotcomQuery{
				ProductSubscriptionByAccessToken: dotcom.CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription{
					ProductSubscriptionState: dotcom.ProductSubscriptionState{
						Id:         "UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==",
						Uuid:       "6452a8fc-e650-45a7-a0a2-357f776b3b46",
						IsArchived: false,
						CodyGatewayAccess: dotcom.ProductSubscriptionStateCodyGatewayAccess{
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
										Limit:           10,
										IntervalSeconds: 10,
									},
								},
							},
						},
						ActiveLicense: &dotcom.ProductSubscriptionStateActiveLicenseProductLicense{
							Info: &dotcom.ProductSubscriptionStateActiveLicenseProductLicenseInfo{
								Tags: []string{""},
							},
						},
					},
				},
			}
			return nil
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer sgs_abc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250a4a861f37")
		(&Authenticator{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     actor.Sources{productsubscription.NewSource(logger, cache, client, true)},
		}).Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("internal mode, authenticated dev license", func(t *testing.T) {
		cache := NewMockCache()
		client := NewMockClient()
		client.MakeRequestFunc.SetDefaultHook(func(_ context.Context, _ *graphql.Request, resp *graphql.Response) error {
			resp.Data.(*dotcom.CheckAccessTokenResponse).Dotcom = dotcom.CheckAccessTokenDotcomDotcomQuery{
				ProductSubscriptionByAccessToken: dotcom.CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription{
					ProductSubscriptionState: dotcom.ProductSubscriptionState{
						Id:         "UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==",
						Uuid:       "6452a8fc-e650-45a7-a0a2-357f776b3b46",
						IsArchived: false,
						CodyGatewayAccess: dotcom.ProductSubscriptionStateCodyGatewayAccess{
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
										Limit:           10,
										IntervalSeconds: 10,
									},
								},
							},
						},
						ActiveLicense: &dotcom.ProductSubscriptionStateActiveLicenseProductLicense{
							Info: &dotcom.ProductSubscriptionStateActiveLicenseProductLicenseInfo{
								Tags: []string{licensing.DevTag},
							},
						},
					},
				},
			}
			return nil
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer sgs_abc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250a4a861f37")
		(&Authenticator{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     actor.Sources{productsubscription.NewSource(logger, cache, client, true)},
		}).Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("internal mode, authenticated internal license", func(t *testing.T) {
		cache := NewMockCache()
		client := NewMockClient()
		client.MakeRequestFunc.SetDefaultHook(func(_ context.Context, _ *graphql.Request, resp *graphql.Response) error {
			resp.Data.(*dotcom.CheckAccessTokenResponse).Dotcom = dotcom.CheckAccessTokenDotcomDotcomQuery{
				ProductSubscriptionByAccessToken: dotcom.CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription{
					ProductSubscriptionState: dotcom.ProductSubscriptionState{
						Id:         "UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==",
						Uuid:       "6452a8fc-e650-45a7-a0a2-357f776b3b46",
						IsArchived: false,
						CodyGatewayAccess: dotcom.ProductSubscriptionStateCodyGatewayAccess{
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
										Limit:           10,
										IntervalSeconds: 10,
									},
								},
							},
						},
						ActiveLicense: &dotcom.ProductSubscriptionStateActiveLicenseProductLicense{
							Info: &dotcom.ProductSubscriptionStateActiveLicenseProductLicenseInfo{
								Tags: []string{licensing.DevTag},
							},
						},
					},
				},
			}
			return nil
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer sgs_abc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250a4a861f37")
		(&Authenticator{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     actor.Sources{productsubscription.NewSource(logger, cache, client, true)},
		}).Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
