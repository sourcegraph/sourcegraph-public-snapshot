package productsubscription

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/gregjones/httpcache"
	"github.com/sourcegraph/log"
	"golang.org/x/exp/slices"

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

	devLicensesOnly bool
}

var _ actor.Source = &Source{}
var _ actor.SourceUpdater = &Source{}
var _ actor.SourceSyncer = &Source{}

func NewSource(logger log.Logger, cache httpcache.Cache, dotComClient graphql.Client, devLicensesOnly bool) *Source {
	return &Source{
		log:    logger.Scoped("productsubscriptions", "product subscription actor source"),
		cache:  cache,
		dotcom: dotComClient,

		devLicensesOnly: devLicensesOnly,
	}
}

func (s *Source) Name() string { return "dotcom-product-subscriptions" }

func (s *Source) Get(ctx context.Context, token string) (*actor.Actor, error) {
	// "sgs_" is productSubscriptionAccessTokenPrefix
	if token == "" || !strings.HasPrefix(token, "sgs_") {
		return nil, actor.ErrNotFromSource{}
	}

	data, hit := s.cache.Get(token)
	if !hit {
		return s.fetchAndCache(ctx, token)
	}

	var act *actor.Actor
	if err := json.Unmarshal(data, &act); err != nil {
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

func (s *Source) Update(ctx context.Context, actor *actor.Actor) {
	if time.Since(*actor.LastUpdated) < minUpdateInterval {
		// Last update was too recent - do it later.
		return
	}

	if _, err := s.fetchAndCache(ctx, actor.Key); err != nil {
		s.log.Info("failed to update actor", log.Error(err))
	}
}

// Sync retrieves all known actors from this source and updates its cache.
// All Sync implementations are called periodically - implementations can decide
// to skip syncs if the frequency is too high.
func (s *Source) Sync(ctx context.Context) (seen int, errs error) {
	resp, err := dotcom.ListProductSubscriptions(ctx, s.dotcom)
	if err != nil {
		return seen, errors.Wrap(err, "failed to list subscriptions from dotcom")
	}
	for _, sub := range resp.Dotcom.ProductSubscriptions.Nodes {
		for _, token := range sub.SourcegraphAccessTokens {
			act := NewActor(s, token, sub.ProductSubscriptionState, s.devLicensesOnly)
			data, err := json.Marshal(act)
			if err != nil {
				s.log.Error("failed to marshal actor", log.Error(err))
				errs = errors.Append(errs, err)
				continue
			}
			s.cache.Set(token, data)
			seen++
		}
	}
	// TODO: Here we should prune all cache keys that we haven't seen in the sync
	// loop.
	return seen, errs
}

func (s *Source) fetchAndCache(ctx context.Context, token string) (*actor.Actor, error) {
	var act *actor.Actor
	resp, checkErr := dotcom.CheckAccessToken(ctx, s.dotcom, token)
	if checkErr != nil {
		// Generate a stateless actor so that we aren't constantly hitting the dotcom API
		act = NewActor(s, token, dotcom.ProductSubscriptionState{}, s.devLicensesOnly)
	} else {
		act = NewActor(s, token,
			resp.Dotcom.ProductSubscriptionByAccessToken.ProductSubscriptionState,
			s.devLicensesOnly)
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
func NewActor(source *Source, token string, s dotcom.ProductSubscriptionState, devLicensesOnly bool) *actor.Actor {
	var rateLimit actor.RateLimit
	if s.LlmProxyAccess.RateLimit != nil {
		rateLimit = actor.RateLimit{
			Limit:    s.LlmProxyAccess.RateLimit.Limit,
			Interval: time.Duration(s.LlmProxyAccess.RateLimit.IntervalSeconds) * time.Second,
		}
	}

	// In dev mode, only allow dev licenses.
	disabledBecauseOfDevMode := devLicensesOnly &&
		(s.ActiveLicense == nil || s.ActiveLicense.Info == nil || !slices.Contains(s.ActiveLicense.Info.Tags, "dev"))

	now := time.Now()
	return &actor.Actor{
		Key:           token,
		ID:            s.Uuid,
		AccessEnabled: !disabledBecauseOfDevMode && !s.IsArchived && s.LlmProxyAccess.Enabled && rateLimit.IsValid(),
		RateLimit:     rateLimit,
		LastUpdated:   &now,
		Source:        source,
	}
}
