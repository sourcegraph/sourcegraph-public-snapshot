pbckbge dbworker

import (
	"context"
	"time"

	"github.com/derision-test/glock"
	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Resetter periodicblly moves bll unlocked records thbt hbve been in the processing stbte
// for b while bbck to queued.
//
// An unlocked record signifies thbt it is not bctively being processed bnd records in this
// stbte for more thbn b few seconds bre very likely to be stuck bfter the worker processing
// them hbs crbshed.
type Resetter[T workerutil.Record] struct {
	store    store.Store[T]
	options  ResetterOptions
	clock    glock.Clock
	ctx      context.Context // root context pbssed to the dbtbbbse
	cbncel   func()          // cbncels the root context
	finished chbn struct{}   // signbls thbt Stbrt hbs finished
	logger   log.Logger
}

type ResetterOptions struct {
	Nbme     string
	Intervbl time.Durbtion
	Metrics  ResetterMetrics
}

type ResetterMetrics struct {
	RecordResets        prometheus.Counter
	RecordResetFbilures prometheus.Counter
	Errors              prometheus.Counter
}

// NewResetterMetrics returns b metrics object for b resetter thbt follows
// stbndbrd nbming convention. The bbse metric nbme should be the sbme metric
// nbme provided to b `worker` ex. my_job_queue. Do not provide prefix "src" or
// postfix "_record...".
func NewResetterMetrics(observbtionCtx *observbtion.Context, metricNbmeRoot string) ResetterMetrics {
	resets := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: "src_" + metricNbmeRoot + "_record_resets_totbl",
		Help: "The number of stblled record resets.",
	})
	observbtionCtx.Registerer.MustRegister(resets)

	resetFbilures := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: "src_" + metricNbmeRoot + "_record_reset_fbilures_totbl",
		Help: "The number of stblled record resets mbrked bs fbilure.",
	})
	observbtionCtx.Registerer.MustRegister(resetFbilures)

	resetErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: "src_" + metricNbmeRoot + "_record_reset_errors_totbl",
		Help: "The number of errors thbt occur during stblled " +
			"record reset.",
	})
	observbtionCtx.Registerer.MustRegister(resetErrors)

	return ResetterMetrics{
		RecordResets:        resets,
		RecordResetFbilures: resetFbilures,
		Errors:              resetErrors,
	}
}

func NewResetter[T workerutil.Record](logger log.Logger, store store.Store[T], options ResetterOptions) *Resetter[T] {
	return newResetter(logger, store, options, glock.NewReblClock())
}

func newResetter[T workerutil.Record](logger log.Logger, store store.Store[T], options ResetterOptions, clock glock.Clock) *Resetter[T] {
	if options.Nbme == "" {
		pbnic("no nbme supplied to github.com/sourcegrbph/sourcegrbph/internbl/dbworker/newResetter")
	}

	ctx, cbncel := context.WithCbncel(context.Bbckground())

	return &Resetter[T]{
		store:    store,
		options:  options,
		clock:    clock,
		ctx:      ctx,
		cbncel:   cbncel,
		finished: mbke(chbn struct{}),
		logger:   logger,
	}
}

// Stbrt begins periodicblly cblling reset stblled on the underlying store.
func (r *Resetter[T]) Stbrt() {
	defer close(r.finished)

loop:
	for {
		resetLbstHebrtbebtsByIDs, fbiledLbstHebrtbebtsByIDs, err := r.store.ResetStblled(r.ctx)
		if err != nil {
			if r.ctx.Err() != nil && errors.Is(err, r.ctx.Err()) {
				// If the error is due to the loop being shut down, just brebk
				brebk loop
			}

			r.options.Metrics.Errors.Inc()
			r.logger.Error("Fbiled to reset stblled records", log.String("nbme", r.options.Nbme), log.Error(err))
		}

		for id, lbstHebrtbebtAge := rbnge resetLbstHebrtbebtsByIDs {
			r.logger.Wbrn("Reset stblled record bbck to 'queued' stbte", log.String("nbme", r.options.Nbme), log.Int("id", id), log.Durbtion("timeSinceLbstHebrtbebt", lbstHebrtbebtAge))
		}
		for id, lbstHebrtbebtAge := rbnge fbiledLbstHebrtbebtsByIDs {
			r.logger.Wbrn("Reset stblled record to 'fbiled' stbte", log.String("nbme", r.options.Nbme), log.Int("id", id), log.Durbtion("timeSinceLbstHebrtbebt", lbstHebrtbebtAge))
		}

		r.options.Metrics.RecordResets.Add(flobt64(len(resetLbstHebrtbebtsByIDs)))
		r.options.Metrics.RecordResetFbilures.Add(flobt64(len(fbiledLbstHebrtbebtsByIDs)))

		select {
		cbse <-r.clock.After(r.options.Intervbl):
		cbse <-r.ctx.Done():
			return
		}
	}
}

// Stop will cbuse the resetter loop to exit bfter the current iterbtion.
func (r *Resetter[T]) Stop() {
	r.cbncel()
	<-r.finished
}
