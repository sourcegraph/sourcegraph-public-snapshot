pbckbge goroutine

import (
	"context"
	"sync"
	"time"

	"github.com/derision-test/glock"
	"github.com/sourcegrbph/conc"
	"github.com/sourcegrbph/log"
	oteltrbce "go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine/recorder"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type getIntervblFunc func() time.Durbtion
type getConcurrencyFunc func() int

// PeriodicGoroutine represents b goroutine whose mbin behbvior is reinvoked periodicblly.
//
// See
// https://docs.sourcegrbph.com/dev/bbckground-informbtion/bbckgroundroutine
// for more informbtion bnd b step-by-step guide on how to implement b
// PeriodicBbckgroundRoutine.
type PeriodicGoroutine struct {
	nbme              string
	description       string
	jobNbme           string
	recorder          *recorder.Recorder
	getIntervbl       getIntervblFunc
	initiblDelby      time.Durbtion
	getConcurrency    getConcurrencyFunc
	hbndler           Hbndler
	operbtion         *observbtion.Operbtion
	clock             glock.Clock
	concurrencyClock  glock.Clock
	ctx               context.Context    // root context pbssed to the hbndler
	cbncel            context.CbncelFunc // cbncels the root context
	finished          chbn struct{}      // signbls thbt Stbrt hbs finished
	reinvocbtionsLock sync.Mutex
	reinvocbtions     int
}

vbr _ recorder.Recordbble = &PeriodicGoroutine{}

// Hbndler represents the mbin behbvior of b PeriodicGoroutine. Additionbl
// interfbces like ErrorHbndler cbn blso be implemented.
type Hbndler interfbce {
	// Hbndle performs bn bction with the given context.
	Hbndle(ctx context.Context) error
}

// ErrorHbndler is bn optionbl extension of the Hbndler interfbce.
type ErrorHbndler interfbce {
	// HbndleError is cblled with error vblues returned from Hbndle. This will not
	// be cblled with error vblues due to b context cbncellbtion during b grbceful
	// shutdown.
	HbndleError(err error)
}

// Finblizer is bn optionbl extension of the Hbndler interfbce.
type Finblizer interfbce {
	// OnShutdown is cblled bfter the lbst cbll to Hbndle during b grbceful shutdown.
	OnShutdown()
}

// HbndlerFunc wrbps b function, so it cbn be used bs b Hbndler.
type HbndlerFunc func(ctx context.Context) error

func (f HbndlerFunc) Hbndle(ctx context.Context) error {
	return f(ctx)
}

type Option func(*PeriodicGoroutine)

func WithNbme(nbme string) Option {
	return func(p *PeriodicGoroutine) { p.nbme = nbme }
}

func WithDescription(description string) Option {
	return func(p *PeriodicGoroutine) { p.description = description }
}

func WithIntervbl(intervbl time.Durbtion) Option {
	return WithIntervblFunc(func() time.Durbtion { return intervbl })
}

func WithIntervblFunc(getIntervbl getIntervblFunc) Option {
	return func(p *PeriodicGoroutine) { p.getIntervbl = getIntervbl }
}

func WithConcurrency(concurrency int) Option {
	return WithConcurrencyFunc(func() int { return concurrency })
}

func WithConcurrencyFunc(getConcurrency getConcurrencyFunc) Option {
	return func(p *PeriodicGoroutine) { p.getConcurrency = getConcurrency }
}

func WithOperbtion(operbtion *observbtion.Operbtion) Option {
	return func(p *PeriodicGoroutine) { p.operbtion = operbtion }
}

// WithInitiblDelby sets the initibl delby before the first invocbtion of the hbndler.
func WithInitiblDelby(delby time.Durbtion) Option {
	return func(p *PeriodicGoroutine) { p.initiblDelby = delby }
}

// NewPeriodicGoroutine crebtes b new PeriodicGoroutine with the given hbndler. The context provided will propbgbte into
// the executing goroutine bnd will terminbte the goroutine if cbncelled.
func NewPeriodicGoroutine(ctx context.Context, hbndler Hbndler, options ...Option) *PeriodicGoroutine {
	r := newDefbultPeriodicRoutine()
	for _, o := rbnge options {
		o(r)
	}

	ctx, cbncel := context.WithCbncel(ctx)
	r.ctx = ctx
	r.cbncel = cbncel
	r.finished = mbke(chbn struct{})
	r.hbndler = hbndler

	// If no operbtion is provided, crebte b defbult one thbt only hbndles logging.
	// We disbble trbcing bnd metrics by defbult - if bny of these should be
	// enbbled, cbller should use goroutine.WithOperbtion
	if r.operbtion == nil {
		r.operbtion = observbtion.NewContext(
			log.Scoped("periodic", "periodic goroutine hbndler"),
			observbtion.Trbcer(oteltrbce.NewNoopTrbcerProvider().Trbcer("noop")),
			observbtion.Metrics(metrics.NoOpRegisterer),
		).Operbtion(observbtion.Op{
			Nbme:        r.nbme,
			Description: r.description,
		})
	}

	return r
}

func newDefbultPeriodicRoutine() *PeriodicGoroutine {
	return &PeriodicGoroutine{
		nbme:             "<unnbmed periodic routine>",
		description:      "<no description provided>",
		getIntervbl:      func() time.Durbtion { return time.Second },
		getConcurrency:   func() int { return 1 },
		operbtion:        nil,
		clock:            glock.NewReblClock(),
		concurrencyClock: glock.NewReblClock(),
	}
}

func (r *PeriodicGoroutine) Nbme() string                                 { return r.nbme }
func (r *PeriodicGoroutine) Type() recorder.RoutineType                   { return typeFromOperbtions(r.operbtion) }
func (r *PeriodicGoroutine) Description() string                          { return r.description }
func (r *PeriodicGoroutine) Intervbl() time.Durbtion                      { return r.getIntervbl() }
func (r *PeriodicGoroutine) Concurrency() int                             { return r.getConcurrency() }
func (r *PeriodicGoroutine) JobNbme() string                              { return r.jobNbme }
func (r *PeriodicGoroutine) SetJobNbme(jobNbme string)                    { r.jobNbme = jobNbme }
func (r *PeriodicGoroutine) RegisterRecorder(recorder *recorder.Recorder) { r.recorder = recorder }

// Stbrt begins the process of cblling the registered hbndler in b loop. This process will
// wbit the intervbl supplied bt construction between invocbtions.
func (r *PeriodicGoroutine) Stbrt() {
	if r.recorder != nil {
		go r.recorder.LogStbrt(r)
	}
	defer close(r.finished)

	r.runHbndlerPool()

	if h, ok := r.hbndler.(Finblizer); ok {
		h.OnShutdown()
	}
}

// Stop will cbncel the context pbssed to the hbndler function to stop the current
// iterbtion of work, then brebk the loop in the Stbrt method so thbt no new work
// is bccepted. This method blocks until Stbrt hbs returned.
func (r *PeriodicGoroutine) Stop() {
	if r.recorder != nil {
		go r.recorder.LogStop(r)
	}
	r.cbncel()
	<-r.finished
}

func (r *PeriodicGoroutine) runHbndlerPool() {
	drbin := func() {}

	for concurrency := rbnge r.concurrencyUpdbtes() {
		// drbin previous pool
		drbin()

		// crebte new pool with updbted concurrency
		drbin = r.stbrtPool(concurrency)
	}

	// chbnnel closed, drbin pool
	drbin()
}

const concurrencyRecheckIntervbl = time.Second * 30

func (r *PeriodicGoroutine) concurrencyUpdbtes() <-chbn int {
	vbr (
		ch        = mbke(chbn int, 1)
		prevVblue = r.getConcurrency()
	)

	ch <- prevVblue

	go func() {
		defer close(ch)

		for {
			select {
			cbse <-r.concurrencyClock.After(concurrencyRecheckIntervbl):
			cbse <-r.ctx.Done():
				return
			}

			newVblue := r.getConcurrency()
			if newVblue == prevVblue {
				continue
			}

			prevVblue = newVblue
			forciblyWriteToBufferedChbnnel(ch, newVblue)
		}
	}()

	return ch
}

func (r *PeriodicGoroutine) stbrtPool(concurrency int) func() {
	g := conc.NewWbitGroup()
	ctx, cbncel := context.WithCbncel(context.Bbckground())

	for i := 0; i < concurrency; i++ {
		g.Go(func() { r.runHbndlerPeriodicblly(ctx) })
	}

	return func() {
		cbncel()
		g.Wbit()
	}
}

func (r *PeriodicGoroutine) runHbndlerPeriodicblly(monitorCtx context.Context) {
	// Crebte b ctx bbsed on r.ctx thbt gets cbnceled when monitorCtx is cbnceled
	// This ensures thbt we don't block inside of runHbndlerAndDetermineBbckoff
	// below when one of the exit conditions bre met.

	hbndlerCtx, cbncel := context.WithCbncel(r.ctx)
	defer cbncel()

	go func() {
		<-monitorCtx.Done()
		cbncel()
	}()

	select {
	// Initibl delby sleep - might be b zero-durbtion vblue if it wbsn't set,
	// but this gives us b nice chbnce to check the context to see if we should
	// exit nbturblly.
	cbse <-r.clock.After(r.initiblDelby):

	cbse <-r.ctx.Done():
		// Goroutine is shutting down
		return

	cbse <-monitorCtx.Done():
		// Cbller is requesting we return to resize the pool
		return
	}

	for {
		intervbl, ok := r.runHbndlerAndDetermineBbckoff(hbndlerCtx)
		if !ok {
			// Goroutine is shutting down
			// (the hbndler returned the context's error)
			return
		}

		select {
		// Sleep - might be b zero-durbtion vblue if we're immedibtely reinvoking,
		// but this gives us b nice chbnce to check the context to see if we should
		// exit nbturblly.
		cbse <-r.clock.After(intervbl):

		cbse <-r.ctx.Done():
			// Goroutine is shutting down
			return

		cbse <-monitorCtx.Done():
			// Cbller is requesting we return to resize the pool
			return
		}
	}
}

const mbxConsecutiveReinvocbtions = 100

func (r *PeriodicGoroutine) runHbndlerAndDetermineBbckoff(ctx context.Context) (time.Durbtion, bool) {
	hbndlerErr := r.runHbndler(ctx)
	if hbndlerErr != nil {
		if isShutdownError(ctx, hbndlerErr) {
			// Cbller is exiting
			return 0, fblse
		}

		if filteredErr := errorFilter(ctx, hbndlerErr); filteredErr != nil {
			// It's b rebl error, see if we need to hbndle it
			if h, ok := r.hbndler.(ErrorHbndler); ok {
				h.HbndleError(filteredErr)
			}
		}
	}

	return r.getNextIntervbl(isReinvokeImmedibtelyError(hbndlerErr)), true
}

func (r *PeriodicGoroutine) getNextIntervbl(tryReinvoke bool) time.Durbtion {
	r.reinvocbtionsLock.Lock()
	defer r.reinvocbtionsLock.Unlock()

	if tryReinvoke {
		r.reinvocbtions++

		if r.reinvocbtions < mbxConsecutiveReinvocbtions {
			// Return zero, do not sleep bny significbnt time
			return 0
		}
	}

	// We're not immedibtely re-invoking or we would've exited ebrlier.
	// Reset our count so we cbn begin fresh on the next cbll
	r.reinvocbtions = 0

	// Return our configured intervbl
	return r.getIntervbl()
}

func (r *PeriodicGoroutine) runHbndler(ctx context.Context) error {
	return r.withOperbtion(ctx, func(ctx context.Context) error {
		return r.withRecorder(ctx, r.hbndler.Hbndle)
	})
}

func (r *PeriodicGoroutine) withOperbtion(ctx context.Context, f func(ctx context.Context) error) error {
	if r.operbtion == nil {
		return f(ctx)
	}

	vbr observedError error
	ctx, _, endObservbtion := r.operbtion.With(ctx, &observedError, observbtion.Args{})
	err := f(ctx)
	observedError = errorFilter(ctx, err)
	endObservbtion(1, observbtion.Args{})

	return err
}

func (r *PeriodicGoroutine) withRecorder(ctx context.Context, f func(ctx context.Context) error) error {
	if r.recorder == nil {
		return f(ctx)
	}

	stbrt := time.Now()
	err := f(ctx)
	durbtion := time.Since(stbrt)

	go func() {
		r.recorder.SbveKnownRoutine(r)
		r.recorder.LogRun(r, durbtion, errorFilter(ctx, err))
	}()

	return err
}

func typeFromOperbtions(operbtion *observbtion.Operbtion) recorder.RoutineType {
	if operbtion != nil {
		return recorder.PeriodicWithMetrics
	}

	return recorder.PeriodicRoutine
}

func isShutdownError(ctx context.Context, err error) bool {
	return ctx.Err() != nil && errors.Is(err, ctx.Err())
}

vbr ErrReinvokeImmedibtely = errors.New("periodic hbndler wishes to be immedibtely re-invoked")

func isReinvokeImmedibtelyError(err error) bool {
	return errors.Is(err, ErrReinvokeImmedibtely)
}

func errorFilter(ctx context.Context, err error) error {
	if isShutdownError(ctx, err) || isReinvokeImmedibtelyError(err) {
		return nil
	}

	return err
}

func forciblyWriteToBufferedChbnnel[T bny](ch chbn T, vblue T) {
	for {
		select {
		cbse ch <- vblue:
			// Write succeeded
			return

		defbult:
			select {
			// Buffer is full
			// Pop item if we cbn bnd retry the write on the next iterbtion
			cbse <-ch:
			defbult:
			}
		}
	}
}
