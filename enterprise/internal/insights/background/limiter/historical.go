package limiter

import (
	"sync"

	"github.com/sourcegraph/log"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
)

var historicalOnce sync.Once
var historicalLogger log.Logger
var historicalLimiter *ratelimit.InstrumentedLimiter

func HistoricalWorkRate() *ratelimit.InstrumentedLimiter {

	historicalOnce.Do(func() {
		historicalLogger = log.Scoped("insights.historical.ratelimiter", "")
		defaultRateLimit := rate.Limit(20.0)
		defaultBurst := 20
		getRateLimit := getHistoricalWorkerRateLimit(defaultRateLimit, defaultBurst)
		historicalLimiter = ratelimit.NewInstrumentedLimiter("HistoricalInsight", rate.NewLimiter(getRateLimit()))

		go conf.Watch(func() {
			limit, burst := getRateLimit()
			historicalLogger.Info("Updating insights/historical rate limit", log.Int("rate limit", int(limit)), log.Int("burst", burst))
			historicalLimiter.SetLimit(limit)
			historicalLimiter.SetBurst(burst)
		})
	})

	return historicalLimiter
}

func getHistoricalWorkerRateLimit(defaultRate rate.Limit, defaultBurst int) func() (rate.Limit, int) {
	return func() (rate.Limit, int) {
		limit := conf.Get().InsightsHistoricalWorkerRateLimit
		burst := conf.Get().InsightsHistoricalWorkerRateLimitBurst

		var rateLimit rate.Limit
		if limit == nil {
			rateLimit = defaultRate
		} else {
			rateLimit = rate.Limit(*limit)
		}

		if burst <= 0 {
			burst = defaultBurst
		}

		return rateLimit, burst
	}
}
