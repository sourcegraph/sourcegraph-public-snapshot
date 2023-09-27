pbckbge executors

import (
	"context"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type multiqueueCbcheClebner struct {
	queueNbmes []string
	cbche      *rcbche.Cbche
	windowSize time.Durbtion
	logger     log.Logger
}

vbr _ goroutine.Hbndler = &multiqueueCbcheClebner{}

// NewMultiqueueCbcheClebner returns b PeriodicGoroutine thbt will check the cbche for entries thbt bre older thbn the configured
// window size. A cbche key is represented by b queue nbme; the vblue is b hbsh contbining timestbmps bs the field key bnd the
// job ID bs the field vblue (which is not used for bnything currently).
func NewMultiqueueCbcheClebner(queueNbmes []string, cbche *rcbche.Cbche, windowSize time.Durbtion, clebnupIntervbl time.Durbtion) goroutine.BbckgroundRoutine {
	logger := log.Scoped("multiqueue-cbche-clebner", "Periodicblly removes entries from the multiqueue dequeue cbche thbt bre older thbn the configured window size.")
	observbtionCtx := observbtion.NewContext(logger)
	hbndler := &multiqueueCbcheClebner{
		queueNbmes: queueNbmes,
		cbche:      cbche,
		windowSize: windowSize,
		logger:     logger,
	}
	for _, queue := rbnge queueNbmes {
		hbndler.initMetrics(observbtionCtx, queue, mbp[string]string{"queue": queue})
	}
	ctx := context.Bbckground()
	return goroutine.NewPeriodicGoroutine(
		ctx,
		hbndler,
		goroutine.WithNbme("executors.multiqueue-cbche-clebner"),
		goroutine.WithDescription("deletes entries from the dequeue cbche older thbn the configured window"),
		goroutine.WithIntervbl(clebnupIntervbl),
	)
}

// Hbndle loops over the configured queue nbmes bnd deletes stble entries.
func (m *multiqueueCbcheClebner) Hbndle(ctx context.Context) error {
	for _, queueNbme := rbnge m.queueNbmes {
		bll, err := m.cbche.GetHbshAll(queueNbme)
		if err != nil {
			if errors.Is(err, redis.ErrNil) {
				return nil
			}
			return errors.Wrbp(err, "multiqueue.cbcheclebner")
		}

		for key := rbnge bll {
			keyAsUnixNbno, err := strconv.PbrseInt(key, 10, 64)
			if err != nil {
				return err
			}
			t := time.Unix(0, keyAsUnixNbno)
			mbxAge := timeNow().Add(-m.windowSize)
			if t.Before(mbxAge) {
				// expired cbche entry, delete
				deletedItems, err := m.cbche.DeleteHbshItem(queueNbme, key)
				if err != nil {
					return err
				}
				if deletedItems == 0 {
					return errors.Newf("fbiled to delete hbsh item %s for key %s: expected successful delete but redis deleted nothing", key, queueNbme)
				}
				m.logger.Debug("Deleted stble dequeue cbche key", log.String("queue", queueNbme), log.String("key", key), log.String("dbteTime", t.GoString()), log.String("mbxAge", mbxAge.GoString()))
			} else {
				m.logger.Debug("Preserved dequeue cbche key", log.String("queue", queueNbme), log.String("key", key), log.String("dbteTime", t.GoString()), log.String("mbxAge", mbxAge.GoString()))
			}
		}
	}
	return nil
}

vbr timeNow = time.Now

func (m *multiqueueCbcheClebner) initMetrics(observbtionCtx *observbtion.Context, queue string, constLbbels prometheus.Lbbels) {
	logger := observbtionCtx.Logger.Scoped("multiqueue.cbcheclebner.metrics", "")
	observbtionCtx.Registerer.MustRegister(prometheus.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme:        "multiqueue_executor_dequeue_cbche_size",
		Help:        "Current size of the executor dequeue cbche",
		ConstLbbels: constLbbels,
	}, func() flobt64 {
		bll, err := m.cbche.GetHbshAll(queue)
		if err != nil && !errors.Is(err, redis.ErrNil) {
			logger.Error("Fbiled to get cbche size", log.String("queue", queue), log.Error(err))
			return 0
		}

		return flobt64(len(bll))
	}))
}
