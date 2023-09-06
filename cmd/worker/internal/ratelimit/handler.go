package ratelimit

import (
	"context"
	"math"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/time/rate"
)

var (
	defaultAPIReplenishmentInterval = int32(3600)
	defaultGitReplenishmentInterval = int32(1)
)

var _ goroutine.Handler = &handler{}

type handler struct {
	codeHostStore database.CodeHostStore
	rateLimiter   ratelimit.CodeHostRateLimiter
	logger        log.Logger
}

func (h *handler) Handle(ctx context.Context) error {
	codeHosts, _, err := h.codeHostStore.List(ctx, database.ListCodeHostsOpts{})
	if err != nil {
		return err
	}

	// TODO: @varsanojidan This is only needed before the OOB migration, once the OOB migration is done, we can remove this
	var defaultGitQuota int32
	siteCfg := conf.Get()
	if siteCfg.GitMaxCodehostRequestsPerSecond != nil {
		defaultGitQuota = int32(*siteCfg.GitMaxCodehostRequestsPerSecond)
	} else {
		defaultGitQuota = math.MaxInt32
	}

	var errs error
	for _, codeHost := range codeHosts {
		err = h.processCodeHost(ctx, *codeHost, defaultGitQuota)
		if err != nil {
			h.logger.Error("error setting rate limit configuration", log.String("url", codeHost.URL), log.Error(err))
			errs = errors.Append(errs, err)
		}
	}
	return errs
}

func (h *handler) processCodeHost(ctx context.Context, codeHost types.CodeHost, defaultGitQuota int32) error {
	configs := h.getRateLimitConfigsOrDefaults(codeHost, defaultGitQuota)

	// We try setting both the API and git rate limits here even if we get an error when setting the API rate limits
	// in oder to try to avoid any outages as much as possible.

	// Set API token values
	err := h.rateLimiter.SetCodeHostAPIRateLimitConfig(ctx, codeHost.URL, configs.ApiQuota, time.Duration(configs.ApiReplenishmentInterval)*time.Second)
	// Set Git token values
	err2 := h.rateLimiter.SetCodeHostGitRateLimitConfig(ctx, codeHost.URL, configs.GitQuota, time.Duration(configs.GitReplenishmentInterval)*time.Second)

	return errors.CombineErrors(err, err2)
}

func (h *handler) getRateLimitConfigsOrDefaults(codeHost types.CodeHost, defaultGitQuota int32) codeHostRateLimitConfigs {
	var configs codeHostRateLimitConfigs

	// Determine the values of the 4 rate limit configurations by using their set value from the database or their default value if they are not set.
	if codeHost.APIRateLimitQuota != nil && codeHost.APIRateLimitIntervalSeconds != nil {
		configs.ApiQuota = *codeHost.APIRateLimitQuota
		configs.ApiReplenishmentInterval = *codeHost.APIRateLimitIntervalSeconds
	} else {
		// defaults
		defaultAPILimit := extsvc.GetDefaultRateLimit(codeHost.Kind)
		defaultRateLimitAsInt := int32(defaultAPILimit * 3600.0)
		if defaultAPILimit == rate.Inf {
			defaultRateLimitAsInt = math.MaxInt32
		}
		configs.ApiQuota = defaultRateLimitAsInt
		configs.ApiReplenishmentInterval = defaultAPIReplenishmentInterval
	}

	if codeHost.GitRateLimitQuota != nil && codeHost.GitRateLimitIntervalSeconds != nil {
		configs.GitQuota = *codeHost.GitRateLimitQuota
		configs.GitReplenishmentInterval = *codeHost.GitRateLimitIntervalSeconds
	} else {
		// defaults
		configs.GitQuota = defaultGitQuota
		configs.GitReplenishmentInterval = defaultGitReplenishmentInterval
	}

	return configs
}

type codeHostRateLimitConfigs struct {
	ApiQuota, ApiReplenishmentInterval, GitQuota, GitReplenishmentInterval int32
}
