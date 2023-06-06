package dotcomuser

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/gregjones/httpcache"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// dotcom user gateway tokens are always a prefix of 4 characters (sgd_)
// followed by a 64-character hex-encoded SHA256 hash
const tokenLength = 4 + 64

var (
	defaultUpdateInterval = 24 * time.Hour
)

type Source struct {
	log    log.Logger
	cache  httpcache.Cache
	dotcom graphql.Client
}

var _ actor.Source = &Source{}

func NewSource(logger log.Logger, cache httpcache.Cache, dotComClient graphql.Client) *Source {
	return &Source{
		log:    logger.Scoped("dotcomuser", "dotcom user actor source"),
		cache:  cache,
		dotcom: dotComClient,
	}
}

func (s *Source) Name() string { return string(codygateway.ActorSourceDotcomUser) }

func (s *Source) Get(ctx context.Context, token string) (*actor.Actor, error) {
	// "sgd_" is dotcomUserGatewayAccessTokenPrefix
	if token == "" || !strings.HasPrefix(token, "sgd_") {
		return nil, actor.ErrNotFromSource{}
	}

	if len(token) != tokenLength {
		return nil, errors.New("invalid token format")
	}

	data, hit := s.cache.Get(token)
	if !hit {
		return s.fetchAndCache(ctx, token)
	}

	var act *actor.Actor
	if err := json.Unmarshal(data, &act); err != nil || act == nil {
		s.log.Error("failed to unmarshal subscription", log.Error(err))

		// Delete the corrupted record.
		s.cache.Delete(token)

		return s.fetchAndCache(ctx, token)
	}

	if act.LastUpdated != nil && time.Since(*act.LastUpdated) > defaultUpdateInterval {
		return s.fetchAndCache(ctx, token)
	}

	act.Source = s
	return act, nil
}

// fetchAndCache fetches the dotcom user data for the given user token and caches it
func (s *Source) fetchAndCache(ctx context.Context, token string) (*actor.Actor, error) {
	var act *actor.Actor
	resp, checkErr := dotcom.CheckDotcomUserAccessToken(ctx, s.dotcom, token)
	if checkErr != nil {
		// Generate a stateless actor so that we aren't constantly hitting the dotcom API
		act = NewActor(s, token, dotcom.DotcomUserState{})
	} else {
		act = NewActor(s, token,
			resp.Dotcom.DotcomCodyGatewayUserByToken.DotcomUserState)
	}

	if data, err := json.Marshal(act); err != nil {
		s.log.Error("failed to marshal actor",
			log.Error(err))
	} else {
		s.cache.Set(token, data)
	}

	if checkErr != nil {
		return nil, errors.Wrap(checkErr, "failed to validate access token")
	}
	return act, nil
}

// NewActor creates an actor from Sourcegraph.com user.
func NewActor(source *Source, cacheKey string, user dotcom.DotcomUserState) *actor.Actor {
	now := time.Now()
	a := &actor.Actor{
		Key:           cacheKey,
		ID:            user.UserName,
		AccessEnabled: user.GetCodyGatewayAccess().Enabled,
		RateLimits:    zeroRequestsAllowed(),
		LastUpdated:   &now,
		Source:        source,
	}

	if user.CodyGatewayAccess.ChatCompletionsRateLimit != nil {
		a.RateLimits[types.CompletionsFeatureChat] = actor.RateLimit{
			AllowedModels: user.CodyGatewayAccess.ChatCompletionsRateLimit.AllowedModels,
			Limit:         user.CodyGatewayAccess.ChatCompletionsRateLimit.Limit,
			Interval:      time.Duration(user.CodyGatewayAccess.ChatCompletionsRateLimit.IntervalSeconds) * time.Second,
		}
	}

	if user.CodyGatewayAccess.CodeCompletionsRateLimit != nil {
		a.RateLimits[types.CompletionsFeatureCode] = actor.RateLimit{
			AllowedModels: user.CodyGatewayAccess.CodeCompletionsRateLimit.AllowedModels,
			Limit:         user.CodyGatewayAccess.CodeCompletionsRateLimit.Limit,
			Interval:      time.Duration(user.CodyGatewayAccess.CodeCompletionsRateLimit.IntervalSeconds) * time.Second,
		}
	}

	return a
}

func zeroRequestsAllowed() map[types.CompletionsFeature]actor.RateLimit {
	return map[types.CompletionsFeature]actor.RateLimit{
		types.CompletionsFeatureChat: {},
		types.CompletionsFeatureCode: {},
	}
}
