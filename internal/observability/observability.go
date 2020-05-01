package observability

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// FinishFn is the shape of the function returned by WithObservation and should be
// invoked within a defer directly before the observed function returns.
type FinishFn func(
	// The number of things processed.
	count float64,
	// Fields to log after the operation is performed.
	logFields ...log.Field,
)

// WithObservation prepares the necessary timers, loggers, and metrics to observe an
// operation. This returns a decorated context, which should be used in place of the input
// context in the observed operation, and ah FinishFn function. This function should
// be invoked  on defer. If your function does not process a variable number of items and the
// counting metric counts invocations, the method should be deferred as follows:
//
//     func observedFoo(ctx context.Context) (err error) {
//         ctx, finish := WithObservation(ctx, &err, logger, metrics, tracer, "TraceName", "log-name")
//         defer finish(1)
//
//         return realFoo()
//     }
//
// If the function processes a variable number of items which are known only after the
// operation completes, the method should be deferred as follows:
//
//     func observedFoo(ctx context.Context) (items []Foo err error) {
//         ctx, finish := WithObservation(ctx, &err, logger, metrics, tracer, "TraceName", "log-name")
//         defer func() {
//             finish(float64(len(items)))
//         }()
//
//         return realFoo()
//     }
//
// Both WithObservation and finish can be supplied a variable number of log fields which
// will be logged in the trace and when an error occurs.
func WithObservation(
	// The input context.
	ctx context.Context,
	// The error logger instance.
	logger logging.ErrorLogger,
	// The OperationMetrics to observe.
	metrics *metrics.OperationMetrics,
	// The root tracer.
	tracer trace.Tracer,
	// The pointer to the operation's error value.
	err *error,
	// The name of the trace.
	traceName string,
	// The prefix of the error log mesage.
	logName string,
	// Fields to log before the operation is performed.
	preFields ...log.Field,
) (context.Context, FinishFn) {
	began := time.Now()
	tr, ctx := tracer.New(ctx, traceName, "")
	tr.LogFields(preFields...)

	finish := func(count float64, postFields ...log.Field) {
		elapsed := time.Since(began).Seconds()

		logFields := append(append(append(
			make([]log.Field, 0, len(preFields)+len(postFields)+1),
			preFields...),
			log.Float64("count", count)),
			postFields...,
		)

		kvs := make([]interface{}, 0, len(logFields)*2)
		for _, field := range logFields {
			kvs = append(kvs, field.Key(), field.Value())
		}

		if metrics != nil {
			metrics.Observe(elapsed, count, err)
		}

		logging.Log(logger, logName, err, kvs...)
		tr.LogFields(logFields...)
		tr.SetErrorPtr(err)
		tr.Finish()
	}

	return ctx, finish
}
