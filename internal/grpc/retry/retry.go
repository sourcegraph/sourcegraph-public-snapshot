// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package retry

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/grpc/grpcutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"

	"google.golang.org/grpc/codes"
	grpcMetadata "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	AttemptMetadataKey       = "x-retry-attempt"
	retriedTraceAttributeKey = "x-sourcegraph-grpc-retried"

	retryTraceEventName                    = "request-retry-decision"
	requestTraceRetryAttemptsAttributeName = "x-sourcegraph-grpc-retry-attempts-number"
)

var metricRetryAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_grpc_client_retry_attempts_total",
	Help: "Total number of gRPC client retries",
}, []string{
	"grpc_service", // e.g. "gitserver.v1.GitserverService"
	"grpc_method",  // e.g. "Exec"
	"is_retried",   // e.g. "true"
})

// UnaryClientInterceptor returns a new retrying unary client interceptor.
//
// The default configuration of the interceptor is to not retry *at all*. This behaviour can be
// changed through options (e.g. WithMax) on creation of the interceptor or on call (through grpc.CallOptions).
func UnaryClientInterceptor(logger log.Logger, optFuncs ...CallOption) grpc.UnaryClientInterceptor {
	intOpts := newWithCallOptions(defaultOptions, optFuncs)
	return func(parentCtx context.Context, fullMethod string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		tr := trace.FromContext(parentCtx)
		tr.SetAttributes(attribute.Bool(retriedTraceAttributeKey, false))

		grpcOpts, retryOpts := filterCallOptions(opts)
		callOpts := newWithCallOptions(intOpts, retryOpts)

		service, method := grpcutil.SplitMethodName(fullMethod)

		observer := newRetryObserver(parentCtx, logger, service, method)
		defer func() {
			defer observer.FinishRPC()
		}()

		originalCallback := callOpts.onRetryCallback
		callOpts.onRetryCallback = func(ctx context.Context, attempt uint, err error) {
			observer.OnRetry(attempt, err)

			if originalCallback != nil {
				originalCallback(ctx, attempt, err)
			}
		}

		// short circuit for simplicity, and avoiding allocations.
		if callOpts.max == 0 {
			return invoker(parentCtx, fullMethod, req, reply, cc, grpcOpts...)
		}
		var lastErr error
		for attempt := uint(0); attempt < callOpts.max; attempt++ {
			if err := waitRetryBackoff(attempt, parentCtx, callOpts); err != nil {
				return err
			}
			if attempt > 0 {
				callOpts.onRetryCallback(parentCtx, attempt, lastErr)
			}
			callCtx, cancel := perCallContext(parentCtx, callOpts, attempt)
			defer cancel() // Clean up potential resources.
			lastErr = invoker(callCtx, fullMethod, req, reply, cc, grpcOpts...)
			// TODO(mwitkow): Maybe dial and transport errors should be retriable?
			if lastErr == nil {
				return nil
			}
			if isContextError(lastErr) {
				if parentCtx.Err() != nil {
					logTrace(parentCtx, "grpc_retry parent context error", attribute.Int("attempt", int(attempt)), attribute.String("error", parentCtx.Err().Error()))
					// its parent context deadline or cancellation.
					return lastErr
				} else if callOpts.perCallTimeout != 0 {
					// We have set a perCallTimeout in the retry middleware, which would result in a context error if
					// the deadline was exceeded, in which case, try again.
					logTrace(parentCtx, "grpc_retry context error from retry call", attribute.Int("attempt", int(attempt)))
					continue
				}
			}
			if !isRetriable(lastErr, callOpts) {
				return lastErr
			}
		}
		return lastErr
	}
}

// StreamClientInterceptor returns a new retrying stream client interceptor for server side streaming calls.
//
// The default configuration of the interceptor is to not retry *at all*. This behaviour can be
// changed through options (e.g. WithMax) on creation of the interceptor or on call (through grpc.CallOptions).
//
// Retry logic is available *only for ServerStreams*, i.e. 1:n streams, as the internal logic needs
// to buffer the messages sent by the client. If retry is enabled on any other streams (ClientStreams,
// BidiStreams), the retry interceptor will fail the call.
func StreamClientInterceptor(logger log.Logger, optFuncs ...CallOption) grpc.StreamClientInterceptor {
	intOpts := newWithCallOptions(defaultOptions, optFuncs)
	return func(parentCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, fullMethod string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		service, method := grpcutil.SplitMethodName(fullMethod)
		observer := newRetryObserver(parentCtx, logger, service, method)

		wrappedStreamer := func(parentCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, fullMethod string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			tr := trace.FromContext(parentCtx)
			tr.SetAttributes(attribute.Bool(retriedTraceAttributeKey, false))

			grpcOpts, retryOpts := filterCallOptions(opts)
			callOpts := newWithCallOptions(intOpts, retryOpts)
			// short circuit for simplicity, and avoiding allocations.

			originalCallback := callOpts.onRetryCallback
			callOpts.onRetryCallback = func(ctx context.Context, attempt uint, err error) {
				observer.OnRetry(attempt, err)

				if originalCallback != nil {
					originalCallback(ctx, attempt, err)
				}
			}

			if callOpts.max == 0 {
				return streamer(parentCtx, desc, cc, fullMethod, grpcOpts...)
			}
			if desc.ClientStreams {
				return nil, status.Error(codes.Unimplemented, "grpc_retry: cannot retry on ClientStreams, set grpc_retry.Disable()")
			}

			var lastErr error
			for attempt := uint(0); attempt < callOpts.max; attempt++ {
				if err := waitRetryBackoff(attempt, parentCtx, callOpts); err != nil {
					return nil, err
				}
				if attempt > 0 {
					callOpts.onRetryCallback(parentCtx, attempt, lastErr)
				}

				var newStreamer grpc.ClientStream
				newStreamer, lastErr = streamer(parentCtx, desc, cc, fullMethod, grpcOpts...)
				if lastErr == nil {

					retryingStreamer := &serverStreamingRetryingStream{
						ClientStream: newStreamer,
						callOpts:     callOpts,
						parentCtx:    parentCtx,

						streamerCall: func(ctx context.Context) (grpc.ClientStream, error) {
							return streamer(ctx, desc, cc, fullMethod, grpcOpts...)
						},
					}

					return newRetryingStreamerWithMetrics(retryingStreamer, observer), nil
				}
				if isContextError(lastErr) {
					if parentCtx.Err() != nil {
						logTrace(parentCtx, "grpc_retry parent context error",
							attribute.Int("attempt", int(attempt)), attribute.String("error", parentCtx.Err().Error()))
						// its the parent context deadline or cancellation.
						return nil, lastErr
					} else if callOpts.perCallTimeout != 0 {
						// We have set a perCallTimeout in the retry middleware, which would result in a context error if
						// the deadline was exceeded, in which case try again.
						logTrace(parentCtx, "grpc_retry context error from retry call", attribute.Int("attempt", int(attempt)))
						continue
					}
				}
				if !isRetriable(lastErr, callOpts) {
					return nil, lastErr
				}
			}
			return nil, lastErr
		}

		ss, err := wrappedStreamer(parentCtx, desc, cc, fullMethod, opts...)
		if err != nil {
			observer.FinishRPC() // We never established the stream, so we can finish the observation since we won't be sending any messages.
			return nil, err
		}

		return ss, nil

	}
}

func newRetryingStreamerWithMetrics(s *serverStreamingRetryingStream, observer *retryObserver) grpc.ClientStream {
	postMessageSend := func(message any, err error) {
		if err != nil {
			observer.FinishRPC() // We received an error, so we won't be sending any more messages.
		}
	}

	postMessageReceive := func(message any, err error) {
		if err != nil {
			observer.FinishRPC() // We received an error, so we won't be receiving any more messages.
		}
	}

	return grpcutil.NewCallBackClientStream(s, postMessageSend, postMessageReceive)
}

// type serverStreamingRetryingStream is the implementation of grpc.ClientStream that acts as a
// proxy to the underlying call. If the first RecvMsg() call fails, it will try to reestablish
// a new ClientStream according to the retry policy. However, if the first RecvMsg() call succeeds
// but a subsequent RecvMsg() call fails, it not automatically retry and instead return the error directly to
// the caller (as it is not possible to know if the server processed the message).
type serverStreamingRetryingStream struct {
	grpc.ClientStream
	bufferedSends []any // single message that the client can sen
	wasClosedSend bool  // indicates that CloseSend was closed
	parentCtx     context.Context
	callOpts      *options
	streamerCall  func(ctx context.Context) (grpc.ClientStream, error)

	successfullyReceivedFirstMessage bool
	mu                               sync.RWMutex
}

func (s *serverStreamingRetryingStream) setStream(clientStream grpc.ClientStream) {
	s.mu.Lock()
	s.ClientStream = clientStream
	s.mu.Unlock()
}

func (s *serverStreamingRetryingStream) getStream() grpc.ClientStream {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ClientStream
}

func (s *serverStreamingRetryingStream) SendMsg(m any) error {
	s.mu.Lock()
	s.bufferedSends = append(s.bufferedSends, m)
	s.mu.Unlock()
	return s.getStream().SendMsg(m)
}

func (s *serverStreamingRetryingStream) CloseSend() error {
	s.mu.Lock()
	s.wasClosedSend = true
	s.mu.Unlock()
	return s.getStream().CloseSend()
}

func (s *serverStreamingRetryingStream) Header() (grpcMetadata.MD, error) {
	return s.getStream().Header()
}

func (s *serverStreamingRetryingStream) Trailer() grpcMetadata.MD {
	return s.getStream().Trailer()
}

func (s *serverStreamingRetryingStream) RecvMsg(m any) error {
	attemptRetry, lastErr := s.receiveMsgAndIndicateRetry(m)
	if !attemptRetry {
		return lastErr // success or hard failure
	}
	// We start off from attempt 1, because zeroth was already made on normal SendMsg().
	for attempt := uint(1); attempt < s.callOpts.max; attempt++ {
		if err := waitRetryBackoff(attempt, s.parentCtx, s.callOpts); err != nil {
			return err
		}
		s.callOpts.onRetryCallback(s.parentCtx, attempt, lastErr)
		newStream, err := s.reestablishStreamAndResendBuffer(s.parentCtx)
		if err != nil {
			// Retry dial and transport errors of establishing stream as grpc doesn't retry.
			if isRetriable(err, s.callOpts) {
				continue
			}
			return err
		}

		s.setStream(newStream)
		attemptRetry, lastErr = s.receiveMsgAndIndicateRetry(m)

		if !attemptRetry {
			return lastErr
		}
	}
	return lastErr
}

func (s *serverStreamingRetryingStream) receiveMsgAndIndicateRetry(m any) (bool, error) {
	s.mu.RLock()
	successfullyReceivedFirstMessage := s.successfullyReceivedFirstMessage
	s.mu.RUnlock()

	err := s.getStream().RecvMsg(m)
	if err == nil || err == io.EOF {
		s.mu.Lock()
		s.successfullyReceivedFirstMessage = true
		s.mu.Unlock()

		return false, err
	} else if successfullyReceivedFirstMessage {
		// We have already received the first message, so we can't retry automatically since we don't know if
		// someone has already processed the message.
		// Instead, we return the error to the user and let them decide if they want to retry.
		return false, err
	}

	if isContextError(err) {
		if s.parentCtx.Err() != nil {
			logTrace(s.parentCtx, "grpc_retry parent context error", attribute.String("error", s.parentCtx.Err().Error()))
			return false, err
		} else if s.callOpts.perCallTimeout != 0 {
			// We have set a perCallTimeout in the retry middleware, which would result in a context error if
			// the deadline was exceeded, in which case try again.
			logTrace(s.parentCtx, "grpc_retry context error from retry call")
			return true, err
		}
	}
	return isRetriable(err, s.callOpts), err
}

func (s *serverStreamingRetryingStream) reestablishStreamAndResendBuffer(callCtx context.Context) (grpc.ClientStream, error) {
	s.mu.RLock()
	bufferedSends := s.bufferedSends
	s.mu.RUnlock()
	newStream, err := s.streamerCall(callCtx)
	if err != nil {
		logTrace(callCtx, "grpc_retry failed redialing new stream", attribute.String("error", err.Error()))
		return nil, err
	}
	for _, msg := range bufferedSends {
		if err := newStream.SendMsg(msg); err != nil {
			logTrace(callCtx, "grpc_retry failed resending message", attribute.String("error", err.Error()))
			return nil, err
		}
	}
	if err := newStream.CloseSend(); err != nil {
		logTrace(callCtx, "grpc_retry failed CloseSend on new stream", attribute.String("error", err.Error()))
		return nil, err
	}
	return newStream, nil
}

func waitRetryBackoff(attempt uint, parentCtx context.Context, callOpts *options) error {
	var waitTime time.Duration = 0
	if attempt > 0 {
		waitTime = callOpts.backoffFunc(parentCtx, attempt)
	}
	if waitTime > 0 {
		logTrace(parentCtx, "grpc_retry: backing off", attribute.Int("attempt", int(attempt)), attribute.String("duration", waitTime.String()))
		timer := time.NewTimer(waitTime)
		select {
		case <-parentCtx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return contextErrToGrpcErr(parentCtx.Err())
		case <-timer.C:
		}
	}
	return nil
}

func isRetriable(err error, callOpts *options) bool {
	errCode := status.Code(err)
	if isContextError(err) {
		// context errors are not retriable based on user settings.
		return false
	}
	for _, code := range callOpts.codes {
		if code == errCode {
			return true
		}
	}
	return false
}

func isContextError(err error) bool {
	code := status.Code(err)
	return code == codes.DeadlineExceeded || code == codes.Canceled
}

func perCallContext(parentCtx context.Context, callOpts *options, attempt uint) (context.Context, context.CancelFunc) {
	cancel := context.CancelFunc(func() {})

	ctx := parentCtx
	if callOpts.perCallTimeout != 0 {
		ctx, cancel = context.WithTimeout(ctx, callOpts.perCallTimeout)
	}
	if attempt > 0 && callOpts.includeHeader {
		mdClone := metadata.ExtractOutgoing(ctx).Clone().Set(AttemptMetadataKey, fmt.Sprintf("%d", attempt))
		ctx = mdClone.ToOutgoing(ctx)
	}
	return ctx, cancel
}

func contextErrToGrpcErr(err error) error {
	switch err {
	case context.DeadlineExceeded:
		return status.Error(codes.DeadlineExceeded, err.Error())
	case context.Canceled:
		return status.Error(codes.Canceled, err.Error())
	default:
		return status.Error(codes.Unknown, err.Error())
	}
}

func logTrace(ctx context.Context, message string, attrs ...attribute.KeyValue) {
	trace.FromContext(ctx).AddEvent(message, attrs...)
}

func newRetryObserver(ctx context.Context, logger log.Logger, serviceName, methodName string) *retryObserver {
	tracingCallback := func(attempt uint, lastErr error) {
		if attempt == 0 {
			return
		}

		tr := trace.FromContext(ctx)

		if !tr.IsRecording() {
			return
		}

		fields := []attribute.KeyValue{
			attribute.Bool(retriedTraceAttributeKey, true),
			attribute.Int("attempt", int(attempt)),
			attribute.String("grpc.service", serviceName),
			attribute.String("grpc.method", methodName),
			attribute.String("grpc.code", fmt.Sprintf("%v", status.Code(lastErr))),
		}

		if lastErr != nil {
			fields = append(fields, trace.Error(lastErr))
		}

		tr.AddEvent(retryTraceEventName, fields...)

		// Record on span itself as well for ease of querying, updates
		// will overwrite previous values.
		tr.SetAttributes(
			attribute.Bool(retriedTraceAttributeKey, true),
			attribute.Int(requestTraceRetryAttemptsAttributeName, int(attempt)),
		)

		logFields := []log.Field{
			log.Int("attempt", int(attempt)),
			log.String("grpc.service", serviceName),
			log.String("grpc.method", methodName),
			log.String("grpc.code", fmt.Sprintf("%v", status.Code(lastErr))),
		}

		if lastErr != nil {
			logFields = append(logFields, log.Error(lastErr))
		}

		logger := logger.Scoped("grpcRetryInterceptor")
		trace.Logger(ctx, logger).Debug("request",
			log.Object("retry", logFields...),
		)
	}

	return &retryObserver{
		onRetryFunc: func(attempt uint, err error) {
			tracingCallback(attempt, err)
		},

		onFinishFunc: func(hasRetried bool) {
			metricRetryAttempts.WithLabelValues(serviceName, methodName, strconv.FormatBool(hasRetried)).Inc()
		},
	}
}

// retryObserver is a simple abstraction that emits traces and Prometheus metrics whenever
// an RPC is retried. It is passed to retry middleware.
type retryObserver struct {
	hasRetried  atomic.Bool
	onRetryFunc func(attempt uint, err error)

	finishOnce   sync.Once
	onFinishFunc func(hasRetried bool)
}

func (r *retryObserver) OnRetry(attempt uint, err error) {
	r.hasRetried.Store(true)

	if r.onRetryFunc != nil {
		r.onRetryFunc(attempt, err)
	}
}

func (r *retryObserver) FinishRPC() {
	r.finishOnce.Do(func() {
		if r.onFinishFunc != nil {
			r.onFinishFunc(r.hasRetried.Load())
		}
	})
}
