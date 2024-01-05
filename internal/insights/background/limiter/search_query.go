package limiter

import (
	"sync"

	"golang.org/x/time/rate"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
)

var searchOnce sync.Once
var searchLogger log.Logger
var searchLimiter *ratelimit.InstrumentedLimiter

func SearchQueryRate() *ratelimit.InstrumentedLimiter {

	searchOnce.Do(func() {
		searchLogger = log.Scoped("insights.search.ratelimiter")
		defaultRateLimit := rate.Limit(20.0)
		defaultBurst := 20
		getRateLimit := getSearchQueryRateLimit(defaultRateLimit, defaultBurst)
		limiter := rate.NewLimiter(getRateLimit())
		searchLimiter = ratelimit.NewInstrumentedLimiter("QueryRunner", limiter)

		go conf.Watch(func() {
			limit, burst := getRateLimit()
			searchLogger.Info("Updating insights/query-worker ", log.Int("rate limit", int(limit)), log.Int("burst", burst))
			limiter.SetLimit(limit)
			limiter.SetBurst(burst)
		})
	})

	return searchLimiter
}

func getSearchQueryRateLimit(defaultRate rate.Limit, defaultBurst int) func() (rate.Limit, int) {
	return func() (rate.Limit, int) {
		limit := conf.Get().InsightsQueryWorkerRateLimit
		burst := conf.Get().InsightsQueryWorkerRateLimitBurst

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
