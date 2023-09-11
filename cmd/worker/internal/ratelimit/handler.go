package ratelimit

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
)

var _ goroutine.Handler = &handler{}

type handler struct {
	externalServiceStore database.ExternalServiceStore
	logger               log.Logger
	newRateLimiterFunc   func(bucketName string) ratelimit.GlobalLimiter
}

func (h *handler) Handle(ctx context.Context) (err error) {
	defer func() {
		// Be very vocal about these issues.
		if err != nil {
			h.logger.Error("failed to sync rate limit configs to redis", log.Error(err))
		}
	}()

	var defaultGitQuota int32 = -1 // rate.Inf
	siteCfg := conf.Get()
	if siteCfg.GitMaxCodehostRequestsPerSecond != nil {
		defaultGitQuota = int32(*siteCfg.GitMaxCodehostRequestsPerSecond)
	}

	gitRL := h.newRateLimiterFunc(ratelimit.GitRPSLimiterBucketName)
	if err := gitRL.SetTokenBucketConfig(ctx, defaultGitQuota, time.Second); err != nil {
		return err
	}

	svcs, err := h.externalServiceStore.List(ctx, database.ExternalServicesListOptions{})
	if err != nil {
		return err
	}

	return syncServices(ctx, svcs, h.newRateLimiterFunc)
}
