// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ot // import "go.opentelemetry.io/contrib/propagators/ot"

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/multierr"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	// Default OT Header names.
	traceIDHeader       = "ot-tracer-traceid"
	spanIDHeader        = "ot-tracer-spanid"
	sampledHeader       = "ot-tracer-sampled"
	baggageHeaderPrefix = "ot-baggage-"

	otTraceIDPadding = "0000000000000000"

	traceID64BitsWidth = 64 / 4 // 16 hex character Trace ID.
)

var (
	empty = trace.SpanContext{}

	errInvalidSampledHeader = errors.New("invalid OT Sampled header found")
	errInvalidTraceIDHeader = errors.New("invalid OT traceID header found")
	errInvalidSpanIDHeader  = errors.New("invalid OT spanID header found")
	errInvalidScope         = errors.New("require either both traceID and spanID or none")
)

// OT propagator serializes SpanContext to/from ot-trace-* headers.
type OT struct{}

var _ propagation.TextMapPropagator = OT{}

// Inject injects a context into the carrier as OT headers.
// NOTE: In order to interop with systems that use the OT header format, trace ids MUST be 64-bits.
func (o OT) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sc := trace.SpanFromContext(ctx).SpanContext()

	if !sc.TraceID().IsValid() || !sc.SpanID().IsValid() {
		// don't bother injecting anything if either trace or span IDs are not valid
		return
	}

	carrier.Set(traceIDHeader, sc.TraceID().String()[len(sc.TraceID().String())-traceID64BitsWidth:])
	carrier.Set(spanIDHeader, sc.SpanID().String())

	if sc.IsSampled() {
		carrier.Set(sampledHeader, "true")
	} else {
		carrier.Set(sampledHeader, "false")
	}

	for _, m := range baggage.FromContext(ctx).Members() {
		carrier.Set(fmt.Sprintf("%s%s", baggageHeaderPrefix, m.Key()), m.Value())
	}
}

// Extract extracts a context from the carrier if it contains OT headers.
func (o OT) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	var (
		sc  trace.SpanContext
		err error
	)

	var (
		traceID = carrier.Get(traceIDHeader)
		spanID  = carrier.Get(spanIDHeader)
		sampled = carrier.Get(sampledHeader)
	)
	sc, err = extract(traceID, spanID, sampled)
	if err != nil || !sc.IsValid() {
		return ctx
	}

	bags, err := extractBags(carrier)
	if err != nil {
		return trace.ContextWithRemoteSpanContext(ctx, sc)
	}
	ctx = baggage.ContextWithBaggage(ctx, bags)
	return trace.ContextWithRemoteSpanContext(ctx, sc)
}

// Fields returns the OT header keys whose values are set with Inject.
func (o OT) Fields() []string {
	return []string{traceIDHeader, spanIDHeader, sampledHeader}
}

// extractBags extracts OpenTracing baggage information from carrier.
func extractBags(carrier propagation.TextMapCarrier) (baggage.Baggage, error) {
	var err error
	var members []baggage.Member
	for _, key := range carrier.Keys() {
		lowerKey := strings.ToLower(key)
		if !strings.HasPrefix(lowerKey, baggageHeaderPrefix) {
			continue
		}
		strippedKey := strings.TrimPrefix(lowerKey, baggageHeaderPrefix)
		member, e := baggage.NewMember(strippedKey, carrier.Get(key))
		if e != nil {
			err = multierr.Append(err, e)
			continue
		}
		members = append(members, member)
	}
	bags, e := baggage.New(members...)
	if err != nil {
		return bags, multierr.Append(err, e)
	}
	return bags, err
}

// extract reconstructs a SpanContext from header values based on OT
// headers.
func extract(traceID, spanID, sampled string) (trace.SpanContext, error) {
	var (
		err           error
		requiredCount int
		scc           = trace.SpanContextConfig{}
	)

	switch strings.ToLower(sampled) {
	case "0", "false":
		// Zero value for TraceFlags sample bit is unset.
	case "1", "true":
		scc.TraceFlags = trace.FlagsSampled
	case "":
		// Zero value for TraceFlags sample bit is unset.
	default:
		return empty, errInvalidSampledHeader
	}

	if traceID != "" {
		requiredCount++
		id := traceID
		if len(traceID) == 16 {
			// Pad 64-bit trace IDs.
			id = otTraceIDPadding + traceID
		}
		if scc.TraceID, err = trace.TraceIDFromHex(id); err != nil {
			return empty, errInvalidTraceIDHeader
		}
	}

	if spanID != "" {
		requiredCount++
		if scc.SpanID, err = trace.SpanIDFromHex(spanID); err != nil {
			return empty, errInvalidSpanIDHeader
		}
	}

	if requiredCount != 0 && requiredCount != 2 {
		return empty, errInvalidScope
	}

	return trace.NewSpanContext(scc), nil
}
