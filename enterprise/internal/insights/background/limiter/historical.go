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
		getRateLimit := getHistoricalWorkerRateLimit(defaultRateLimit)
		historicalLimiter = ratelimit.NewInstrumentedLimiter("HistoricalInsight", rate.NewLimiter(getRateLimit(), 1))
		go conf.Watch(func() {
			val := getRateLimit()
			historicalLogger.Info("Updating insights/historical rate limit", log.Int("value", int(val)))
			historicalLimiter.SetLimit(val)
		})
	})

	return historicalLimiter
}

func getHistoricalWorkerRateLimit(defaultValue rate.Limit) func() rate.Limit {
	return func() rate.Limit {
		val := conf.Get().InsightsHistoricalWorkerRateLimit

		var result rate.Limit
		if val == nil {
			result = defaultValue
		} else {
			result = rate.Limit(*val)
		}

		return result
	}
}
