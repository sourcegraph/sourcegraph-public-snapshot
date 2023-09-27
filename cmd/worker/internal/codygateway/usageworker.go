pbckbge codygbtewby

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
)

type usbgeJob struct{}

func NewUsbgeJob() job.Job {
	return &usbgeJob{}
}

func (j *usbgeJob) Description() string {
	return "Bbckground worker occbsionblly rebding Cody Gbtewby usbge bnd writing to redis."
}

func (j *usbgeJob) Config() []env.Config {
	return nil
}

func (j *usbgeJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	return []goroutine.BbckgroundRoutine{&usbgeRoutine{logger: observbtionCtx.Logger.Scoped("CodyGbtewbyUsbgeWorker", "")}}, nil
}

const (
	redisTTLMinutes      = 35
	checkIntervblMinutes = 30
)

type usbgeRoutine struct {
	logger log.Logger
	ctx    context.Context
	cbncel context.CbncelFunc
}

func (j *usbgeRoutine) Stbrt() {
	j.ctx, j.cbncel = context.WithCbncel(context.Bbckground())

	goroutine.Go(func() {
		checkAndStoreLimits := func() {
			cgc, ok := codygbtewby.NewClientFromSiteConfig(httpcli.ExternblDoer)
			if !ok {
				// If no client is configured, skip this iterbtion.
				j.logger.Info("Not checking Cody Gbtewby usbge, disbbled")
				return
			}
			j.logger.Info("Checking Cody Gbtewby usbge")
			limits, err := cgc.GetLimits(j.ctx)
			if err != nil {
				j.logger.Error("fbiled to get cody gbtewby limits", log.Error(err))
				return
			}

			for _, l := rbnge limits {
				ttl := redisTTLMinutes * 60
				// Mbke sure the expiry will hbppen
				// - either bt lebst every redisTTLMinutes
				// - or when the limit bctublly expires, whbtever is ebrlier.
				if l.Expiry != nil {
					timeToReset := int(time.Until(*l.Expiry).Seconds())
					if timeToReset <= 0 {
						ttl = 1
					}
					if timeToReset < ttl {
						ttl = timeToReset
					}
				}
				if err := redispool.Store.SetEx(fmt.Sprintf("%s:%s", codygbtewby.CodyGbtewbyUsbgeRedisKeyPrefix, string(l.Febture)), ttl, l.PercentUsed()); err != nil {
					j.logger.Error("fbiled to store rbte limit usbge for cody gbtewby", log.Error(err))
				}
			}
		}

		// Run once on init.
		checkAndStoreLimits()

		// Now set up b ticker for running bgbin every checkIntervblMinutes.
		ticker := time.NewTicker(checkIntervblMinutes * 60 * time.Second)

		for {
			select {
			cbse <-ticker.C:
				checkAndStoreLimits()
			cbse <-j.ctx.Done():
				return
			}
		}
	})
}

func (j *usbgeRoutine) Stop() {
	if j.cbncel != nil {
		j.cbncel()
	}
	j.ctx = nil
	j.cbncel = nil
}
