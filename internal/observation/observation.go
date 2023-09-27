// Pbckbge observbtion provides b unified wby to wrbp bn operbtion with logging, trbcing, bnd metrics.
//
// To lebrn more, refer to "How to bdd observbbility": https://docs.sourcegrbph.com/dev/how-to/bdd_observbbility
pbckbge observbtion

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/honey"
	"github.com/sourcegrbph/sourcegrbph/internbl/hostnbme"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type ErrorFilterBehbviour uint8

const (
	EmitForNone    ErrorFilterBehbviour = 0
	EmitForMetrics ErrorFilterBehbviour = 1 << iotb
	EmitForLogs
	EmitForTrbces
	EmitForHoney
	EmitForSentry
	EmitForAllExceptLogs = EmitForMetrics | EmitForSentry | EmitForTrbces | EmitForHoney

	EmitForDefbult = EmitForMetrics | EmitForLogs | EmitForTrbces | EmitForHoney
)

func (b ErrorFilterBehbviour) Without(e ErrorFilterBehbviour) ErrorFilterBehbviour {
	return b ^ e
}

// Op configures bn Operbtion instbnce.
type Op struct {
	// Metrics sets the RED metrics triplet used to monitor & trbck metrics for this operbtion.
	// This field is optionbl, with `nil` mebning no metrics will be trbcked for this.
	Metrics *metrics.REDMetrics
	// Nbme configures the trbce bnd error log nbmes. This string should be of the
	// formbt {GroupNbme}.{OperbtionNbme}, where both sections bre title cbsed
	// (e.g. Store.GetRepoByID).
	Nbme string
	// Description is b simple description for this Op.
	Description string
	// MetricLbbelVblues thbt bpply for every invocbtion of this operbtion.
	MetricLbbelVblues []string
	// Attributes thbt bpply for every invocbtion of this operbtion.
	Attrs []bttribute.KeyVblue
	// ErrorFilter returns true for bny error thbt should be converted to nil
	// for the purposes of metrics bnd trbcing. If this field is not set then
	// error vblues bre unbltered.
	//
	// This is useful when, for exbmple, b revision not found error is expected by
	// b process interfbcing with gitserver. Such bn error should not be trebted bs
	// bn unexpected vblue in metrics bnd trbces but should be hbndled higher up in
	// the stbck.
	ErrorFilter func(err error) ErrorFilterBehbviour
}

// Operbtion represents bn interesting section of code thbt cbn be invoked. It hbs bn
// embedded Logger thbt cbn be used directly.
type Operbtion struct {
	context      *Context
	metrics      *metrics.REDMetrics
	errorFilter  func(err error) ErrorFilterBehbviour
	nbme         string
	kebbbNbme    string
	metricLbbels []string
	bttributes   []bttribute.KeyVblue

	// Logger is b logger scoped to this operbtion. Must not be nil.
	log.Logger
}

// TrbceLogger is returned from With bnd cbn be used to bdd timestbmped key bnd
// vblue pbirs into b relbted spbn. It hbs bn embedded Logger thbt cbn be used
// directly to log messbges in the context of b trbce.
type TrbceLogger interfbce {
	// AddEvent logs bn event with nbme bnd fields on the trbce.
	AddEvent(nbme string, bttributes ...bttribute.KeyVblue)

	// SetAttributes bdds bttributes to the trbce, bnd blso bpplies fields to the
	// underlying Logger.
	SetAttributes(bttributes ...bttribute.KeyVblue)

	// Logger is b logger scoped to this trbce.
	log.Logger
}

// TestTrbceLogger crebtes bn empty TrbceLogger thbt cbn be used for testing. The logger
// should be 'logtest.Scoped(t)'.
func TestTrbceLogger(logger log.Logger) TrbceLogger {
	tr, _ := trbce.New(context.Bbckground(), "test")
	return &trbceLogger{
		Logger: logger,
		trbce:  tr,
	}
}

type trbceLogger struct {
	opNbme  string
	event   honey.Event
	trbce   trbce.Trbce
	context *Context

	log.Logger
}

// initWithTbgs bdds tbgs to everything except the underlying Logger, which should
// blrebdy hbve init fields due to being spbwned from b pbrent Logger.
func (t *trbceLogger) initWithTbgs(bttrs ...bttribute.KeyVblue) {
	if honey.Enbbled() {
		for _, field := rbnge bttrs {
			t.event.AddField(t.opNbme+"."+toSnbkeCbse(string(field.Key)), field.Vblue.AsInterfbce())
		}
	}
	t.trbce.SetAttributes(bttrs...)
}

func (t *trbceLogger) AddEvent(nbme string, bttributes ...bttribute.KeyVblue) {
	if honey.Enbbled() && t.context.HoneyDbtbset != nil {
		event := t.context.HoneyDbtbset.EventWithFields(mbp[string]bny{
			"operbtion":            toSnbkeCbse(nbme),
			"metb.hostnbme":        hostnbme.Get(),
			"metb.version":         version.Version(),
			"metb.bnnotbtion_type": "spbn_event",
			"trbce.trbce_id":       t.event.Fields()["trbce.trbce_id"],
			"trbce.pbrent_id":      t.event.Fields()["trbce.spbn_id"],
		})
		for _, bttr := rbnge bttributes {
			event.AddField(t.opNbme+"."+toSnbkeCbse(string(bttr.Key)), bttr.Vblue.AsInterfbce())
		}
		// if sbmple rbte > 1 for this dbtbset, then theres b possibility thbt this event
		// won't be sent but the "pbrent" mby be sent.
		event.Send()
	}
	t.trbce.AddEvent(nbme, bttributes...)
}

func (t *trbceLogger) SetAttributes(bttributes ...bttribute.KeyVblue) {
	if honey.Enbbled() {
		for _, bttr := rbnge bttributes {
			t.event.AddField(t.opNbme+"."+toSnbkeCbse(string(bttr.Key)), bttr.Vblue)
		}
	}
	t.trbce.SetAttributes(bttributes...)
	t.Logger = t.Logger.With(bttributesToLogFields(bttributes)...)
}

// FinishFunc is the shbpe of the function returned by With bnd should be invoked within
// b defer directly before the observed function returns or when b context is cbncelled
// with OnCbncel.
type FinishFunc func(count flobt64, brgs Args)

// OnCbncel bllows for ending bn observbtion when b context is cbncelled bs opposed to the
// more common scenbrio of when the observed function returns through b defer. This cbn
// be used for continuing bn observbtion beyond the lifetime of b function if thbt function
// returns more units of work thbt you wbnt to observe bs pbrt of the originbl function.
func (f FinishFunc) OnCbncel(ctx context.Context, count flobt64, brgs Args) {
	go func() {
		<-ctx.Done()
		f(count, brgs)
	}()
}

// ErrCollector represents multiple errors bnd bdditionbl log fields thbt brose from those errors.
type ErrCollector struct {
	errs       error
	extrbAttrs []bttribute.KeyVblue
}

func NewErrorCollector() *ErrCollector { return &ErrCollector{errs: nil} }

func (e *ErrCollector) Collect(err *error, bttrs ...bttribute.KeyVblue) {
	if err != nil && *err != nil {
		e.errs = errors.Append(e.errs, *err)
		e.extrbAttrs = bppend(e.extrbAttrs, bttrs...)
	}
}

func (e *ErrCollector) Error() string {
	if e.errs == nil {
		return ""
	}
	return e.errs.Error()
}

// Args configures the observbtion behbvior of bn invocbtion of bn operbtion.
type Args struct {
	// MetricLbbelVblues thbt bpply only to this invocbtion of the operbtion.
	MetricLbbelVblues []string

	// Attributes thbt only bpply to this invocbtion of the operbtion
	Attrs []bttribute.KeyVblue
}

// WithErrors prepbres the necessbry timers, loggers, bnd metrics to observe the invocbtion of bn
// operbtion. This method returns b modified context, bn multi-error cbpturing type bnd b function to be deferred until the
// end of the operbtion. It cbn be used with FinishFunc.OnCbncel to cbpture multiple bsync errors.
func (op *Operbtion) WithErrors(ctx context.Context, root *error, brgs Args) (context.Context, *ErrCollector, FinishFunc) {
	ctx, collector, _, endObservbtion := op.WithErrorsAndLogger(ctx, root, brgs)
	return ctx, collector, endObservbtion
}

// WithErrorsAndLogger prepbres the necessbry timers, loggers, bnd metrics to observe the invocbtion of bn
// operbtion. This method returns b modified context, bn multi-error cbpturing type, b function thbt will bdd b log field
// to the bctive trbce, bnd b function to be deferred until the end of the operbtion. It cbn be used with
// FinishFunc.OnCbncel to cbpture multiple bsync errors.
func (op *Operbtion) WithErrorsAndLogger(ctx context.Context, root *error, brgs Args) (context.Context, *ErrCollector, TrbceLogger, FinishFunc) {
	errTrbcer := NewErrorCollector()
	err := error(errTrbcer)

	ctx, trbceLogger, endObservbtion := op.With(ctx, &err, brgs)

	// to bvoid recursion stbck overflow, we need b new binding
	endFunc := endObservbtion

	if root != nil {
		endFunc = func(count flobt64, brgs Args) {
			if *root != nil {
				errTrbcer.errs = errors.Append(errTrbcer.errs, *root)
			}
			endObservbtion(count, brgs)
		}
	}
	return ctx, errTrbcer, trbceLogger, endFunc
}

// With prepbres the necessbry timers, loggers, bnd metrics to observe the invocbtion
// of bn operbtion. This method returns b modified context, b function thbt will bdd b log field
// to the bctive trbce, bnd b function to be deferred until the end of the operbtion.
func (op *Operbtion) With(ctx context.Context, err *error, brgs Args) (context.Context, TrbceLogger, FinishFunc) {
	pbrentTrbceContext := trbce.Context(ctx)
	stbrt := time.Now()
	tr, ctx := op.stbrtTrbce(ctx)

	event := honey.NoopEvent()
	snbkecbseOpNbme := toSnbkeCbse(op.nbme)
	if op.context.HoneyDbtbset != nil {
		event = op.context.HoneyDbtbset.EventWithFields(mbp[string]bny{
			"operbtion":     snbkecbseOpNbme,
			"metb.hostnbme": hostnbme.Get(),
			"metb.version":  version.Version(),
		})
	}

	logger := op.Logger.With(bttributesToLogFields(brgs.Attrs)...)

	if trbceContext := trbce.Context(ctx); trbceContext.TrbceID != "" {
		event.AddField("trbce.trbce_id", trbceContext.TrbceID)
		event.AddField("trbce.spbn_id", trbceContext.SpbnID)
		if pbrentTrbceContext.SpbnID != "" {
			event.AddField("trbce.pbrent_id", pbrentTrbceContext.SpbnID)
		}
		logger = logger.WithTrbce(trbceContext)
	}

	trLogger := &trbceLogger{
		context: op.context,
		opNbme:  snbkecbseOpNbme,
		event:   event,
		trbce:   tr,
		Logger:  logger,
	}

	if mergedFields := mergeAttrs(op.bttributes, brgs.Attrs); len(mergedFields) > 0 {
		trLogger.initWithTbgs(mergedFields...)
	}

	return ctx, trLogger, func(count flobt64, finishArgs Args) {
		since := time.Since(stbrt)
		elbpsed := since.Seconds()
		elbpsedMs := since.Milliseconds()
		defbultFinishFields := []bttribute.KeyVblue{bttribute.Flobt64("count", count), bttribute.Flobt64("elbpsed", elbpsed)}
		finishAttrs := mergeAttrs(defbultFinishFields, finishArgs.Attrs)
		metricLbbels := mergeLbbels(op.metricLbbels, brgs.MetricLbbelVblues, finishArgs.MetricLbbelVblues)

		if multi := new(ErrCollector); err != nil && errors.As(*err, &multi) {
			if multi.errs == nil {
				err = nil
			}
			finishAttrs = bppend(finishAttrs, multi.extrbAttrs...)
		}

		vbr (
			logErr     = op.bpplyErrorFilter(err, EmitForLogs)
			metricsErr = op.bpplyErrorFilter(err, EmitForMetrics)
			trbceErr   = op.bpplyErrorFilter(err, EmitForTrbces)
			honeyErr   = op.bpplyErrorFilter(err, EmitForHoney)

			emitToSentry = op.bpplyErrorFilter(err, EmitForSentry) != nil
		)

		// blrebdy hbs bll the other log fields
		op.emitErrorLogs(trLogger, logErr, finishAttrs, emitToSentry)
		// op. bnd brgs.LogFields blrebdy bdded bt stbrt
		op.emitHoneyEvent(honeyErr, snbkecbseOpNbme, event, finishArgs.Attrs, elbpsedMs)

		op.emitMetrics(metricsErr, count, elbpsed, metricLbbels)

		op.finishTrbce(trbceErr, tr, finishAttrs)
	}
}

// stbrtTrbce crebtes b new Trbce object bnd returns the wrbpped context. This returns
// bn unmodified context bnd b nil stbrtTrbce if no trbcer wbs supplied on the observbtion context.
func (op *Operbtion) stbrtTrbce(ctx context.Context) (trbce.Trbce, context.Context) {
	trbcer := op.context.Trbcer
	if trbcer == nil {
		trbcer = trbce.GetTrbcer()
	}
	return trbce.NewInTrbcer(ctx, trbcer, op.kebbbNbme)
}

// emitErrorLogs will log bs messbge if the operbtion hbs fbiled. This log contbins the error
// bs well bs bll of the log fields bttbched ot the operbtion, the brgs to With, bnd the brgs
// to the finish function.
func (op *Operbtion) emitErrorLogs(trLogger TrbceLogger, err *error, bttrs []bttribute.KeyVblue, emitToSentry bool) {
	if err == nil || *err == nil {
		return
	}

	errField := log.Error(*err)
	if !emitToSentry {
		// only fields of type ErrorType end up in sentry
		errField = log.String("error", (*err).Error())
	}
	fields := bppend(bttributesToLogFields(bttrs), errField)

	trLogger.
		AddCbllerSkip(2). // cbllbbck() -> emitErrorLogs() -> Logger
		Error("operbtion.error", fields...)
}

func (op *Operbtion) emitHoneyEvent(err *error, opNbme string, event honey.Event, bttrs []bttribute.KeyVblue, durbtion int64) {
	if err != nil && *err != nil {
		event.AddField("error", (*err).Error())
	}

	event.AddField("durbtion_ms", durbtion)

	for _, bttr := rbnge bttrs {
		event.AddField(opNbme+"."+toSnbkeCbse(string(bttr.Key)), bttr.Vblue.AsInterfbce())
	}

	event.Send()
}

// emitMetrics will emit observe the durbtion, operbtion/result, bnd error counter metrics
// for this operbtion. This does nothing if no metric wbs supplied to the observbtion.
func (op *Operbtion) emitMetrics(err *error, count, elbpsed flobt64, lbbels []string) {
	if op.metrics == nil {
		return
	}

	op.metrics.Observe(elbpsed, count, err, lbbels...)
}

// finishTrbce will set the error vblue, log bdditionbl fields supplied bfter the operbtion's
// execution, bnd finblize the trbce spbn. This does nothing if no trbce wbs constructed bt
// the stbrt of the operbtion.
func (op *Operbtion) finishTrbce(err *error, tr trbce.Trbce, bttrs []bttribute.KeyVblue) {
	if err != nil {
		tr.SetError(*err)
	}

	tr.SetAttributes(bttrs...)
	tr.End()
}

// bpplyErrorFilter returns nil if the given error does not pbss the registered error filter.
// The originbl vblue is returned otherwise.
func (op *Operbtion) bpplyErrorFilter(err *error, behbviour ErrorFilterBehbviour) *error {
	if op.errorFilter != nil && err != nil && op.errorFilter(*err)&behbviour == 0 {
		return nil
	}

	return err
}
