package ratelimit

import (
	"context"
	"math"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/time/rate"
)

var (
	defaultAPIReplenishmentInterval = int32(3600)
	defaultGitReplenishmentInterval = int32(1)
)

type hdlr struct {
	codeHostStore  database.CodeHostStore
	ratelimiter    ratelimit.CodeHostRateLimiter
	observationCtx *observation.Context
}

type handler interface {
	Handle(ctx context.Context) error
}

func newHandler(observationCtx *observation.Context, codeHostStore database.CodeHostStore, ratelimiter ratelimit.CodeHostRateLimiter) handler {
	return &hdlr{
		observationCtx: observationCtx,
		codeHostStore:  codeHostStore,
		ratelimiter:    ratelimiter,
	}
}

func (h *hdlr) Handle(ctx context.Context) error {
	codeHosts, _, err := h.codeHostStore.List(ctx, database.ListCodeHostsOpts{})
	if err != nil {
		return err
	}

	// TODO: @varsanojidan This is only needed before the OOB migration, once the OOB migration is done, we can remove this
	var fallbackGitQuota int32
	siteCfg := conf.Get()
	if siteCfg.GitMaxCodehostRequestsPerSecond != nil {
		fallbackGitQuota = int32(*siteCfg.GitMaxCodehostRequestsPerSecond)
	} else {
		fallbackGitQuota = math.MaxInt32
	}

	var errs error
	for _, codeHost := range codeHosts {
		err = h.processCodeHost(ctx, *codeHost, fallbackGitQuota)
		if err != nil {
			h.observationCtx.Logger.Error("error setting rate limit configuration", log.String("url", codeHost.URL), log.Error(err))
			errs = errors.Append(errs, err)
		}
	}
	return errs
}

func (h *hdlr) processCodeHost(ctx context.Context, codeHost types.CodeHost, fallbackGitQuota int32) error {
	configs, err := h.getRateLimitConfigsOrDefaults(codeHost, fallbackGitQuota)
	if err != nil {
		return err
	}

	// We try setting both the API and git rate limits here even if we get an error when setting the API rate limits
	// in oder to try to avoid any outages as much as possible.

	// Set API token values
	err = h.ratelimiter.SetCodeHostAPIRateLimitConfig(ctx, codeHost.URL, configs.ApiQuota, configs.ApiReplenishmentInterval)
	// Set Git token values
	err2 := h.ratelimiter.SetCodeHostGitRateLimitConfig(ctx, codeHost.URL, configs.GitQuota, configs.GitReplenishmentInterval)

	return errors.CombineErrors(err, err2)
}

func (h *hdlr) getRateLimitConfigsOrDefaults(codeHost types.CodeHost, fallbackGitQuota int32) (codeHostRateLimitConfigs, error) {
	var configs codeHostRateLimitConfigs

	// Determine the values of the 4 rate limit configurations by using their set value from the database or their default value if they are not set.
	isDefaultAPILimit := true
	if codeHost.APIRateLimitQuota != nil {
		configs.ApiQuota = *codeHost.APIRateLimitQuota
		isDefaultAPILimit = false
	} else {
		defaultAPILimit := extsvc.GetDefaultRateLimit(codeHost.Kind)
		defaultRateLimitAsInt := int32(defaultAPILimit)
		// Basically only happens if the rate limit is set to rate.Inf
		if defaultAPILimit > rate.Limit(math.MaxInt32) {
			defaultRateLimitAsInt = math.MaxInt32
		}
		configs.ApiQuota = defaultRateLimitAsInt
	}

	if !isDefaultAPILimit && codeHost.APIRateLimitIntervalSeconds != nil {
		configs.ApiReplenishmentInterval = *codeHost.APIRateLimitIntervalSeconds
	} else {
		configs.ApiReplenishmentInterval = defaultAPIReplenishmentInterval
	}

	defaultGitLimit := true
	if codeHost.GitRateLimitQuota != nil {
		configs.GitQuota = *codeHost.GitRateLimitQuota
		defaultGitLimit = false
	} else {
		configs.GitQuota = fallbackGitQuota
	}

	if !defaultGitLimit && codeHost.GitRateLimitIntervalSeconds != nil {
		configs.GitReplenishmentInterval = *codeHost.GitRateLimitIntervalSeconds
	} else {
		configs.GitReplenishmentInterval = defaultGitReplenishmentInterval
	}
	return configs, nil
}

type codeHostRateLimitConfigs struct {
	ApiQuota, ApiReplenishmentInterval, GitQuota, GitReplenishmentInterval int32
}
