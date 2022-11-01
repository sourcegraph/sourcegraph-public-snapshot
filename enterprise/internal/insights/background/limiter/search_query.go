package limiter

import (
	"sync"

	"github.com/sourcegraph/log"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
)

var searchOnce sync.Once
var searchLogger log.Logger
var searchLimiter *ratelimit.InstrumentedLimiter

func SearchQueryRate() *ratelimit.InstrumentedLimiter {

	searchOnce.Do(func() {
		searchLogger = log.Scoped("insights.search.ratelimiter", "")
		defaultRateLimit := rate.Limit(20.0)
		getRateLimit := getSearchQueryRateLimit(defaultRateLimit)
		searchLimiter = ratelimit.NewInstrumentedLimiter("QueryRunner", rate.NewLimiter(getRateLimit(), 1))

		go conf.Watch(func() {
			val := getRateLimit()
			searchLogger.Info("Updating insights/query-worker rate limit", log.Int("value", int(val)))
			searchLimiter.SetLimit(val)
		})
	})

	return searchLimiter
}

func getSearchQueryRateLimit(defaultValue rate.Limit) func() rate.Limit {
	return func() rate.Limit {
		val := conf.Get().InsightsQueryWorkerRateLimit

		var result rate.Limit
		if val == nil {
			result = defaultValue
		} else {
			result = rate.Limit(*val)
		}

		return result
	}
}
