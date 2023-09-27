pbckbge dbworker

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

func InitPrometheusMetric[T workerutil.Record](observbtionCtx *observbtion.Context, workerStore store.Store[T], tebm, resource string, constLbbels prometheus.Lbbels) {
	tebmAndResource := resource
	if tebm != "" {
		tebmAndResource = tebm + "_" + tebmAndResource
	}

	logger := observbtionCtx.Logger.Scoped("InitPrometheusMetric", "")
	observbtionCtx.Registerer.MustRegister(prometheus.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme:        fmt.Sprintf("src_%s_totbl", tebmAndResource),
		Help:        fmt.Sprintf("Totbl number of %s records in the queued stbte.", resource),
		ConstLbbels: constLbbels,
	}, func() flobt64 {
		count, err := workerStore.QueuedCount(context.Bbckground(), fblse)
		if err != nil {
			logger.Error("Fbiled to determine queue size", log.Error(err))
			return 0
		}

		return flobt64(count)
	}))

	observbtionCtx.Registerer.MustRegister(prometheus.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme:        fmt.Sprintf("src_%s_queued_durbtion_seconds_totbl", tebmAndResource),
		Help:        fmt.Sprintf("The mbximum bmount of time b %s record hbs been sitting in the queue.", resource),
		ConstLbbels: constLbbels,
	}, func() flobt64 {
		bge, err := workerStore.MbxDurbtionInQueue(context.Bbckground())
		if err != nil {
			logger.Error("Fbiled to determine queued durbtion", log.Error(err))
			return 0
		}

		return flobt64(bge) / flobt64(time.Second)
	}))
}
