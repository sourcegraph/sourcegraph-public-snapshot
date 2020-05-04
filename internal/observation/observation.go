package observation

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// Args are the arguments to the With function.
type Args struct {
	Logger  logging.ErrorLogger
	Metrics *metrics.OperationMetrics
	Tracer  *trace.Tracer
	// Err is a pointer to the operation's err result.
	Err          *error
	TraceName    string
	LogName      string
	MetricLabels []string
	// LogFields are logged prior to the operation being performed.
	LogFields []log.Field
}

// FinishFn is the shape of the function returned by With and should be
// invoked within a defer directly before the observed function returns.
type FinishFn func(
	// The number of things processed.
	count float64,
	// Fields to log after the operation is performed.
	additionalLogFields ...log.Field,
)

// With prepares the necessary timers, loggers, and metrics to observe an
// operation.
//
// If your function does not process a variable number of items and the
// counting metric counts invocations, the method should be deferred as follows:
//
//     func observedFoo(ctx context.Context) (err error) {
//         ctx, finish := observation.With(ctx, observation.Args{
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
//         ctx, finish := observation.With(ctx, observation.Args{
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
// The finish function can be supplied a variable number of log fields which will be logged
// in the trace and when an error occurs.
func With(ctx context.Context, args Args) (context.Context, FinishFn) {
	began := time.Now()

	var tr *trace.Trace
	if args.Tracer != nil {
		tr, ctx = args.Tracer.New(ctx, args.TraceName, "")
		tr.LogFields(args.LogFields...)
	}

	return ctx, func(count float64, additionalLogFields ...log.Field) {
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

		args.Metrics.Observe(elapsed, count, args.Err, args.MetricLabels...)
		logging.Log(args.Logger, args.LogName, args.Err, kvs...)

		if tr != nil {
			tr.LogFields(logFields...)
			if args.Err != nil {
				tr.SetError(*args.Err)
			}
			tr.Finish()
		}
	}
}

// Spec specifies observation.Args for a single operation.
type Spec struct {
	LogName      string
	TraceName    string
	Metrics      *metrics.OperationMetrics
	MetricLabels []string
}

// Specs encapsulates the operation args for an observed struct.
type Specs struct {
	logger  logging.ErrorLogger
	tracer  trace.Tracer
	specs   map[string]Spec
	metrics []*metrics.OperationMetrics
}

// NewSpecs creates a new Specs object that can be used with an observed struct. Before use,
// the metrics in each Spec must be registered in a Prometheus registry via MustRegister.
func NewSpecs(logger logging.ErrorLogger, tracer trace.Tracer, specs map[string]Spec) Specs {
	uniqueMetrics := map[*metrics.OperationMetrics]struct{}{}
	for _, spec := range specs {
		uniqueMetrics[spec.Metrics] = struct{}{}
	}

	var metrics []*metrics.OperationMetrics
	for k := range uniqueMetrics {
		metrics = append(metrics, k)
	}

	return Specs{
		logger:  logger,
		tracer:  tracer,
		specs:   specs,
		metrics: metrics,
	}
}

// MustRegister registers all metrics in OperationSpecs in the given
// prometheus.Registerer. It panics in case of failure.
func (s Specs) MustRegister(r prometheus.Registerer) {
	for _, om := range s.metrics {
		om.MustRegister(prometheus.DefaultRegisterer)
	}
}

// With prepares the necessary timers, loggers, and metrics to observe
// an operation. For usage details, see the With function in this package.
func (s Specs) With(
	ctx context.Context,
	err *error,
	operationName string,
	preFields ...log.Field,
) (context.Context, FinishFn) {
	spec, ok := s.specs[operationName]
	if !ok {
		return ctx, func(count float64, additionalLogFields ...log.Field) {}
	}

	return With(ctx, Args{
		Logger:       s.logger,
		Metrics:      spec.Metrics,
		Tracer:       &s.tracer,
		Err:          err,
		TraceName:    spec.TraceName,
		LogName:      spec.LogName,
		MetricLabels: spec.MetricLabels,
		LogFields:    preFields,
	})
}
