package productsubscription

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/gregjones/httpcache"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	minUpdateInterval = 10 * time.Minute

	defaultUpdateInterval = 24 * time.Hour
)

type Source struct {
	log    log.Logger
	cache  httpcache.Cache // TODO: add something to regularly clean up the cache
	dotcom graphql.Client
}

var _ actor.Source = &Source{}
var _ actor.SourceUpdater = &Source{}

func NewSource(logger log.Logger, cache httpcache.Cache, dotComClient graphql.Client) *Source {
	return &Source{
		log:    logger.Scoped("productsubscriptions", "product subscription actor source"),
		cache:  cache,
		dotcom: dotComClient,
	}
}

func (s *Source) Name() string { return "sourcegraph.com product subscriptions" }

func (s *Source) Get(ctx context.Context, token string) (*actor.Actor, error) {
	// "sgs_" is productSubscriptionAccessTokenPrefix
	if token == "" && !strings.HasPrefix(token, "sgs_") {
		return nil, actor.ErrNotFromSource{}
	}

	data, hit := s.cache.Get(token)
	if !hit {
		return s.fetchAndCache(ctx, token)
	}

	var act *actor.Actor
	if err := json.Unmarshal(data, &act); err != nil {
		s.log.Error("failed to unmarshal subscription", log.Error(err))
		return s.fetchAndCache(ctx, token)
	}

	if act.LastUpdated != nil && time.Since(*act.LastUpdated) > defaultUpdateInterval {
		return s.fetchAndCache(ctx, token)
	}

	return act, nil
}

func (s *Source) Update(ctx context.Context, actor *actor.Actor) {
	if time.Since(*actor.LastUpdated) < minUpdateInterval {
		// Last update was too recent - do it later.
		return
	}

	if _, err := s.fetchAndCache(ctx, actor.Key); err != nil {
		s.log.Info("failed to update actor", log.Error(err))
	}
}

func (s *Source) fetchAndCache(ctx context.Context, token string) (*actor.Actor, error) {
	var act *actor.Actor
	resp, checkErr := dotcom.CheckAccessToken(ctx, s.dotcom, token)
	if checkErr != nil {
		// Generate a stateless actor so that we aren't constantly hitting the dotcom API
		act = NewActor(s, token, dotcom.ProductSubscriptionState{})
	} else {
		act = NewActor(s, token, resp.Dotcom.ProductSubscriptionByAccessToken.ProductSubscriptionState)
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

// NewActor creates an actor from Sourcegraph.com product subscription state.
func NewActor(source *Source, token string, s dotcom.ProductSubscriptionState) *actor.Actor {
	var rateLimit actor.RateLimit
	if s.LlmProxyAccess.RateLimit != nil {
		rateLimit = actor.RateLimit{
			Limit:    s.LlmProxyAccess.RateLimit.Limit,
			Interval: time.Duration(s.LlmProxyAccess.RateLimit.IntervalSeconds) * time.Second,
		}
	}

	now := time.Now()
	return &actor.Actor{
		Key:           token,
		ID:            s.Id,
		AccessEnabled: !s.IsArchived && s.LlmProxyAccess.Enabled && rateLimit.IsValid(),
		RateLimit:     rateLimit,
		LastUpdated:   &now,
		Source:        source,
	}
}
