// Copyright 2022-2023 The Connect Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otelconnect

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	connect "connectrpc.com/connect"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

// Interceptor implements [connect.Interceptor] that adds
// OpenTelemetry metrics and tracing to connect handlers and clients.
type Interceptor struct {
	config            config
	clientInstruments instruments
	serverInstruments instruments
}

var _ connect.Interceptor = &Interceptor{}

// NewInterceptor returns an interceptor that implements [connect.Interceptor].
// It adds OpenTelemetry metrics and tracing to connect handlers and clients.
// Use options to configure the interceptor. Any invalid options will cause an
// error to be returned. The interceptor will use the default tracer and meter
// providers. To use a custom tracer or meter provider pass in the
// [WithTracerProvider] or [WithMeterProvider] options. To disable metrics or
// tracing pass in the [WithoutMetrics] or [WithoutTracing] options.
func NewInterceptor(options ...Option) (*Interceptor, error) {
	cfg := config{
		now: time.Now,
		tracer: otel.GetTracerProvider().Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(semanticVersion),
		),
		propagator: otel.GetTextMapPropagator(),
		meter: otel.GetMeterProvider().Meter(
			instrumentationName,
			metric.WithInstrumentationVersion(semanticVersion)),
	}
	for _, opt := range options {
		opt.apply(&cfg)
	}
	clientInstruments, err := createInstruments(cfg.meter, clientKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create client instruments: %w", err)
	}
	serverInstruments, err := createInstruments(cfg.meter, serverKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create server instruments: %w", err)
	}
	return &Interceptor{
		config:            cfg,
		clientInstruments: clientInstruments,
		serverInstruments: serverInstruments,
	}, nil
}

// getInstruments returns the correct instrumentation for the interceptor.
func (i *Interceptor) getInstruments(isClient bool) *instruments {
	if isClient {
		return &i.clientInstruments
	}
	return &i.serverInstruments
}

// WrapUnary implements otel tracing and metrics for unary handlers.
func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		requestStartTime := i.config.now()
		if i.config.filter != nil {
			if !i.config.filter(ctx, request.Spec()) {
				return next(ctx, request)
			}
		}
		attributeFilter := i.config.filterAttribute.filter
		isClient := request.Spec().IsClient
		name := strings.TrimLeft(request.Spec().Procedure, "/")
		protocol := protocolToSemConv(request.Peer().Protocol)
		attributes := attributeFilter(request.Spec(), requestAttributes(request.Spec(), request.Peer())...)
		instrumentation := i.getInstruments(isClient)
		carrier := propagation.HeaderCarrier(request.Header())
		spanKind := trace.SpanKindClient
		requestSpan, responseSpan := semconv.MessageTypeSent, semconv.MessageTypeReceived
		traceOpts := []trace.SpanStartOption{
			trace.WithAttributes(attributes...),
			trace.WithAttributes(headerAttributes(protocol, requestKey, request.Header(), i.config.requestHeaderKeys)...),
		}
		if !isClient {
			spanKind = trace.SpanKindServer
			requestSpan, responseSpan = semconv.MessageTypeReceived, semconv.MessageTypeSent
			// if a span already exists in ctx then there must have already been another interceptor
			// that set it, so don't extract from carrier.
			if !trace.SpanContextFromContext(ctx).IsValid() {
				ctx = i.config.propagator.Extract(ctx, carrier)
				if !i.config.trustRemote {
					traceOpts = append(traceOpts,
						trace.WithNewRoot(),
						trace.WithLinks(trace.LinkFromContext(ctx)),
					)
				}
			}
		}
		traceOpts = append(traceOpts, trace.WithSpanKind(spanKind))
		ctx, span := i.config.tracer.Start(
			ctx,
			name,
			traceOpts...,
		)
		defer span.End()
		if isClient {
			i.config.propagator.Inject(ctx, carrier)
		}
		var requestSize int
		if request != nil {
			if msg, ok := request.Any().(proto.Message); ok {
				requestSize = proto.Size(msg)
			}
		}
		if !i.config.omitTraceEvents {
			span.AddEvent(messageKey,
				trace.WithAttributes(
					requestSpan,
					semconv.MessageIDKey.Int(1),
					semconv.MessageUncompressedSizeKey.Int(requestSize),
				),
			)
		}
		response, err := next(ctx, request)
		if statusCode, ok := statusCodeAttribute(protocol, err); ok {
			attributes = append(attributes, statusCode)
		}
		var responseSize int
		if err == nil {
			if msg, ok := response.Any().(proto.Message); ok {
				responseSize = proto.Size(msg)
			}
			span.SetAttributes(headerAttributes(protocol, responseKey, response.Header(), i.config.responseHeaderKeys)...)
		}
		if !i.config.omitTraceEvents {
			span.AddEvent(messageKey,
				trace.WithAttributes(
					responseSpan,
					semconv.MessageIDKey.Int(1),
					semconv.MessageUncompressedSizeKey.Int(responseSize),
				),
			)
		}
		attributes = attributeFilter(request.Spec(), attributes...)
		if isClient {
			span.SetStatus(clientSpanStatus(protocol, err))
		} else {
			span.SetStatus(serverSpanStatus(protocol, err))
		}
		span.SetAttributes(attributes...)
		instrumentation.duration.Record(ctx, i.config.now().Sub(requestStartTime).Milliseconds(), metric.WithAttributes(attributes...))
		instrumentation.requestSize.Record(ctx, int64(requestSize), metric.WithAttributes(attributes...))
		instrumentation.requestsPerRPC.Record(ctx, 1, metric.WithAttributes(attributes...))
		instrumentation.responseSize.Record(ctx, int64(responseSize), metric.WithAttributes(attributes...))
		instrumentation.responsesPerRPC.Record(ctx, 1, metric.WithAttributes(attributes...))
		return response, err
	}
}

// WrapStreamingClient implements otel tracing and metrics for streaming connect clients.
func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		if i.config.filter != nil {
			if !i.config.filter(ctx, spec) {
				return next(ctx, spec)
			}
		}
		requestStartTime := i.config.now()
		name := strings.TrimLeft(spec.Procedure, "/")
		ctx, span := i.config.tracer.Start(
			ctx,
			name,
			trace.WithSpanKind(trace.SpanKindClient),
		)
		conn := next(ctx, spec)
		instrumentation := i.getInstruments(spec.IsClient)
		// inject the newly created span into the carrier
		carrier := propagation.HeaderCarrier(conn.RequestHeader())
		i.config.propagator.Inject(ctx, carrier)
		state := newStreamingState(
			spec,
			conn.Peer(),
			i.config.filterAttribute,
			i.config.omitTraceEvents,
			instrumentation.responseSize,
			instrumentation.requestSize,
		)
		protocol := protocolToSemConv(conn.Peer().Protocol)
		var requestOnce sync.Once
		setRequestAttributes := func() {
			span.SetAttributes(
				headerAttributes(
					protocol,
					requestKey,
					conn.RequestHeader(),
					i.config.requestHeaderKeys,
				)...,
			)
		}
		closeSpan := func() {
			requestOnce.Do(setRequestAttributes)
			// state.attributes is updated with the final error that was recorded.
			// If error is nil a "success" is recorded on the span and on the final duration
			// metric. The "rpc.<protocol>.status_code" is not defined for any other metrics for
			// streams because the error only exists when finishing the stream.
			if statusCode, ok := statusCodeAttribute(protocol, state.error); ok {
				state.addAttributes(statusCode)
			}
			span.SetAttributes(state.attributes...)
			span.SetAttributes(headerAttributes(protocol, responseKey, conn.ResponseHeader(), i.config.responseHeaderKeys)...)
			span.SetStatus(clientSpanStatus(protocol, state.error))
			span.End()
			instrumentation.requestsPerRPC.Record(ctx, state.sentCounter,
				metric.WithAttributes(state.attributes...))
			instrumentation.responsesPerRPC.Record(ctx, state.receivedCounter,
				metric.WithAttributes(state.attributes...))
			duration := i.config.now().Sub(requestStartTime).Milliseconds()
			instrumentation.duration.Record(ctx, duration,
				metric.WithAttributes(state.attributes...))
		}
		stopCtxClose := afterFunc(ctx, closeSpan)
		return &streamingClientInterceptor{
			StreamingClientConn: conn,
			onClose: func() {
				if stopCtxClose() {
					closeSpan()
				}
			},
			receive: func(msg any, conn connect.StreamingClientConn) error {
				return state.receive(ctx, msg, conn)
			},
			send: func(msg any, conn connect.StreamingClientConn) error {
				requestOnce.Do(setRequestAttributes)
				return state.send(ctx, msg, conn)
			},
		}
	}
}

// WrapStreamingHandler implements otel tracing and metrics for streaming connect handlers.
func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		requestStartTime := i.config.now()
		isClient := conn.Spec().IsClient
		instrumentation := i.getInstruments(isClient)
		if i.config.filter != nil {
			if !i.config.filter(ctx, conn.Spec()) {
				return next(ctx, conn)
			}
		}
		name := strings.TrimLeft(conn.Spec().Procedure, "/")
		protocol := protocolToSemConv(conn.Peer().Protocol)
		state := newStreamingState(
			conn.Spec(),
			conn.Peer(),
			i.config.filterAttribute,
			i.config.omitTraceEvents,
			instrumentation.requestSize,
			instrumentation.responseSize,
		)
		// extract any request headers into the context
		carrier := propagation.HeaderCarrier(conn.RequestHeader())
		traceOpts := []trace.SpanStartOption{
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(state.attributes...),
			trace.WithAttributes(headerAttributes(protocol, requestKey, conn.RequestHeader(), i.config.requestHeaderKeys)...),
		}
		if !trace.SpanContextFromContext(ctx).IsValid() {
			ctx = i.config.propagator.Extract(ctx, carrier)
			if !i.config.trustRemote {
				traceOpts = append(traceOpts,
					trace.WithNewRoot(),
					trace.WithLinks(trace.LinkFromContext(ctx)),
				)
			}
		}
		// start a new span with any trace that is in the context
		ctx, span := i.config.tracer.Start(
			ctx,
			name,
			traceOpts...,
		)
		defer span.End()
		streamingHandler := &streamingHandlerInterceptor{
			StreamingHandlerConn: conn,
			receive: func(msg any, conn connect.StreamingHandlerConn) error {
				return state.receive(ctx, msg, conn)
			},
			send: func(msg any, conn connect.StreamingHandlerConn) error {
				return state.send(ctx, msg, conn)
			},
		}
		err := next(ctx, streamingHandler)
		if statusCode, ok := statusCodeAttribute(protocol, err); ok {
			state.addAttributes(statusCode)
		}
		span.SetAttributes(state.attributes...)
		span.SetAttributes(headerAttributes(protocol, responseKey, conn.ResponseHeader(), i.config.responseHeaderKeys)...)
		span.SetStatus(serverSpanStatus(protocol, err))
		instrumentation.requestsPerRPC.Record(ctx, state.receivedCounter,
			metric.WithAttributes(state.attributes...))
		instrumentation.responsesPerRPC.Record(ctx, state.sentCounter,
			metric.WithAttributes(state.attributes...))
		duration := i.config.now().Sub(requestStartTime).Milliseconds()
		instrumentation.duration.Record(ctx, duration,
			metric.WithAttributes(state.attributes...))
		return err
	}
}

// protocolToSemConv converts the protocol string to the OpenTelemetry format.
func protocolToSemConv(protocol string) string {
	switch protocol {
	case grpcwebString:
		return grpcwebProtocol
	case grpcProtocol:
		return grpcProtocol
	case connectString:
		return connectProtocol
	default:
		return protocol
	}
}

func clientSpanStatus(protocol string, err error) (codes.Code, string) {
	if err == nil {
		return codes.Unset, ""
	}
	if protocol == connectProtocol && connect.IsNotModifiedError(err) {
		return codes.Unset, ""
	}
	if connectErr := new(connect.Error); errors.As(err, &connectErr) {
		return codes.Error, connectErr.Message()
	}
	return codes.Error, err.Error()
}

func serverSpanStatus(protocol string, err error) (codes.Code, string) {
	if err == nil {
		return codes.Unset, ""
	}
	if protocol == connectProtocol && connect.IsNotModifiedError(err) {
		return codes.Unset, ""
	}

	if connectErr := new(connect.Error); errors.As(err, &connectErr) {
		switch connectErr.Code() {
		case connect.CodeUnknown,
			connect.CodeDeadlineExceeded,
			connect.CodeUnimplemented,
			connect.CodeInternal,
			connect.CodeUnavailable,
			connect.CodeDataLoss:
			return codes.Error, connectErr.Message()
		case connect.CodeCanceled,
			connect.CodeInvalidArgument,
			connect.CodeNotFound,
			connect.CodeAlreadyExists,
			connect.CodePermissionDenied,
			connect.CodeResourceExhausted,
			connect.CodeFailedPrecondition,
			connect.CodeAborted,
			connect.CodeOutOfRange,
			connect.CodeUnauthenticated:
			return codes.Unset, ""
		default:
			return codes.Unset, ""
		}
	}

	return codes.Error, err.Error()
}
