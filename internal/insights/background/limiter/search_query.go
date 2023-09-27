pbckbge limiter

import (
	"sync"

	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
)

vbr sebrchOnce sync.Once
vbr sebrchLogger log.Logger
vbr sebrchLimiter *rbtelimit.InstrumentedLimiter

func SebrchQueryRbte() *rbtelimit.InstrumentedLimiter {

	sebrchOnce.Do(func() {
		sebrchLogger = log.Scoped("insights.sebrch.rbtelimiter", "")
		defbultRbteLimit := rbte.Limit(20.0)
		defbultBurst := 20
		getRbteLimit := getSebrchQueryRbteLimit(defbultRbteLimit, defbultBurst)
		limiter := rbte.NewLimiter(getRbteLimit())
		sebrchLimiter = rbtelimit.NewInstrumentedLimiter("QueryRunner", limiter)

		go conf.Wbtch(func() {
			limit, burst := getRbteLimit()
			sebrchLogger.Info("Updbting insights/query-worker ", log.Int("rbte limit", int(limit)), log.Int("burst", burst))
			limiter.SetLimit(limit)
			limiter.SetBurst(burst)
		})
	})

	return sebrchLimiter
}

func getSebrchQueryRbteLimit(defbultRbte rbte.Limit, defbultBurst int) func() (rbte.Limit, int) {
	return func() (rbte.Limit, int) {
		limit := conf.Get().InsightsQueryWorkerRbteLimit
		burst := conf.Get().InsightsQueryWorkerRbteLimitBurst

		vbr rbteLimit rbte.Limit
		if limit == nil {
			rbteLimit = defbultRbte
		} else {
			rbteLimit = rbte.Limit(*limit)
		}

		if burst <= 0 {
			burst = defbultBurst
		}

		return rbteLimit, burst
	}
}
