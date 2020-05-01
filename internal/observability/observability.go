package observability

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// ObservationArgs are the arguments to the WithObservation function.
type ObservationArgs struct {
	// The error logger instance.
	Logger logging.ErrorLogger
	// The OperationMetrics to observe.
	Metrics *metrics.OperationMetrics
	// The root tracer.
	Tracer *trace.Tracer
	// The pointer to the operation's error value.
	Err *error
	// The name of the trace.
	TraceName string
	// The prefix of the error log mesage.
	LogName string
	// Fields to log before the operation is performed.
	LogFields []log.Field
}

// FinishFn is the shape of the function returned by WithObservation and should be
// invoked within a defer directly before the observed function returns.
type FinishFn func(
	// The number of things processed.
	count float64,
	// Fields to log after the operation is performed.
	additionalLogFields ...log.Field,
)

// WithObservation prepares the necessary timers, loggers, and metrics to observe an
// operation. This returns a decorated context, which should be used in place of the input
// context in the observed operation, and ah FinishFn function. This function should
// be invoked  on defer. If your function does not process a variable number of items and the
// counting metric counts invocations, the method should be deferred as follows:
//
//     func observedFoo(ctx context.Context) (err error) {
//         ctx, finish := WithObservation(ctx, ObservationArgs{
//             Err: &err,
//             Logger: logger,
//             Metrics: metrics,
//             Tracer: tracer,
//             TraceName: "TraceName",
//             LogName: "log-name"
//         })
//         defer finish(1)
//
//         return realFoo()
//     }
//
// If the function processes a variable number of items which are known only after the
// operation completes, the method should be deferred as follows:
//
//     func observedFoo(ctx context.Context) (items []Foo err error) {
//         ctx, finish := WithObservation(ctx, ObservationArgs{
//             Err: &err,
//             Logger: logger,
//             Metrics: metrics,
//             Tracer: tracer,
//             TraceName: "TraceName",
//             LogName: "log-name"
//         })
//         defer func() {
//             finish(float64(len(items)))
//         }()
//
//         return realFoo()
//     }
//
// Both WithObservation and finish can be supplied a variable number of log fields which
// will be logged in the trace and when an error occurs.
func WithObservation(ctx context.Context, args ObservationArgs) (context.Context, FinishFn) {
	began := time.Now()

	var tr *trace.Trace
	if args.Tracer != nil {
		tr, ctx = args.Tracer.New(ctx, args.TraceName, "")
		tr.LogFields(args.LogFields...)
	}

	finish := func(count float64, additionalLogFields ...log.Field) {
		elapsed := time.Since(began).Seconds()

		logFields := append(append(append(
			make([]log.Field, 0, len(args.LogFields)+len(additionalLogFields)+1),
			args.LogFields...),
			log.Float64("count", count)),
			additionalLogFields...,
		)
		kvs := make([]interface{}, 0, len(logFields)*2)
		for _, field := range logFields {
			kvs = append(kvs, field.Key(), field.Value())
		}

		args.Metrics.Observe(elapsed, count, args.Err)
		logging.Log(args.Logger, args.LogName, args.Err, kvs...)

		if tr != nil {
			tr.LogFields(logFields...)
			tr.SetErrorPtr(args.Err)
			tr.Finish()
		}
	}

	return ctx, finish
}
