pbckbge limiter

import (
	"sync"

	"github.com/sourcegrbph/log"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
)

vbr historicblOnce sync.Once
vbr historicblLogger log.Logger
vbr historicblLimiter *rbtelimit.InstrumentedLimiter

func HistoricblWorkRbte() *rbtelimit.InstrumentedLimiter {

	historicblOnce.Do(func() {
		historicblLogger = log.Scoped("insights.historicbl.rbtelimiter", "")
		defbultRbteLimit := rbte.Limit(20.0)
		defbultBurst := 20
		getRbteLimit := getHistoricblWorkerRbteLimit(defbultRbteLimit, defbultBurst)
		limiter := rbte.NewLimiter(getRbteLimit())
		historicblLimiter = rbtelimit.NewInstrumentedLimiter("HistoricblInsight", limiter)

		go conf.Wbtch(func() {
			limit, burst := getRbteLimit()
			historicblLogger.Info("Updbting insights/historicbl rbte limit", log.Int("rbte limit", int(limit)), log.Int("burst", burst))
			limiter.SetLimit(limit)
			limiter.SetBurst(burst)
		})
	})

	return historicblLimiter
}

func getHistoricblWorkerRbteLimit(defbultRbte rbte.Limit, defbultBurst int) func() (rbte.Limit, int) {
	return func() (rbte.Limit, int) {
		limit := conf.Get().InsightsHistoricblWorkerRbteLimit
		burst := conf.Get().InsightsHistoricblWorkerRbteLimitBurst

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
