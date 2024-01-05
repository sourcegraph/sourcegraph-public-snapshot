package productsubscription

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/gregjones/httpcache"
	"github.com/sourcegraph/log"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/productsubscription"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// SourceVersion should be bumped whenever the format of any cached data in this
// actor source implementation is changed. This effectively expires all entries.
const SourceVersion = "v2"

// product subscription tokens are always a prefix of 4 characters (sgs_ or slk_)
// followed by a 64-character hex-encoded SHA256 hash
const tokenLength = 4 + 64

var (
	minUpdateInterval = 10 * time.Minute

	defaultUpdateInterval = 24 * time.Hour
)

type Source struct {
	log    log.Logger
	cache  httpcache.Cache // cache is expected to be something with automatic TTL
	dotcom graphql.Client

	// internalMode, if true, indicates only dev and internal licenses may use
	// this Cody Gateway instance.
	internalMode bool

	concurrencyConfig codygateway.ActorConcurrencyLimitConfig
}

var _ actor.Source = &Source{}
var _ actor.SourceUpdater = &Source{}
var _ actor.SourceSyncer = &Source{}

func NewSource(logger log.Logger, cache httpcache.Cache, dotcomClient graphql.Client, internalMode bool, concurrencyConfig codygateway.ActorConcurrencyLimitConfig) *Source {
	return &Source{
		log:    logger.Scoped("productsubscriptions"),
		cache:  cache,
		dotcom: dotcomClient,

		internalMode: internalMode,

		concurrencyConfig: concurrencyConfig,
	}
}

func (s *Source) Name() string { return string(codygateway.ActorSourceProductSubscription) }

func (s *Source) Get(ctx context.Context, token string) (*actor.Actor, error) {
	if token == "" {
		return nil, actor.ErrNotFromSource{}
	}

	// NOTE: For back-compat, we support both the old and new token prefixes.
	// However, as we use the token as part of the cache key, we need to be
	// consistent with the prefix we use.
	token = strings.Replace(token, productsubscription.AccessTokenPrefix, license.LicenseKeyBasedAccessTokenPrefix, 1)
	if !strings.HasPrefix(token, license.LicenseKeyBasedAccessTokenPrefix) {
		return nil, actor.ErrNotFromSource{Reason: "unknown token prefix"}
	}

	if len(token) != tokenLength {
		return nil, errors.New("invalid token format")
	}

	span := trace.SpanFromContext(ctx)

	data, hit := s.cache.Get(token)
	if !hit {
		span.SetAttributes(attribute.Bool("actor-cache-miss", true))
		return s.fetchAndCache(ctx, token)
	}

	var act *actor.Actor
	if err := json.Unmarshal(data, &act); err != nil {
		span.SetAttributes(attribute.Bool("actor-corrupted", true))
		sgtrace.Logger(ctx, s.log).Error("failed to unmarshal subscription", log.Error(err))

		// Delete the corrupted record.
		s.cache.Delete(token)

		return s.fetchAndCache(ctx, token)
	}

	if act.LastUpdated != nil && time.Since(*act.LastUpdated) > defaultUpdateInterval {
		span.SetAttributes(attribute.Bool("actor-expired", true))
		return s.fetchAndCache(ctx, token)
	}

	act.Source = s
	return act, nil
}

func (s *Source) Update(ctx context.Context, act *actor.Actor) error {
	if time.Since(*act.LastUpdated) < minUpdateInterval {
		// Last update was too recent - do it later.
		return actor.ErrActorRecentlyUpdated{
			RetryAt: act.LastUpdated.Add(minUpdateInterval),
		}
	}

	_, err := s.fetchAndCache(ctx, act.Key)
	return err
}

// Sync retrieves all known actors from this source and updates its cache.
// All Sync implementations are called periodically - implementations can decide
// to skip syncs if the frequency is too high.
func (s *Source) Sync(ctx context.Context) (seen int, errs error) {
	syncLog := sgtrace.Logger(ctx, s.log)

	resp, err := dotcom.ListProductSubscriptions(ctx, s.dotcom)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			syncLog.Warn("sync context cancelled")
			return seen, nil
		}
		return seen, errors.Wrap(err, "failed to list subscriptions from dotcom")
	}

	for _, sub := range resp.Dotcom.ProductSubscriptions.Nodes {
		for _, token := range sub.SourcegraphAccessTokens {
			select {
			case <-ctx.Done():
				return seen, ctx.Err()
			default:
			}

			act := newActor(s, token, sub.ProductSubscriptionState, s.internalMode, s.concurrencyConfig)
			data, err := json.Marshal(act)
			if err != nil {
				act.Logger(syncLog).Error("failed to marshal actor",
					log.Error(err))
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

func (s *Source) checkAccessToken(ctx context.Context, token string) (*dotcom.CheckAccessTokenResponse, error) {
	resp, err := dotcom.CheckAccessToken(ctx, s.dotcom, token)
	if err == nil {
		return resp, nil
	}

	// Inspect the error to see if it's a list of GraphQL errors.
	gqlerrs, ok := err.(gqlerror.List)
	if !ok {
		return nil, err
	}

	for _, gqlerr := range gqlerrs {
		if gqlerr.Extensions != nil && gqlerr.Extensions["code"] == productsubscription.GQLErrCodeProductSubscriptionNotFound {
			return nil, actor.ErrAccessTokenDenied{
				Source: s.Name(),
				Reason: "associated product subscription not found",
			}
		}
	}
	return nil, err
}

func (s *Source) fetchAndCache(ctx context.Context, token string) (*actor.Actor, error) {
	var act *actor.Actor
	resp, checkErr := s.checkAccessToken(ctx, token)
	if checkErr != nil {
		// Generate a stateless actor so that we aren't constantly hitting the dotcom API
		act = newActor(s, token, dotcom.ProductSubscriptionState{}, s.internalMode, s.concurrencyConfig)
	} else {
		act = newActor(
			s,
			token,
			resp.Dotcom.ProductSubscriptionByAccessToken.ProductSubscriptionState,
			s.internalMode,
			s.concurrencyConfig,
		)
	}

	if data, err := json.Marshal(act); err != nil {
		sgtrace.Logger(ctx, s.log).Error("failed to marshal actor",
			log.Error(err))
	} else {
		s.cache.Set(token, data)
	}

	if checkErr != nil {
		return nil, errors.Wrap(checkErr, "failed to validate access token")
	}
	return act, nil
}

// getSubscriptionAccountName attempts to get the account name from the product
// subscription. It returns an empty string if no account name is available.
func getSubscriptionAccountName(s dotcom.ProductSubscriptionState) string {
	// 1. Check if the special "customer:" tag is present
	if s.ActiveLicense != nil && s.ActiveLicense.Info != nil {
		for _, tag := range s.ActiveLicense.Info.Tags {
			if strings.HasPrefix(tag, "customer:") {
				return strings.TrimPrefix(tag, "customer:")
			}
		}
	}

	// 2. Use the username of the account
	if s.Account != nil && s.Account.Username != "" {
		return s.Account.Username
	}
	return ""
}

// newActor creates an actor from Sourcegraph.com product subscription state.
func newActor(source *Source, token string, s dotcom.ProductSubscriptionState, internalMode bool, concurrencyConfig codygateway.ActorConcurrencyLimitConfig) *actor.Actor {
	name := getSubscriptionAccountName(s)
	if name == "" {
		name = s.Uuid
	}

	// In internal mode, only allow dev and internal licenses.
	disallowedLicense := internalMode &&
		(s.ActiveLicense == nil || s.ActiveLicense.Info == nil ||
			!containsOneOf(s.ActiveLicense.Info.Tags, licensing.DevTag, licensing.InternalTag))

	now := time.Now()
	a := &actor.Actor{
		Key:           token,
		ID:            s.Uuid,
		Name:          name,
		AccessEnabled: !disallowedLicense && !s.IsArchived && s.CodyGatewayAccess.Enabled,
		RateLimits:    map[codygateway.Feature]actor.RateLimit{},
		LastUpdated:   &now,
		Source:        source,
	}

	if rl := s.CodyGatewayAccess.ChatCompletionsRateLimit; rl != nil {
		a.RateLimits[codygateway.FeatureChatCompletions] = actor.NewRateLimitWithPercentageConcurrency(
			int64(rl.Limit),
			time.Duration(rl.IntervalSeconds)*time.Second,
			rl.AllowedModels,
			concurrencyConfig,
		)
	}

	if rl := s.CodyGatewayAccess.CodeCompletionsRateLimit; rl != nil {
		a.RateLimits[codygateway.FeatureCodeCompletions] = actor.NewRateLimitWithPercentageConcurrency(
			int64(rl.Limit),
			time.Duration(rl.IntervalSeconds)*time.Second,
			rl.AllowedModels,
			concurrencyConfig,
		)
	}

	if rl := s.CodyGatewayAccess.EmbeddingsRateLimit; rl != nil {
		a.RateLimits[codygateway.FeatureEmbeddings] = actor.NewRateLimitWithPercentageConcurrency(
			int64(rl.Limit),
			time.Duration(rl.IntervalSeconds)*time.Second,
			rl.AllowedModels,
			// TODO: Once we split interactive and on-interactive, we want to apply
			// stricter limits here than percentage based for this heavy endpoint.
			concurrencyConfig,
		)
	}

	return a
}

func containsOneOf(s []string, needles ...string) bool {
	for _, needle := range needles {
		if slices.Contains(s, needle) {
			return true
		}
	}
	return false
}
