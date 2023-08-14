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
	"golang.org/x/time/rate"
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
	configs, err := h.getRateLimitConfigsOrDefaults(ctx, codeHostURL)
	if err != nil {
		return err
	}
	// Set API token values
	err = h.ratelimiter.SetTokenBucketReplenishment(ctx, fmt.Sprintf("%s:%s", codeHostURL, redispool.CodeHostAPITokenBucketSuffix), configs.ApiQuota, configs.ApiReplenishmentInterval)
	// Set Git token values
	err2 := h.ratelimiter.SetTokenBucketReplenishment(ctx, fmt.Sprintf("%s:%s", codeHostURL, redispool.CodeHostGitTokenBucketSuffix), configs.GitQuota, configs.GitReplenishmentInterval)

	return errors.CombineErrors(err, err2)
}

func (h *handler) getRateLimitConfigsOrDefaults(ctx context.Context, codeHostURL string) (CodeHostRateLimitConfigs, error) {
	var configs CodeHostRateLimitConfigs
	// Retrieve the actual rate limit values from the source of truth (database).
	codeHost, err := h.codeHostStore.GetByURL(ctx, codeHostURL)
	if err != nil {
		return CodeHostRateLimitConfigs{}, errors.Wrapf(err, "rate limit config worker unable to get code host by URL: %s", codeHostURL)
	}

	// Determine the values of the 4 rate limit configurations by using their set value from the database or their default value if they are not set.
	if codeHost.APIRateLimitQuota != nil {
		configs.ApiQuota = *codeHost.APIRateLimitQuota
	} else {
		defaultRateLimitAsInt := int32(extsvc.GetDefaultRateLimit(codeHost.Kind))
		// Basically only happens if the rate limit is set to rate.Inf
		if extsvc.GetDefaultRateLimit(codeHost.Kind) > rate.Limit(math.MaxInt32) {
			defaultRateLimitAsInt = math.MaxInt32
		}
		configs.ApiQuota = defaultRateLimitAsInt
	}

	if codeHost.APIRateLimitIntervalSeconds != nil {
		configs.ApiReplenishmentInterval = *codeHost.APIRateLimitIntervalSeconds
	} else {
		configs.ApiReplenishmentInterval = int32(3600)
	}

	if codeHost.GitRateLimitQuota != nil {
		configs.GitQuota = *codeHost.GitRateLimitQuota
	} else {
		siteCfg := conf.Get()
		if siteCfg.GitMaxCodehostRequestsPerSecond != nil {
			configs.GitQuota = int32(*siteCfg.GitMaxCodehostRequestsPerSecond)
		} else {
			configs.GitQuota = math.MaxInt32
		}
	}

	if codeHost.GitRateLimitIntervalSeconds != nil {
		configs.GitReplenishmentInterval = *codeHost.GitRateLimitIntervalSeconds
	} else {
		configs.GitReplenishmentInterval = int32(1)
	}
	return configs, nil
}

type CodeHostRateLimitConfigs struct {
	ApiQuota, ApiReplenishmentInterval, GitQuota, GitReplenishmentInterval int32
}
