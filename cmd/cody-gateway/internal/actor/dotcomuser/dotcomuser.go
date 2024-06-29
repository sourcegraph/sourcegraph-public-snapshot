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
	"github.com/sourcegraph/sourcegraph/internal/codygateway/codygatewayactor"

	"github.com/gregjones/httpcache"
	"github.com/sourcegraph/log"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/limiter"
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
)

type Source struct {
	log               log.Logger
	cache             httpcache.Cache // cache is expected to be something with automatic TTL
	dotcom            graphql.Client
	concurrencyConfig codygatewayactor.ActorConcurrencyLimitConfig
	usageStore        limiter.RedisStore
	coolDownInterval  time.Duration
}

var _ actor.SourceUpdater = &Source{}

func NewSource(logger log.Logger, cache httpcache.Cache, dotComClient graphql.Client, concurrencyConfig codygatewayactor.ActorConcurrencyLimitConfig, usageStore limiter.RedisStore, coolDownInterval time.Duration) *Source {
	return &Source{
		log:               logger.Scoped("dotcomuser"),
		cache:             cache,
		dotcom:            dotComClient,
		concurrencyConfig: concurrencyConfig,
		usageStore:        usageStore,
		coolDownInterval:  coolDownInterval,
	}
}

func (s *Source) Name() string { return string(codygatewayactor.ActorSourceDotcomUser) }

func (s *Source) Get(ctx context.Context, token string) (*actor.Actor, error) {
	return s.get(ctx, token, false)
}

func (s *Source) Update(ctx context.Context, act *actor.Actor) error {
	if act.LastUpdated != nil && time.Since(*act.LastUpdated) < s.coolDownInterval {
		s.log.Debug("actor recently updated, skipping update", log.Duration("secondsSinceUpdate", time.Since(*act.LastUpdated)))
		return actor.ErrActorRecentlyUpdated{
			RetryAt: act.LastUpdated.Add(s.coolDownInterval),
		}
	}

	_, err := s.get(ctx, act.Key, true)
	return err
}

// fetchAndCache fetches the dotcom user data for the given user token and caches it.
// We compare against the previous state of the actor (if available) to trigger logic based
// on any updates.
func (s *Source) fetchAndCache(ctx context.Context, token string, oldAct *actor.Actor) (*actor.Actor, error) {
	var act *actor.Actor
	resp, checkErr := s.checkAccessToken(ctx, token)
	if checkErr != nil {
		// Generate a stateless actor so that we aren't constantly hitting the dotcom API.
		act = newActor(s, token, dotcom.DotcomUserState{}, s.concurrencyConfig)
	} else {
		act = newActor(s, token,
			resp.Dotcom.CodyGatewayDotcomUserByToken.DotcomUserState, s.concurrencyConfig)
	}

	// Marshall the actor into JSON so we can persist it.
	if data, err := json.Marshal(act); err != nil {
		s.log.Error("failed to marshal actor", log.Error(err))
	} else {
		// Now that we know the Actor ID, try to load the cached actor data if needed.
		if oldAct == nil {
			var cachedActor actor.Actor
			if previousActorData, hit := s.cache.Get(act.ID); hit {
				if err := json.Unmarshal(previousActorData, &cachedActor); err != nil {
					trace.Logger(ctx, s.log).Error("failed to unmarshal old actor", log.Error(err))
				} else {
					oldAct = &cachedActor
				}
			}
		}

		// As part of fetching the actor data, we may also want to reset their usage data.
		// e.g. the usage limits have changed from the data we just loaded vs. oldAct.
		if err = s.maybeResetUsageData(*act, oldAct); err != nil {
			return nil, errors.Wrap(err, "resetting usage data")
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

// maybeResetUsageData will reset the actor's usage data for all Cody features if the previous
// actor state is unknown (e.g. wasn't present in our cache) or if the rate limit interval has
// changed (e.g. changed their Cody Pro subscription plan).
func (s *Source) maybeResetUsageData(current actor.Actor, oldAct *actor.Actor) error {
	for _, feature := range codygateway.AllFeatures {
		rl, ok := current.RateLimits[feature]
		if !ok {
			continue
		}

		// The actual RateLimit will be a zero value in the event of an error reading the
		// Actor from dotcom. In these cases, rl.IsValid() is false and so we just ignore
		// resetting the usage data to avoid cascading failure.
		if !rl.IsValid() {
			continue
		}

		// If the expiry on the key is greater than the intervalSeconds, update the TTL.
		// This is here as a safeguard for the case where the TTL is wrongly set or the
		// usage cache is not reset when the intervalSeconds config change. This only covers
		// the case where the new intervalSeconds in shorter than the present TTL.
		isTTLGreaterThanInterval := false

		featureUsageStore := limiter.NewFeatureUsageStore(s.usageStore, feature)

		if ttl, err := featureUsageStore.TTL(current.ID); err == nil {
			if ttl > int(rl.Interval.Seconds()) {
				isTTLGreaterThanInterval = true
			}
		}

		// If oldActor is nil/empty, then we reset because we don't have any usage data available.
		// If the rate limit interval has changed, reset their previous usage as a side-effect.
		// (e.g. zero out previous usage as part of upgrading the plan. This has the potential for
		// abuse, which we prevent by limiting how frequently a user can change their subscription plan.)
		if oldAct.IsEmpty() || rl.Interval != oldAct.RateLimits[feature].Interval || isTTLGreaterThanInterval {
			if err := featureUsageStore.Del(current.ID); err != nil {
				message := fmt.Sprintf("failed to reset usage cache for feature: %s and actor id: %s", feature, current.ID)
				return errors.Wrap(err, message)
			}
		}
	}

	return nil
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

		// Delete the corrupted record in our cache, and try a fresh fetch.
		s.cache.Delete(token)

		return s.fetchAndCache(ctx, token, nil)
	}

	// If our cached data is sufficiently old, refresh it proactively.
	if bypassCache || act.LastUpdated == nil || time.Since(*act.LastUpdated) > defaultUpdateInterval {
		return s.fetchAndCache(ctx, token, act)
	}

	act.Source = s

	// Try to get data again from the actorID-based key. This is because rate limit data is written to
	// the cache by the actor ID, instead of the access-token based lookup key we used earlier.
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
func newActor(source *Source, cacheKey string, user dotcom.DotcomUserState, concurrencyConfig codygatewayactor.ActorConcurrencyLimitConfig) *actor.Actor {
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
