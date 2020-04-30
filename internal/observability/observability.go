package observability

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type EndTraceFn func(count float64, logFields ...log.Field)

func PrepTrace(
	ctx context.Context,
	logger logging.ErrorLogger,
	metrics *metrics.OperationMetrics,
	tracer trace.Tracer,
	err *error,
	traceName string,
	logName string,
	preFields ...log.Field,
) (context.Context, EndTraceFn) {
	began := time.Now()
	tr, ctx := tracer.New(ctx, traceName, "")
	tr.LogFields(preFields...)

	endTrace := func(count float64, postFields ...log.Field) {
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

		metrics.Observe(elapsed, count, err)
		logging.Log(logger, logName, err, kvs...)
		tr.LogFields(logFields...)
		tr.SetErrorPtr(err)
		tr.Finish()
	}

	return ctx, endTrace
}
