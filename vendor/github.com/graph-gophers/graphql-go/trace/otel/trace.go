package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/introspection"
)

// DefaultTracer creates a tracer using a default name.
func DefaultTracer() *Tracer {
	return &Tracer{
		Tracer: otel.Tracer("graphql-go"),
	}
}

// Tracer is an OpenTelemetry implementation for graphql-go. Set the Tracer
// property to your tracer instance as required.
type Tracer struct {
	Tracer oteltrace.Tracer
}

func (t *Tracer) TraceQuery(ctx context.Context, queryString string, operationName string, variables map[string]interface{}, varTypes map[string]*introspection.Type) (context.Context, func([]*errors.QueryError)) {
	spanCtx, span := t.Tracer.Start(ctx, "GraphQL Request")

	var attributes []attribute.KeyValue
	attributes = append(attributes, attribute.String("graphql.query", queryString))
	if operationName != "" {
		attributes = append(attributes, attribute.String("graphql.operationName", operationName))
	}
	if len(variables) != 0 {
		attributes = append(attributes, attribute.String("graphql.variables", fmt.Sprintf("%v", variables)))
	}
	span.SetAttributes(attributes...)

	return spanCtx, func(errs []*errors.QueryError) {
		if len(errs) > 0 {
			msg := errs[0].Error()
			if len(errs) > 1 {
				msg += fmt.Sprintf(" (and %d more errors)", len(errs)-1)
			}

			span.SetStatus(codes.Error, msg)
		}
		span.End()
	}
}

func (t *Tracer) TraceField(ctx context.Context, label, typeName, fieldName string, trivial bool, args map[string]interface{}) (context.Context, func(*errors.QueryError)) {
	if trivial {
		return ctx, func(*errors.QueryError) {}
	}

	var attributes []attribute.KeyValue

	spanCtx, span := t.Tracer.Start(ctx, fmt.Sprintf("Field: %v", label))
	attributes = append(attributes, attribute.String("graphql.type", typeName))
	attributes = append(attributes, attribute.String("graphql.field", fieldName))
	for name, value := range args {
		attributes = append(attributes, attribute.String("graphql.args."+name, fmt.Sprintf("%v", value)))
	}
	span.SetAttributes(attributes...)

	return spanCtx, func(err *errors.QueryError) {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}
}

func (t *Tracer) TraceValidation(ctx context.Context) func([]*errors.QueryError) {
	_, span := t.Tracer.Start(ctx, "GraphQL Validate")

	return func(errs []*errors.QueryError) {
		if len(errs) > 0 {
			msg := errs[0].Error()
			if len(errs) > 1 {
				msg += fmt.Sprintf(" (and %d more errors)", len(errs)-1)
			}
			span.SetStatus(codes.Error, msg)
		}
		span.End()
	}
}
