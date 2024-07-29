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
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	return []goroutine.BackgroundRoutine{
		newCodyGatewayUsageRoutine(observationCtx),
	}, nil
}

func newCodyGatewayUsageRoutine(observationCtx *observation.Context) goroutine.BackgroundRoutine {
	handler := goroutine.HandlerFunc(func(ctx context.Context) error {
		cgc, ok := codygateway.NewClientFromSiteConfig(httpcli.ExternalDoer)
		if !ok {
			// If no client is configured, skip this iteration.
			observationCtx.Logger.Info("Not checking Cody Gateway usage, disabled")
			return nil
		}

		observationCtx.Logger.Info("Checking Cody Gateway usage")
		limits, err := cgc.GetLimits(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get cody gateway limits")
		}

		var errs error
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
				observationCtx.Logger.Error("failed to store rate limit usage for cody gateway", log.Error(err))
				errs = errors.Append(errs, err)
			}
		}

		return errs
	})

	operation := observationCtx.Operation(observation.Op{
		Name: "cody_gateway.usage.run",
		Metrics: metrics.NewREDMetrics(
			observationCtx.Registerer,
			"cody_gateway_usage_worker",
			metrics.WithCountHelp("Total number of cody gateway usage worker executions"),
		),
	})

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		handler,
		goroutine.WithName("CodyGatewayUsageWorker"),
		goroutine.WithDescription("Fetches current usage data from Cody Gateway to render alerts about LLM token exhaustion"),
		goroutine.WithInterval(checkIntervalMinutes*60*time.Second),
		goroutine.WithOperation(operation),
	)
}

const (
	redisTTLMinutes      = 35
	checkIntervalMinutes = 30
)
