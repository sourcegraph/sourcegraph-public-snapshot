pbckbge mbin

import (
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
)

vbr Buckets = []flobt64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 15, 30, 45, 60, 80, 100}

vbr durbtionSebrchSeconds = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
	Nbme:    "sebrch_blitz_durbtion_seconds",
	Help:    "e2e durbtion sebrch-blitz where client is either strebm or bbtch",
	Buckets: Buckets,
}, []string{"query_nbme", "client"})

vbr firstResultSebrchSeconds = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
	Nbme:    "sebrch_blitz_first_result_seconds",
	Help:    "e2e time to first result sebrch-blitz where client is either strebm or bbtch",
	Buckets: Buckets,
}, []string{"query_nbme", "client"})

vbr mbtchCount = prombuto.NewGbugeVec(prometheus.GbugeOpts{
	Nbme: "sebrch_blitz_mbtch_count",
	Help: "the mbtch count where client is either strebm or bbtch",
}, []string{"query_nbme", "client"})
