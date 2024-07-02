package dotcomuser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go/relay"
	"github.com/gregjones/httpcache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	graphql "github.com/Khan/genqlient/graphql"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/internal/accesstoken"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/codygateway/codygatewayactor"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestNewActor(t *testing.T) {
	concurrencyConfig := codygatewayactor.ActorConcurrencyLimitConfig{
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

func TestActorCacheExpiration(t *testing.T) {
	ctx := context.Background()

	var (
		oneHundredAnHour = dotcom.RateLimitFields{
			Limit:           100,
			IntervalSeconds: 60 * 60,
			AllowedModels:   []string{"sourcegraph/codebot-9000"},
		}
		twoHundredAMinute = dotcom.RateLimitFields{
			Limit:           200,
			IntervalSeconds: 60,
			AllowedModels:   []string{"sourcegraph/codebot-9000"},
		}

		testAccountID         = 123456
		testUserName          = "Chris Smith"
		testCodyGatewayAccess = dotcom.DotcomUserStateCodyGatewayAccess{
			CodyGatewayAccessFields: dotcom.CodyGatewayAccessFields{
				Enabled: true,
				ChatCompletionsRateLimit: &dotcom.CodyGatewayAccessFieldsChatCompletionsRateLimitCodyGatewayRateLimit{
					RateLimitFields: oneHundredAnHour,
				},
				CodeCompletionsRateLimit: &dotcom.CodyGatewayAccessFieldsCodeCompletionsRateLimitCodyGatewayRateLimit{
					RateLimitFields: oneHundredAnHour,
				},
				EmbeddingsRateLimit: &dotcom.CodyGatewayAccessFieldsEmbeddingsRateLimitCodyGatewayRateLimit{
					RateLimitFields: oneHundredAnHour,
				},
			},
		}
	)

	assertRateLimitsEqual := func(t *testing.T, dotcomRL dotcom.RateLimitFields, actorRL actor.RateLimit) {
		t.Helper()
		assert.EqualValues(t, dotcomRL.AllowedModels, actorRL.AllowedModels)
		assert.Equal(t, time.Duration(dotcomRL.IntervalSeconds)*time.Second, actorRL.Interval)
		assert.EqualValues(t, dotcomRL.Limit, actorRL.Limit)
	}

	// Setup fakes and mocks.
	dotcomMock := dotcom.NewMockClient()
	fakeRedisStore := limiter.NewRecordingRedisStoreFake()
	fakeSource := &Source{
		dotcom:     dotcomMock,
		cache:      httpcache.NewMemoryCache(),
		usageStore: fakeRedisStore,
	}

	dotcomMock.MakeRequestFunc.SetDefaultReturn(errors.New("no response configured for mock"))
	// The first dotcom GraphQL call is to fetch the user's state.
	dotcomMock.MakeRequestFunc.PushHook(func(ctx context.Context, inReq *graphql.Request, outResp *graphql.Response) error {
		if inReq.OpName != "CheckDotcomUserAccessToken" {
			t.Fatal("Got unexpected call to dotcomMock.MakeRequestFunc")
		}

		outData, ok := outResp.Data.(*dotcom.CheckDotcomUserAccessTokenResponse)
		if !ok {
			t.Fatal("graphql response data not of expected type")
		}
		*outData = dotcom.CheckDotcomUserAccessTokenResponse{
			Dotcom: dotcom.CheckDotcomUserAccessTokenDotcomDotcomQuery{
				CodyGatewayDotcomUserByToken: &dotcom.CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGatewayDotcomUserByTokenCodyGatewayDotcomUser{
					DotcomUserState: dotcom.DotcomUserState{
						Id:                string(relay.MarshalID("User", testAccountID)),
						Username:          testUserName,
						CodyGatewayAccess: testCodyGatewayAccess,
					},
				},
			},
		}

		return nil
	})

	testAccessToken := accesstoken.DotcomUserGatewayAccessTokenPrefix + strings.Repeat("x", 64)

	// Pop the first element in the RRS' history and confirm it matches the expected string.
	assertFirstOperation := func(t *testing.T, rrs *limiter.RecordingRedisStoreFake, want string) {
		t.Helper()
		if len(rrs.History) == 0 {
			t.Error("No history available for RecordingRedisStoreFake")
		} else {
			got := rrs.History[0]
			rrs.History = rrs.History[1:]
			assert.Equal(t, want, got)
		}
	}

	t.Run("InitialGetCall", func(t *testing.T) {
		gotActor, err := fakeSource.get(ctx, testAccessToken, false /* bypassCache */)
		require.NoError(t, err)
		require.NotNil(t, gotActor)

		assert.Equal(t, fmt.Sprintf("%d", testAccountID), gotActor.ID)
		assert.Equal(t, testUserName, gotActor.Name)

		require.NotNil(t, gotActor.RateLimits)
		assertRateLimitsEqual(t, oneHundredAnHour, gotActor.RateLimits[codygateway.FeatureChatCompletions])
		assertRateLimitsEqual(t, oneHundredAnHour, gotActor.RateLimits[codygateway.FeatureCodeCompletions])

		// A side-effect of calling get will reset the user's usage data. We confirm this by
		// looking at the operations performed on the Redis cache.
		assertFirstOperation(t, fakeRedisStore, "TTL(code_completions:123456)")
		assertFirstOperation(t, fakeRedisStore, "Del(code_completions:123456)")
		assertFirstOperation(t, fakeRedisStore, "TTL(chat_completions:123456)")
		assertFirstOperation(t, fakeRedisStore, "Del(chat_completions:123456)")
		assertFirstOperation(t, fakeRedisStore, "TTL(embeddings:123456)")
		assertFirstOperation(t, fakeRedisStore, "Del(embeddings:123456)")
		// Confirm there are no other RedisStore operations performed.
		assert.Equal(t, 0, len(fakeRedisStore.History))
	})

	// Call get() a second time. This should just return cached data and have no side-effects.
	t.Run("SecondGetCall", func(t *testing.T) {
		gotActor, err := fakeSource.get(ctx, testAccessToken, false /* bypassCache */)
		require.NoError(t, err)

		require.NotNil(t, gotActor)
		assert.Equal(t, fmt.Sprintf("%d", testAccountID), gotActor.ID)
		require.NotNil(t, gotActor.RateLimits)
		assertRateLimitsEqual(t, oneHundredAnHour, gotActor.RateLimits[codygateway.FeatureChatCompletions])

		assert.Equal(t, 0, len(fakeRedisStore.History))
	})

	// Call fakeSource.get but with a new user access token for the same user.
	// Here we rely on the existing data that is in our Redis cache.
	t.Run("NewAccessToken", func(t *testing.T) {
		altAccessToken := accesstoken.DotcomUserGatewayAccessTokenPrefix + strings.Repeat("y", 64)

		// Mock out the required call to verify the new access token is valid.
		dotcomMock.MakeRequestFunc.PushHook(func(ctx context.Context, inReq *graphql.Request, outResp *graphql.Response) error {
			if inReq.OpName != "CheckDotcomUserAccessToken" {
				t.Fatal("Got unexpected call to dotcomMock.MakeRequestFunc")
			}
			outData, ok := outResp.Data.(*dotcom.CheckDotcomUserAccessTokenResponse)
			if !ok {
				t.Fatal("graphql response data not of expected type")
			}
			*outData = dotcom.CheckDotcomUserAccessTokenResponse{
				Dotcom: dotcom.CheckDotcomUserAccessTokenDotcomDotcomQuery{
					CodyGatewayDotcomUserByToken: &dotcom.CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGatewayDotcomUserByTokenCodyGatewayDotcomUser{
						DotcomUserState: dotcom.DotcomUserState{
							Id:                string(relay.MarshalID("User", testAccountID)),
							Username:          testUserName,
							CodyGatewayAccess: testCodyGatewayAccess,
						},
					},
				},
			}
			return nil
		})

		// By using a new access token, we will end up needing to add more data to the Redis cache.
		gotActor, err := fakeSource.get(ctx, altAccessToken, false /* bypassCache */)
		require.NoError(t, err)

		require.NotNil(t, gotActor)
		assert.Equal(t, fmt.Sprintf("%d", testAccountID), gotActor.ID)
		require.NotNil(t, gotActor.RateLimits)
		assertRateLimitsEqual(t, oneHundredAnHour, gotActor.RateLimits[codygateway.FeatureChatCompletions])

		// We check the TTLs on existing keys in the cache, but do not actually modify anything.
		assertFirstOperation(t, fakeRedisStore, "TTL(code_completions:123456)")
		assertFirstOperation(t, fakeRedisStore, "TTL(chat_completions:123456)")
		assertFirstOperation(t, fakeRedisStore, "TTL(embeddings:123456)")
		assert.Equal(t, 0, len(fakeRedisStore.History))
	})

	t.Run("LimitDurationChanged", func(t *testing.T) {
		// Update the HTTP cache and rewrite the Actor data setting the
		// LastUpdate time to a value old enough we expect it to need refreshing.
		t.Run("ExpireCachedActorData", func(t *testing.T) {
			rawActorJSON, ok := fakeSource.cache.Get(testAccessToken)
			require.True(t, ok)

			var cachedActor actor.Actor
			if err := json.Unmarshal(rawActorJSON, &cachedActor); err != nil {
				t.Fatalf("parsing cached Actor JSON: %v", err)
			}

			// Confirm LastUpdated was set and is recent.
			assert.NotNil(t, cachedActor.LastUpdated)
			assert.True(t, time.Since(*cachedActor.LastUpdated) < time.Minute)

			// Confirm RateLimits were part of the Actor data.
			assert.NotNil(t, cachedActor.RateLimits)

			// Rewrite the Actor stored in the cache and act as if it is stale.
			staleTime := time.Now().Add(-defaultUpdateInterval - time.Minute)
			cachedActor.LastUpdated = &staleTime

			updatedActorJSON, err := json.Marshal(cachedActor)
			require.NoError(t, err)
			fakeSource.cache.Set(testAccessToken, updatedActorJSON)
		})

		// Call get(). This will force a re-fetching the user's data.
		// This will trigger a NEW call to the dotcom GraphQL API, so we return updated user data.
		t.Run("RegisterNewMockActorGraphQLResponse", func(t *testing.T) {
			dotcomMock.MakeRequestFunc.PushHook(func(ctx context.Context, inReq *graphql.Request, outResp *graphql.Response) error {
				if inReq.OpName != "CheckDotcomUserAccessToken" {
					t.Fatal("Got unexpected call to dotcomMock.MakeRequestFunc")
				}

				// Update CodeCompletions to a new value.
				testCodyGatewayAccess.CodeCompletionsRateLimit.RateLimitFields = twoHundredAMinute

				outData, ok := outResp.Data.(*dotcom.CheckDotcomUserAccessTokenResponse)
				if !ok {
					t.Fatal("graphql response data not of expected type")
				}
				*outData = dotcom.CheckDotcomUserAccessTokenResponse{
					Dotcom: dotcom.CheckDotcomUserAccessTokenDotcomDotcomQuery{
						CodyGatewayDotcomUserByToken: &dotcom.CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGatewayDotcomUserByTokenCodyGatewayDotcomUser{
							DotcomUserState: dotcom.DotcomUserState{
								Id:                string(relay.MarshalID("User", testAccountID)),
								Username:          testUserName,
								CodyGatewayAccess: testCodyGatewayAccess,
							},
						},
					},
				}

				return nil
			})
		})

		gotActor, err := fakeSource.get(ctx, testAccessToken, false /* bypassCache */)
		require.NoError(t, err)

		require.NotNil(t, gotActor)
		assert.Equal(t, fmt.Sprintf("%d", testAccountID), gotActor.ID)
		require.NotNil(t, gotActor.RateLimits)
		assertRateLimitsEqual(t, oneHundredAnHour, gotActor.RateLimits[codygateway.FeatureChatCompletions])

		// Confirm the CodeCompletions value was updated.
		assertRateLimitsEqual(t, twoHundredAMinute, gotActor.RateLimits[codygateway.FeatureCodeCompletions])

		// Confirm that the CodeCompletions usage data was deleted as well.
		assertFirstOperation(t, fakeRedisStore, "TTL(code_completions:123456)")
		assertFirstOperation(t, fakeRedisStore, "Del(code_completions:123456)")
		assert.Equal(t, 2, len(fakeRedisStore.History))
	})
}
