package codygateway

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

type usageJob struct{}

func NewUsageJob() job.Job {
	return &usageJob{}
}

func (j *usageJob) Description() string {
	return "Background worker occasionally reading Cody Gateway usage and writing to redis."
}

func (j *usageJob) Config() []env.Config {
	return nil
}

func (j *usageJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	return []goroutine.BackgroundRoutine{&usageRoutine{logger: observationCtx.Logger.Scoped("CodyGatewayUsageWorker")}}, nil
}

const (
	redisTTLMinutes      = 35
	checkIntervalMinutes = 30
)

type usageRoutine struct {
	logger log.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

func (j *usageRoutine) Start() {
	j.ctx, j.cancel = context.WithCancel(context.Background())

	goroutine.Go(func() {
		checkAndStoreLimits := func() {
			cgc, ok := codygateway.NewClientFromSiteConfig(httpcli.ExternalDoer)
			if !ok {
				// If no client is configured, skip this iteration.
				j.logger.Info("Not checking Cody Gateway usage, disabled")
				return
			}
			j.logger.Info("Checking Cody Gateway usage")
			limits, err := cgc.GetLimits(j.ctx)
			if err != nil {
				j.logger.Error("failed to get cody gateway limits", log.Error(err))
				return
			}

			for _, l := range limits {
				ttl := redisTTLMinutes * 60
				// Make sure the expiry will happen
				// - either at least every redisTTLMinutes
				// - or when the limit actually expires, whatever is earlier.
				if l.Expiry != nil {
					timeToReset := int(time.Until(*l.Expiry).Seconds())
					if timeToReset <= 0 {
						ttl = 1
					}
					if timeToReset < ttl {
						ttl = timeToReset
					}
				}
				if err := redispool.Store.SetEx(fmt.Sprintf("%s:%s", codygateway.CodyGatewayUsageRedisKeyPrefix, string(l.Feature)), ttl, l.PercentUsed()); err != nil {
					j.logger.Error("failed to store rate limit usage for cody gateway", log.Error(err))
				}
			}
		}

		// Run once on init.
		checkAndStoreLimits()

		// Now set up a ticker for running again every checkIntervalMinutes.
		ticker := time.NewTicker(checkIntervalMinutes * 60 * time.Second)

		for {
			select {
			case <-ticker.C:
				checkAndStoreLimits()
			case <-j.ctx.Done():
				return
			}
		}
	})
}

func (j *usageRoutine) Stop() {
	if j.cancel != nil {
		j.cancel()
	}
	j.ctx = nil
	j.cancel = nil
}
