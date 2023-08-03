package ratelimit

import (
	"context"
	"fmt"
	"math"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type handler struct {
	codeHostStore  database.CodeHostStore
	redisKeyPrefix string
	ratelimiter    redispool.RateLimiter
}

func (h *handler) Handle(ctx context.Context, observationCtx *observation.Context) error {
	var err error
	next := int32(-1)
	for next != 0 {
		codeHosts, next2, err := h.codeHostStore.List(ctx, database.ListCodeHostsOpts{
			LimitOffset: database.LimitOffset{
				Limit: 10,
			},
			Cursor: next,
		})
		if err != nil {
			return err
		}
		next = next2
		for _, codeHost := range codeHosts {
			err = h.processCodeHost(ctx, codeHost.URL)
			if err != nil {
				observationCtx.Logger.Error("error setting rate limit configuration", log.String("url", codeHost.URL), log.Error(err))
			}
		}
	}
	return err
}

func (h *handler) processCodeHost(ctx context.Context, codeHostURL string) error {
	apiCap, apiRepInterval, gitCap, gitRepInterval, err := h.getRateLimitConfigsOrDefaults(ctx, codeHostURL)

	// Set API token values
	err = h.ratelimiter.SetTokenBucketReplenishment(ctx, fmt.Sprintf("%s:api_tokens", codeHostURL), apiCap, apiRepInterval)
	// Set Git token values
	err2 := h.ratelimiter.SetTokenBucketReplenishment(ctx, fmt.Sprintf("%s:git_tokens", codeHostURL), gitCap, gitRepInterval)

	return errors.CombineErrors(err, err2)
}

func (h *handler) getRateLimitConfigsOrDefaults(ctx context.Context, codeHostURL string) (apiCap, apiRepInterval, gitCap, gitRepInterval int32, err error) {
	// Retrieve the actual rate limit values from the source of truth (Postgres).
	ch, err := h.codeHostStore.GetByURL(ctx, codeHostURL)
	if err != nil {
		err = errors.Wrapf(err, "rate limit config worker unable to get code host by URL: %s", codeHostURL)
		return
	}

	// Determine the values of the 4 rate limit configurations by using their set value from Postgres or their default value if they are not set.
	if ch.APIRateLimitQuota != nil {
		apiCap = *ch.APIRateLimitQuota
	} else {
		// TODO: is this reasonable? These limits represent remote calls, so we are never reaching these numbers anyways.
		defaultRateLimitAsInt := int32(extsvc.GetRateLimit(ch.Kind))
		if defaultRateLimitAsInt < 0 {
			defaultRateLimitAsInt = math.MaxInt32
		}
		apiCap = defaultRateLimitAsInt
	}

	if ch.APIRateLimitIntervalSeconds != nil {
		apiRepInterval = *ch.APIRateLimitIntervalSeconds
	} else {
		apiRepInterval = int32(3600)
	}

	if ch.GitRateLimitQuota != nil {
		gitCap = *ch.GitRateLimitQuota
	} else {
		siteCfg := conf.Get()
		if siteCfg.GitMaxCodehostRequestsPerSecond != nil {
			gitCap = int32(*siteCfg.GitMaxCodehostRequestsPerSecond)
		} else {
			gitCap = math.MaxInt32
		}
	}

	if ch.GitRateLimitIntervalSeconds != nil {
		gitRepInterval = *ch.GitRateLimitIntervalSeconds
	} else {
		gitRepInterval = int32(1)
	}
	return
}
