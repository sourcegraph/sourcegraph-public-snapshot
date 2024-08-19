// Copyright Sam Xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otelsql

import (
	"context"
	"database/sql/driver"
	"errors"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.opentelemetry.io/otel/trace"
)

func recordSpanErrorDeferred(span trace.Span, opts SpanOptions, err *error) {
	recordSpanError(span, opts, *err)
}

func recordSpanError(span trace.Span, opts SpanOptions, err error) {
	if span == nil {
		return
	}
	if opts.RecordError != nil && !opts.RecordError(err) {
		return
	}

	switch err {
	case nil:
		return
	case driver.ErrSkip:
		if !opts.DisableErrSkip {
			span.RecordError(err)
			span.SetStatus(codes.Error, "")
		}
	default:
		span.RecordError(err)
		span.SetStatus(codes.Error, "")
	}
}

func recordMetric(
	ctx context.Context,
	instruments *instruments,
	defaultAttributes []attribute.KeyValue,
	method Method,
) func(error) {
	startTime := time.Now()

	return func(err error) {
		duration := float64(time.Since(startTime).Nanoseconds()) / 1e6

		attributes := defaultAttributes
		if err != nil {
			attributes = append(attributes, queryStatusKey.String("error"))
		} else {
			attributes = append(attributes, queryStatusKey.String("ok"))
		}

		attributes = append(attributes, queryMethodKey.String(string(method)))

		instruments.latency.Record(
			ctx,
			duration,
			metric.WithAttributes(attributes...),
		)
	}
}

func createSpan(
	ctx context.Context,
	cfg config,
	method Method,
	enableDBStatement bool,
	query string,
	args []driver.NamedValue,
) (context.Context, trace.Span) {
	attrs := cfg.Attributes
	if enableDBStatement && !cfg.SpanOptions.DisableQuery {
		attrs = append(attrs, semconv.DBStatementKey.String(query))
	}
	if cfg.AttributesGetter != nil {
		attrs = append(attrs, cfg.AttributesGetter(ctx, method, query, args)...)
	}

	return cfg.Tracer.Start(ctx, cfg.SpanNameFormatter(ctx, method, query),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)
}

func filterSpan(
	ctx context.Context,
	spanOptions SpanOptions,
	method Method,
	query string,
	args []driver.NamedValue,
) bool {
	return spanOptions.SpanFilter == nil || spanOptions.SpanFilter(ctx, method, query, args)
}

// Copied from stdlib database/sql package: src/database/sql/ctxutil.go.
func namedValueToValue(named []driver.NamedValue) ([]driver.Value, error) {
	dargs := make([]driver.Value, len(named))
	for n, param := range named {
		if len(param.Name) > 0 {
			return nil, errors.New("sql: driver does not support the use of Named Parameters")
		}
		dargs[n] = param.Value
	}
	return dargs, nil
}
