package dotcomuser

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	graphqltypes "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/accesstoken"

	"github.com/gregjones/httpcache"
	"github.com/sourcegraph/log"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/notify"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// SourceVersion should be bumped whenever the format of any cached data in this
// actor source implementation is changed. This effectively expires all entries.
const SourceVersion = "v2"

// dotcom user gateway tokens are always a prefix of 4 characters ("sgd_")
// followed by a 64-character hex-encoded SHA256 hash
const tokenLength = 4 + 64

var (
	defaultUpdateInterval = 15 * time.Minute
	// defaultRefreshInterval is used for updates, which is also called when a
	// user's rate limit is hit, so we don't want to update every time. We use
	// a shorter interval than the default in this case.
	defaultRefreshInterval = 0 * time.Minute
)

type Source struct {
	log               log.Logger
	cache             httpcache.Cache // cache is expected to be something with automatic TTL
	dotcom            graphql.Client
	concurrencyConfig codygateway.ActorConcurrencyLimitConfig
	rs                limiter.RedisStore
	rateLimitNotifier notify.RateLimitNotifier
}

var _ actor.SourceUpdater = &Source{}

func NewSource(logger log.Logger, cache httpcache.Cache, dotComClient graphql.Client, concurrencyConfig codygateway.ActorConcurrencyLimitConfig, rs limiter.RedisStore, rateLimitNotifier notify.RateLimitNotifier) *Source {
	return &Source{
		log:               logger.Scoped("dotcomuser"),
		cache:             cache,
		dotcom:            dotComClient,
		concurrencyConfig: concurrencyConfig,
		rs:                rs,
		rateLimitNotifier: rateLimitNotifier,
	}
}

func (s *Source) Name() string { return string(codygateway.ActorSourceDotcomUser) }

func (s *Source) Get(ctx context.Context, token string) (*actor.Actor, error) {
	return s.get(ctx, token, false)
}

func (s *Source) Update(ctx context.Context, act *actor.Actor) error {
	if act.LastUpdated != nil && time.Since(*act.LastUpdated) < defaultRefreshInterval {
		return actor.ErrActorRecentlyUpdated{
			RetryAt: act.LastUpdated.Add(defaultRefreshInterval),
		}
	}

	_, err := s.get(ctx, act.Key, true)
	return err
}

// fetchAndCache fetches the dotcom user data for the given user token and caches it
func (s *Source) fetchAndCache(ctx context.Context, token string, oldAct *actor.Actor) (*actor.Actor, error) {
	var act *actor.Actor
	resp, checkErr := s.checkAccessToken(ctx, token)
	if checkErr != nil {
		// if oldAct is present in the cache, even though it is stale,
		// return it as a fallback if we fail to hit the dotcom API.
		if oldAct != nil && oldAct.ID != "" {
			return oldAct, nil
		}
		// Generate a stateless actor and save in the cache so that we aren't constantly hitting the dotcom API
		act = newActor(s, token, dotcom.DotcomUserState{}, s.concurrencyConfig)
	} else {
		act = newActor(s, token,
			resp.Dotcom.CodyGatewayDotcomUserByToken.DotcomUserState, s.concurrencyConfig)
	}

	if data, err := json.Marshal(act); err != nil {
		s.log.Error("failed to marshal actor",
			log.Error(err))
	} else {
		for _, feature := range codygateway.AllFeatures {
			rl, ok := act.RateLimits[feature]
			if !ok {
				continue
			}

			// reset the usage cache if:
			// 1. old actor is not present in the cache
			// 2. new rl interval is different from old rl interval
			if oldAct == nil || oldAct.ID == "" || rl.Interval != oldAct.RateLimits[feature].Interval {
				// get the base limiter for the feature which implements the core rate limiting based on the config
				l, ok := act.BaseLimiter(s.rs, feature, s.rateLimitNotifier)
				if !ok {
					return nil, errors.Wrap(err, fmt.Sprintf("failed to create base limiter for feature: %s and actor id: %s", feature, l.Identifier))
				}
				if err := l.ResetUsage(); err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf("failed to reset usage cache for feature: %s and actor id: %s", feature, l.Identifier))
				}
			}
		}

		s.cache.Set(token, data)
		// Also save to the key based on the actor ID.
		s.cache.Set(act.ID, data)
	}

	if checkErr != nil {
		return nil, errors.Wrap(checkErr, "failed to validate access token")
	}
	return act, nil
}

func (s *Source) checkAccessToken(ctx context.Context, token string) (*dotcom.CheckDotcomUserAccessTokenResponse, error) {
	resp, err := dotcom.CheckDotcomUserAccessToken(ctx, s.dotcom, token)
	if err == nil {
		return resp, nil
	}

	// Inspect the error to see if it's a list of GraphQL errors.
	gqlerrs, ok := err.(gqlerror.List)
	if !ok {
		return nil, err
	}

	for _, gqlerr := range gqlerrs {
		if gqlerr.Extensions != nil && gqlerr.Extensions["code"] == codygateway.GQLErrCodeDotcomUserNotFound {
			return nil, actor.ErrAccessTokenDenied{
				Source: s.Name(),
				Reason: "associated dotcom user not found",
			}
		}
	}
	return nil, err
}

func (s *Source) get(ctx context.Context, token string, bypassCache bool) (*actor.Actor, error) {
	if token == "" || !strings.HasPrefix(token, accesstoken.DotcomUserGatewayAccessTokenPrefix) {
		return nil, actor.ErrNotFromSource{}
	}

	if len(token) != tokenLength {
		return nil, errors.New("invalid token format")
	}
	// NOTE(naman): figure out a way to cache the actor in the limiter
	// based on actor id and not the token to avoid storing the config for each token.
	data, hit := s.cache.Get(token)
	if !hit {
		return s.fetchAndCache(ctx, token, nil)
	}

	var act *actor.Actor
	if err := json.Unmarshal(data, &act); err != nil || act == nil {
		trace.Logger(ctx, s.log).Error("failed to unmarshal actor", log.Error(err))

		// Delete the corrupted record.
		s.cache.Delete(token)

		return s.fetchAndCache(ctx, token, nil)
	}

	// if force fetch of rate-limit data from upstream or act cache older than defaultUpdateInterval (15 minutes)
	if bypassCache || act.LastUpdated == nil || time.Since(*act.LastUpdated) > defaultUpdateInterval {
		return s.fetchAndCache(ctx, token, act)
	}

	act.Source = s

	// Try to get data again from the actorID-based key
	dataWithLatestConfig, hit := s.cache.Get(act.ID)
	if hit {
		var actWithLatestConfig *actor.Actor
		if err := json.Unmarshal(dataWithLatestConfig, &actWithLatestConfig); err != nil || actWithLatestConfig == nil {
			trace.Logger(ctx, s.log).Error("failed to unmarshal actor", log.Error(err))

			// Delete the corrupted record.
			s.cache.Delete(act.ID)
		} else {
			for k, v := range actWithLatestConfig.RateLimits {
				act.RateLimits[k] = v
			}
		}
	}
	return act, nil
}

// newActor creates an actor from Sourcegraph.com user.
func newActor(source *Source, cacheKey string, user dotcom.DotcomUserState, concurrencyConfig codygateway.ActorConcurrencyLimitConfig) *actor.Actor {
	now := time.Now()

	userID := unmarshalUserID(user.Id)

	a := &actor.Actor{
		Key:           cacheKey,
		ID:            userID,
		Name:          user.Username,
		AccessEnabled: userID != "" && user.GetCodyGatewayAccess().Enabled,
		RateLimits:    zeroRequestsAllowed(),
		LastUpdated:   &now,
		Source:        source,
	}

	if rl := user.CodyGatewayAccess.ChatCompletionsRateLimit; rl != nil {
		a.RateLimits[codygateway.FeatureChatCompletions] = actor.NewRateLimitWithPercentageConcurrency(
			int64(rl.Limit),
			time.Duration(rl.IntervalSeconds)*time.Second,
			rl.AllowedModels,
			concurrencyConfig,
		)
	}

	if rl := user.CodyGatewayAccess.CodeCompletionsRateLimit; rl != nil {
		a.RateLimits[codygateway.FeatureCodeCompletions] = actor.NewRateLimitWithPercentageConcurrency(
			int64(rl.Limit),
			time.Duration(rl.IntervalSeconds)*time.Second,
			rl.AllowedModels,
			concurrencyConfig,
		)
	}

	if rl := user.CodyGatewayAccess.EmbeddingsRateLimit; rl != nil {
		a.RateLimits[codygateway.FeatureEmbeddings] = actor.NewRateLimitWithPercentageConcurrency(
			int64(rl.Limit),
			time.Duration(rl.IntervalSeconds)*time.Second,
			rl.AllowedModels,
			concurrencyConfig,
		)
	}

	return a
}

func zeroRequestsAllowed() map[codygateway.Feature]actor.RateLimit {
	return map[codygateway.Feature]actor.RateLimit{
		codygateway.FeatureChatCompletions: {},
		codygateway.FeatureCodeCompletions: {},
		codygateway.FeatureEmbeddings:      {},
	}
}

func unmarshalUserID(id string) (userID string) {
	if id == "" {
		return ""
	}
	var user int32
	err := relay.UnmarshalSpec(graphqltypes.ID(id), &user)
	if err != nil {
		return ""
	}
	return strconv.Itoa(int(user))
}
