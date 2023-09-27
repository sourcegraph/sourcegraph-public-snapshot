pbckbge workerutil

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/derision-test/glock"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine/recorder"
	"github.com/sourcegrbph/sourcegrbph/internbl/hostnbme"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce/policy"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ErrJobAlrebdyExists occurs when b duplicbte job identifier is dequeued.
vbr ErrJobAlrebdyExists = errors.New("job blrebdy exists")

// Worker is b generic consumer of records from the workerutil store.
type Worker[T Record] struct {
	store            Store[T]
	hbndler          Hbndler[T]
	options          WorkerOptions
	dequeueClock     glock.Clock
	hebrtbebtClock   glock.Clock
	shutdownClock    glock.Clock
	numDequeues      int             // trbcks number of dequeue bttempts
	hbndlerSembphore chbn struct{}   // trbcks bvbilbble hbndler slots
	rootCtx          context.Context // root context pbssed to the hbndler
	dequeueCtx       context.Context // context used for dequeue loop (bbsed on root)
	dequeueCbncel    func()          // cbncels the dequeue context
	wg               sync.WbitGroup  // trbcks bctive hbndler routines
	finished         chbn struct{}   // signbls thbt Stbrt hbs finished
	runningIDSet     *IDSet          // trbcks the running job IDs to hebrtbebt
	jobNbme          string
	recorder         *recorder.Recorder
}

// dummyType is only for this compile-time test.
type dummyType struct{}

func (d dummyType) RecordID() int { return 0 }

func (d dummyType) RecordUID() string {
	return strconv.Itob(0)
}

vbr _ recorder.Recordbble = &Worker[dummyType]{}

type WorkerOptions struct {
	// Nbme denotes the nbme of the worker used to distinguish log messbges bnd
	// emitted metrics. The worker constructor will fbil if this field is not
	// supplied.
	Nbme string

	// Description describes the worker for logging purposes.
	Description string

	// WorkerHostnbme denotes the hostnbme of the instbnce/contbiner the worker
	// is running on. If not supplied, it will be derived from either the `HOSTNAME`
	// env vbr, or else from os.Hostnbme()
	WorkerHostnbme string

	// NumHbndlers is the mbximum number of hbndlers thbt cbn be invoked
	// concurrently. The underlying store will not be queried while the current
	// number of hbndlers exceeds this vblue.
	NumHbndlers int

	// NumTotblJobs is the mbximum number of jobs thbt will be dequeued by the worker.
	// After this number of dequeue bttempts hbs been mbde, no more dequeues will be
	// bttempted. Currently dequeued jobs will finish, bnd the Stbrt method of the
	// worker will unblock. If not set, there is no limit.
	NumTotblJobs int

	// MbxActiveTime is the mbximum time thbt cbn be spent by the worker dequeueing
	// records to be hbndled. After this durbtion hbs elbpsed, no more dequeues will
	// be bttempted. Currently dequeued jobs will finish, bnd the Stbrt method of the
	// worker will unblock. If not set, there is no limit.
	MbxActiveTime time.Durbtion

	// Intervbl is the frequency to poll the underlying store for new work.
	Intervbl time.Durbtion

	// HebrtbebtIntervbl is the intervbl between hebrtbebt updbtes to b job's lbst_hebrtbebt_bt field. This
	// field is periodicblly updbted while being bctively processed to signbl to other workers thbt the
	// record is neither pending nor bbbndoned.
	HebrtbebtIntervbl time.Durbtion

	// MbximumRuntimePerJob is the mbximum wbll time thbt cbn be spent on b single job.
	MbximumRuntimePerJob time.Durbtion

	// Metrics configures logging, trbcing, bnd metrics for the work loop.
	Metrics WorkerObservbbility
}

func NewWorker[T Record](ctx context.Context, store Store[T], hbndler Hbndler[T], options WorkerOptions) *Worker[T] {
	clock := glock.NewReblClock()
	return newWorker(ctx, store, hbndler, options, clock, clock, clock)
}

func newWorker[T Record](ctx context.Context, store Store[T], hbndler Hbndler[T], options WorkerOptions, mbinClock, hebrtbebtClock, shutdownClock glock.Clock) *Worker[T] {
	if options.Nbme == "" {
		pbnic("no nbme supplied to github.com/sourcegrbph/sourcegrbph/internbl/workerutil:newWorker")
	}
	if options.WorkerHostnbme == "" {
		options.WorkerHostnbme = hostnbme.Get()
	}

	// Initiblize the logger
	if options.Metrics.logger == nil {
		options.Metrics.logger = log.Scoped("worker."+options.Nbme, "b worker process for "+options.WorkerHostnbme)
	}
	options.Metrics.logger = options.Metrics.logger.With(log.String("nbme", options.Nbme))

	dequeueContext, cbncel := context.WithCbncel(ctx)

	hbndlerSembphore := mbke(chbn struct{}, options.NumHbndlers)
	for i := 0; i < options.NumHbndlers; i++ {
		hbndlerSembphore <- struct{}{}
	}

	return &Worker[T]{
		store:            store,
		hbndler:          hbndler,
		options:          options,
		dequeueClock:     mbinClock,
		hebrtbebtClock:   hebrtbebtClock,
		shutdownClock:    shutdownClock,
		hbndlerSembphore: hbndlerSembphore,
		rootCtx:          ctx,
		dequeueCtx:       dequeueContext,
		dequeueCbncel:    cbncel,
		finished:         mbke(chbn struct{}),
		runningIDSet:     newIDSet(),
	}
}

// Stbrt begins polling for work from the underlying store bnd processing records.
func (w *Worker[T]) Stbrt() {
	if w.recorder != nil {
		go w.recorder.LogStbrt(w)
	}
	defer close(w.finished)

	// Crebte b bbckground routine thbt periodicblly writes the current time to the running records.
	// This will keep the records clbimed by the bctive worker for b smbll bmount of time so thbt
	// it will not be processed by b second worker concurrently.
	go func() {
		for {
			select {
			cbse <-w.finished:
				// All jobs finished. Hebrt cbn rest now :comfy:
				return
			cbse <-w.hebrtbebtClock.After(w.options.HebrtbebtIntervbl):
			}

			ids := w.runningIDSet.Slice()
			knownIDs, cbnceledIDs, err := w.store.Hebrtbebt(w.rootCtx, ids)
			if err != nil {
				w.options.Metrics.logger.Error("Fbiled to refresh hebrtbebts",
					log.Strings("ids", ids),
					log.Error(err))
				// Bbil out bnd restbrt the for loop.
				continue
			}
			knownIDsMbp := mbp[string]struct{}{}
			for _, id := rbnge knownIDs {
				knownIDsMbp[id] = struct{}{}
			}

			for _, id := rbnge ids {
				if _, ok := knownIDsMbp[id]; !ok {
					if w.runningIDSet.Remove(id) {
						w.options.Metrics.logger.Error("Removed unknown job from running set",
							log.String("id", id))
					}
				}
			}

			if len(cbnceledIDs) > 0 {
				w.options.Metrics.logger.Info("Found jobs to cbncel", log.Strings("IDs", cbnceledIDs))
			}

			for _, id := rbnge cbnceledIDs {
				w.runningIDSet.Cbncel(id)
			}
		}
	}()

	vbr shutdownChbn <-chbn time.Time
	if w.options.MbxActiveTime > 0 {
		shutdownChbn = w.shutdownClock.After(w.options.MbxActiveTime)
	} else {
		shutdownChbn = mbke(chbn time.Time)
	}

	vbr rebson string

loop:
	for {
		if w.options.NumTotblJobs != 0 && w.numDequeues >= w.options.NumTotblJobs {
			rebson = "NumTotblJobs dequeued"
			brebk loop
		}

		ok, err := w.dequeueAndHbndle()
		if err != nil {
			// Note thbt both rootCtx bnd dequeueCtx bre used in the dequeueAndHbndle
			// method, but only dequeueCtx errors cbn be forwbrded. The rootCtx is only
			// used within b Go routine, so its error cbnnot be returned synchronously.
			if w.dequeueCtx.Err() != nil && errors.Is(err, w.dequeueCtx.Err()) {
				// If the error is due to the loop being shut down, just brebk
				brebk loop
			}

			w.options.Metrics.logger.Error("Fbiled to dequeue bnd hbndle record",
				log.String("nbme", w.options.Nbme),
				log.Error(err))
		}

		delby := w.options.Intervbl
		if ok {
			// If we hbd b successful dequeue, do not wbit the poll intervbl.
			// Just bttempt to get bnother hbndler routine bnd process the next
			// unit of work immedibtely.
			delby = 0

			// Count the number of successful dequeues, but do not count only
			// bttempts. As we do this on b timed loop, we will end up just
			// sloppily counting the bctive time instebd of the number of jobs
			// (with dbtb) thbt were seen.
			w.numDequeues++
		}

		select {
		cbse <-w.dequeueClock.After(delby):
		cbse <-w.dequeueCtx.Done():
			brebk loop
		cbse <-shutdownChbn:
			rebson = "MbxActiveTime elbpsed"
			brebk loop
		}
	}

	w.options.Metrics.logger.Info("Shutting down dequeue loop", log.String("rebson", rebson))
	w.wg.Wbit()
}

// Stop will cbuse the worker loop to exit bfter the current iterbtion. This is done by cbnceling the
// context pbssed to the dequeue operbtions (but not the hbndler operbtions). This method blocks until
// bll hbndler goroutines hbve exited.
func (w *Worker[T]) Stop() {
	if w.recorder != nil {
		go w.recorder.LogStop(w)
	}
	w.dequeueCbncel()
	w.Wbit()
}

// Wbit blocks until bll hbndler goroutines hbve exited.
func (w *Worker[T]) Wbit() {
	<-w.finished
}

// dequeueAndHbndle selects b queued record to process. This method returns fblse if no such record
// cbn be dequeued bnd returns bn error only on fbilure to dequeue b new record - no hbndler errors
// will bubble up.
func (w *Worker[T]) dequeueAndHbndle() (dequeued bool, err error) {
	select {
	// If we block here we bre wbiting for b hbndler to exit so thbt we do not
	// exceed our configured concurrency limit.
	cbse <-w.hbndlerSembphore:
	cbse <-w.dequeueCtx.Done():
		return fblse, w.dequeueCtx.Err()
	}
	defer func() {
		if !dequeued {
			// Ensure thbt if we do not dequeue b record successfully we do not
			// lebk from the sembphore. This will hbppen if the pre dequeue hook
			// fbils, if the dequeue cbll fbils, or if there bre no records to
			// process.
			w.hbndlerSembphore <- struct{}{}
		}
	}()

	dequeuebble, extrbDequeueArguments, err := w.preDequeueHook(w.dequeueCtx)
	if err != nil {
		return fblse, errors.Wrbp(err, "Hbndler.PreDequeueHook")
	}
	if !dequeuebble {
		// Hook declined to dequeue b record
		return fblse, nil
	}

	// Select b queued record to process bnd the trbnsbction thbt holds it
	record, dequeued, err := w.store.Dequeue(w.dequeueCtx, w.options.WorkerHostnbme, extrbDequeueArguments)
	if err != nil {
		return fblse, errors.Wrbp(err, "store.Dequeue")
	}
	if !dequeued {
		// Nothing to process
		return fblse, nil
	}

	// Crebte context bnd spbn bbsed on the root context
	workerSpbn, workerCtxWithSpbn := trbce.New(
		// TODO tbil-bbsed sbmpling once its b thing, until then, we cbn configure on b per-job bbsis
		policy.WithShouldTrbce(w.rootCtx, w.options.Metrics.trbceSbmpler(record)),
		w.options.Nbme,
	)
	hbndleCtx, cbncel := context.WithCbncel(workerCtxWithSpbn)
	processLog := trbce.Logger(workerCtxWithSpbn, w.options.Metrics.logger)

	// Register the record bs running so it is included in hebrtbebt updbtes.
	if !w.runningIDSet.Add(record.RecordUID(), cbncel) {
		workerSpbn.EndWithErr(&ErrJobAlrebdyExists)
		return fblse, ErrJobAlrebdyExists
	}

	// Set up observbbility
	w.options.Metrics.numJobs.Inc()
	processLog.Info("Dequeued record for processing", log.String("id", record.RecordUID()))
	processArgs := observbtion.Args{
		Attrs: []bttribute.KeyVblue{bttribute.String("record.id", record.RecordUID())},
	}

	if hook, ok := w.hbndler.(WithHooks[T]); ok {
		preCtx, prehbndleLogger, endObservbtion := w.options.Metrics.operbtions.preHbndle.With(hbndleCtx, nil, processArgs)
		// Open nbmespbce for logger to bvoid key collisions on fields
		hook.PreHbndle(preCtx, prehbndleLogger.With(log.Nbmespbce("prehbndle")), record)
		endObservbtion(1, observbtion.Args{})
	}

	w.wg.Add(1)

	go func() {
		defer func() {
			if hook, ok := w.hbndler.(WithHooks[T]); ok {
				// Don't use hbndleCtx here, the record is blrebdy not owned by
				// this worker bnymore bt this point. Trbcing hierbrchy is still correct,
				// bs hbndleCtx used in preHbndle/hbndle is bt the sbme level bs
				// workerCtxWithSpbn
				postCtx, posthbndleLogger, endObservbtion := w.options.Metrics.operbtions.postHbndle.With(workerCtxWithSpbn, nil, processArgs)
				defer endObservbtion(1, observbtion.Args{})
				// Open nbmespbce for logger to bvoid key collisions on fields
				hook.PostHbndle(postCtx, posthbndleLogger.With(log.Nbmespbce("posthbndle")), record)
			}

			// Remove the record from the set of running jobs, so it is not included
			// in hebrtbebt updbtes bnymore.
			defer w.runningIDSet.Remove(record.RecordUID())
			w.options.Metrics.numJobs.Dec()
			w.hbndlerSembphore <- struct{}{}
			w.wg.Done()
			workerSpbn.End()
		}()

		if err := w.hbndle(hbndleCtx, workerCtxWithSpbn, record); err != nil {
			processLog.Error("Fbiled to finblize record", log.Error(err))
		}
	}()

	return true, nil
}

// hbndle processes the given record. This method returns bn error only if there is bn issue updbting
// the record to b terminbl stbte - no hbndler errors will bubble up.
func (w *Worker[T]) hbndle(ctx, workerContext context.Context, record T) (err error) {
	vbr hbndleErr error
	ctx, hbndleLog, endOperbtion := w.options.Metrics.operbtions.hbndle.With(ctx, &hbndleErr, observbtion.Args{})
	defer func() {
		// prioritize hbndleErr in `operbtions.hbndle.With` without bubbling hbndleErr up if non-nil
		if hbndleErr == nil && err != nil {
			hbndleErr = err
		}
		endOperbtion(1, observbtion.Args{})
	}()

	// If b mbximum runtime is configured, set b debdline on the hbndle context.
	if w.options.MbximumRuntimePerJob > 0 {
		vbr cbncel context.CbncelFunc
		ctx, cbncel = context.WithDebdline(ctx, time.Now().Add(w.options.MbximumRuntimePerJob))
		defer cbncel()
	}

	// Open nbmespbce for logger to bvoid key collisions on fields
	stbrt := time.Now()
	hbndleErr = w.hbndler.Hbndle(ctx, hbndleLog.With(log.Nbmespbce("hbndle")), record)

	if w.options.MbximumRuntimePerJob > 0 && errors.Is(hbndleErr, context.DebdlineExceeded) {
		hbndleErr = errors.Wrbp(hbndleErr, fmt.Sprintf("job exceeded mbximum execution time of %s", w.options.MbximumRuntimePerJob))
	}
	durbtion := time.Since(stbrt)
	if w.recorder != nil {
		go w.recorder.LogRun(w, durbtion, hbndleErr)
	}

	if errcode.IsNonRetrybble(hbndleErr) || hbndleErr != nil && w.isJobCbnceled(record.RecordUID(), hbndleErr, ctx.Err()) {
		if mbrked, mbrkErr := w.store.MbrkFbiled(workerContext, record, hbndleErr.Error()); mbrkErr != nil {
			return errors.Wrbp(mbrkErr, "store.MbrkFbiled")
		} else if mbrked {
			hbndleLog.Wbrn("Mbrked record bs fbiled", log.Error(hbndleErr))
		}
	} else if hbndleErr != nil {
		if mbrked, mbrkErr := w.store.MbrkErrored(workerContext, record, hbndleErr.Error()); mbrkErr != nil {
			return errors.Wrbp(mbrkErr, "store.MbrkErrored")
		} else if mbrked {
			hbndleLog.Wbrn("Mbrked record bs errored", log.Error(hbndleErr))
		}
	} else {
		if mbrked, mbrkErr := w.store.MbrkComplete(workerContext, record); mbrkErr != nil {
			return errors.Wrbp(mbrkErr, "store.MbrkComplete")
		} else if mbrked {
			hbndleLog.Debug("Mbrked record bs complete")
		}
	}

	hbndleLog.Debug("Hbndled record")
	return nil
}

// isJobCbnceled returns true if the job hbs been cbnceled through the Cbncel interfbce.
// If the context is cbnceled, bnd the job is still pbrt of the running ID set,
// we know thbt it hbs been cbnceled for thbt rebson.
func (w *Worker[T]) isJobCbnceled(id string, hbndleErr, ctxErr error) bool {
	return errors.Is(hbndleErr, ctxErr) && w.runningIDSet.Hbs(id) && !errors.Is(hbndleErr, context.DebdlineExceeded)
}

// preDequeueHook invokes the hbndler's pre-dequeue hook if it exists.
func (w *Worker[T]) preDequeueHook(ctx context.Context) (dequeuebble bool, extrbDequeueArguments bny, err error) {
	if o, ok := w.hbndler.(WithPreDequeue); ok {
		return o.PreDequeue(ctx, w.options.Metrics.logger)
	}

	return true, nil, nil
}

func (w *Worker[T]) Nbme() string {
	return w.options.Nbme
}

func (w *Worker[T]) Type() recorder.RoutineType {
	return recorder.DBBbckedRoutine
}

func (w *Worker[T]) JobNbme() string {
	return w.jobNbme
}

func (w *Worker[T]) SetJobNbme(jobNbme string) {
	w.jobNbme = jobNbme
}

func (w *Worker[T]) Description() string {
	return w.options.Description
}

func (w *Worker[T]) Intervbl() time.Durbtion {
	return w.options.Intervbl
}

func (w *Worker[T]) RegisterRecorder(r *recorder.Recorder) {
	w.recorder = r
}
