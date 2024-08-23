// Copyright 2022-2023 Buf Technologies, Inc.
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
	"strings"
	"sync"
	"time"

	connect "github.com/bufbuild/connect-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

// Interceptor implements [connect.Interceptor] that adds
// OpenTelemetry metrics and tracing to connect handlers and clients.
type Interceptor struct {
	config             config
	clientInstruments  instruments
	handlerInstruments instruments
}

var _ connect.Interceptor = &Interceptor{}

// NewInterceptor constructs and returns an Interceptor which implements [connect.Interceptor]
// that adds OpenTelemetry metrics and tracing to Connect handlers and clients.
func NewInterceptor(options ...Option) *Interceptor {
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
	return &Interceptor{
		config: cfg,
	}
}

func (i *Interceptor) getAndInitInstrument(isClient bool) (*instruments, error) {
	if isClient {
		i.clientInstruments.init(i.config.meter, isClient)
		return &i.clientInstruments, i.clientInstruments.initErr
	}
	i.handlerInstruments.init(i.config.meter, isClient)
	return &i.handlerInstruments, i.handlerInstruments.initErr
}

// WrapUnary implements otel tracing and metrics for unary handlers.
func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		requestStartTime := i.config.now()
		req := &Request{
			Spec:   request.Spec(),
			Peer:   request.Peer(),
			Header: request.Header(),
		}
		if i.config.filter != nil {
			if !i.config.filter(ctx, req) {
				return next(ctx, request)
			}
		}
		attributeFilter := i.config.filterAttribute.filter
		isClient := request.Spec().IsClient
		name := strings.TrimLeft(request.Spec().Procedure, "/")
		protocol := protocolToSemConv(request.Peer().Protocol)
		attributes := attributeFilter(req, requestAttributes(req)...)
		instrumentation, err := i.getAndInitInstrument(isClient)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		carrier := propagation.HeaderCarrier(request.Header())
		spanKind := trace.SpanKindClient
		requestSpan, responseSpan := semconv.MessageTypeSent, semconv.MessageTypeReceived
		traceOpts := []trace.SpanStartOption{
			trace.WithAttributes(attributes...),
			trace.WithAttributes(headerAttributes(protocol, requestKey, req.Header, i.config.requestHeaderKeys)...),
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
		attributes = attributeFilter(req, attributes...)
		span.SetStatus(spanStatus(protocol, err))
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
		requestStartTime := i.config.now()
		conn := next(ctx, spec)
		instrumentation, err := i.getAndInitInstrument(spec.IsClient)
		if err != nil {
			return &errorStreamingClientInterceptor{
				StreamingClientConn: conn,
				err:                 connect.NewError(connect.CodeInternal, err),
			}
		}
		req := &Request{
			Spec:   conn.Spec(),
			Peer:   conn.Peer(),
			Header: conn.RequestHeader(),
		}
		if i.config.filter != nil {
			if !i.config.filter(ctx, req) {
				return conn
			}
		}
		name := strings.TrimLeft(conn.Spec().Procedure, "/")
		protocol := protocolToSemConv(conn.Peer().Protocol)
		state := newStreamingState(
			req,
			i.config.filterAttribute,
			i.config.omitTraceEvents,
			requestAttributes(req),
			instrumentation.responseSize,
			instrumentation.responsesPerRPC,
			instrumentation.requestSize,
			instrumentation.requestsPerRPC,
		)
		var span trace.Span
		var createSpanOnce sync.Once
		createSpan := func() {
			ctx, span = i.config.tracer.Start(
				ctx,
				name,
				trace.WithSpanKind(trace.SpanKindClient),
				trace.WithAttributes(state.attributes...),
				trace.WithAttributes(headerAttributes(
					protocol,
					requestKey,
					conn.RequestHeader(),
					i.config.requestHeaderKeys)...),
			)
			// inject the newly created span into the carrier
			carrier := propagation.HeaderCarrier(conn.RequestHeader())
			i.config.propagator.Inject(ctx, carrier)
		}
		return &streamingClientInterceptor{
			StreamingClientConn: conn,
			onClose: func() {
				createSpanOnce.Do(createSpan)
				// state.attributes is updated with the final error that was recorded.
				// If error is nil a "success" is recorded on the span and on the final duration
				// metric. The "rpc.<protocol>.status_code" is not defined for any other metrics for
				// streams because the error only exists when finishing the stream.
				if statusCode, ok := statusCodeAttribute(protocol, state.error); ok {
					state.addAttributes(statusCode)
				}
				span.SetAttributes(state.attributes...)
				span.SetAttributes(headerAttributes(protocol, responseKey, conn.ResponseHeader(), i.config.responseHeaderKeys)...)
				span.SetStatus(spanStatus(protocol, state.error))
				span.End()
				instrumentation.duration.Record(ctx, i.config.now().Sub(requestStartTime).Milliseconds(), metric.WithAttributes(state.attributes...))
			},
			receive: func(msg any, conn connect.StreamingClientConn) error {
				return state.receive(ctx, msg, conn)
			},
			send: func(msg any, conn connect.StreamingClientConn) error {
				createSpanOnce.Do(createSpan)
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
		instrumentation, err := i.getAndInitInstrument(isClient)
		if err != nil {
			return err
		}
		req := &Request{
			Spec:   conn.Spec(),
			Peer:   conn.Peer(),
			Header: conn.RequestHeader(),
		}
		if i.config.filter != nil {
			if !i.config.filter(ctx, req) {
				return next(ctx, conn)
			}
		}
		protocol := protocolToSemConv(req.Peer.Protocol)
		name := strings.TrimLeft(conn.Spec().Procedure, "/")
		state := newStreamingState(
			req,
			i.config.filterAttribute,
			i.config.omitTraceEvents,
			requestAttributes(req),
			instrumentation.requestSize,
			instrumentation.requestsPerRPC,
			instrumentation.responseSize,
			instrumentation.responsesPerRPC,
		)
		// extract any request headers into the context
		carrier := propagation.HeaderCarrier(conn.RequestHeader())
		traceOpts := []trace.SpanStartOption{
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(state.attributes...),
			trace.WithAttributes(headerAttributes(protocol, requestKey, req.Header, i.config.requestHeaderKeys)...),
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
		err = next(ctx, streamingHandler)
		if statusCode, ok := statusCodeAttribute(protocol, err); ok {
			state.addAttributes(statusCode)
		}
		span.SetAttributes(state.attributes...)
		span.SetAttributes(headerAttributes(protocol, responseKey, conn.ResponseHeader(), i.config.responseHeaderKeys)...)
		span.SetStatus(spanStatus(protocol, err))
		instrumentation.duration.Record(ctx, i.config.now().Sub(requestStartTime).Milliseconds(), metric.WithAttributes(state.attributes...))
		return err
	}
}

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

func spanStatus(protocol string, err error) (codes.Code, string) {
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
