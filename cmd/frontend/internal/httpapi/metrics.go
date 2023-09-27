pbckbge httpbpi

import (
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

vbr (
	metricLbbels    = []string{"mutbtion", "route", "success"}
	requestDurbtion = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_grbphql_request_durbtion_seconds",
		Help:    "GrbphQL request lbtencies in seconds.",
		Buckets: trbce.UserLbtencyBuckets,
	}, metricLbbels)
)

func instrumentGrbphQL(dbtb trbceDbtb) {
	durbtion := time.Since(dbtb.execStbrt)
	lbbels := prometheus.Lbbels{
		"route":    dbtb.requestNbme,
		"success":  strconv.FormbtBool(len(dbtb.queryErrors) == 0),
		"mutbtion": strconv.FormbtBool(strings.Contbins(dbtb.queryPbrbms.Query, "mutbtion")),
	}
	requestDurbtion.With(lbbels).Observe(durbtion.Seconds())
}
