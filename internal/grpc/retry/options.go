// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package retry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	// DefaultRetriableCodes is a set of well known types gRPC codes that should be retri-able.
	//
	// `ResourceExhausted` means that the user quota, e.g. per-RPC limits, have been reached.
	// `Unavailable` means that system is currently unavailable and the client should retry again.
	DefaultRetriableCodes = []codes.Code{codes.ResourceExhausted, codes.Unavailable}

	defaultOptions = &options{
		max:            0, // disabled
		perCallTimeout: 0, // disabled
		includeHeader:  true,
		codes:          DefaultRetriableCodes,
		backoffFunc:    BackoffLinearWithJitter(50*time.Millisecond /*jitter*/, 0.10),
		onRetryCallback: OnRetryCallback(func(ctx context.Context, attempt uint, err error) {
			logTrace(ctx, "grpc_retry attempting backing off due to error", attribute.Int("attempt", int(attempt)), attribute.String("error", err.Error()))
		}),
	}
)

// BackoffFunc denotes a family of functions that control the backoff duration between call retries.
//
// They are called with an identifier of the attempt, and should return a time the system client should
// hold off for. If the time returned is longer than the `context.Context.Deadline` of the request
// the deadline of the request takes precedence and the wait will be interrupted before proceeding
// with the next iteration. The context can be used to extract request scoped metadata and context values.
type BackoffFunc func(ctx context.Context, attempt uint) time.Duration

// OnRetryCallback is the type of function called when a retry occurs.
type OnRetryCallback func(ctx context.Context, attempt uint, err error)

// Disable disables the retry behaviour on this call, or this interceptor.
//
// Its semantically the same to `WithMax`
func Disable() CallOption {
	return WithMax(0)
}

// WithMax sets the maximum number of retries on this call, or this interceptor.
func WithMax(maxRetries uint) CallOption {
	return CallOption{applyFunc: func(o *options) {
		o.max = maxRetries
	}}
}

// WithBackoff sets the `BackoffFunc` used to control time between retries.
func WithBackoff(bf BackoffFunc) CallOption {
	return CallOption{applyFunc: func(o *options) {
		o.backoffFunc = bf
	}}
}

// WithOnRetryCallback sets the callback to use when a retry occurs.
//
// By default, when no callback function provided, we will just print a log to trace
func WithOnRetryCallback(fn OnRetryCallback) CallOption {
	return CallOption{applyFunc: func(o *options) {
		o.onRetryCallback = fn
	}}
}

// WithCodes sets which codes should be retried.
//
// Please *use with care*, as you may be retrying non-idempotent calls.
//
// You cannot automatically retry on Cancelled and Deadline, please use `WithPerRetryTimeout` for these.
func WithCodes(retryCodes ...codes.Code) CallOption {
	return CallOption{applyFunc: func(o *options) {
		o.codes = retryCodes
	}}
}

// WithPerRetryTimeout sets the RPC timeout per call (including initial call) on this call, or this interceptor.
//
// The context.Deadline of the call takes precedence and sets the maximum time the whole invocation
// will take, but WithPerRetryTimeout can be used to limit the RPC time per each call.
//
// For example, with context.Deadline = now + 10s, and WithPerRetryTimeout(3 * time.Seconds), each
// of the retry calls (including the initial one) will have a deadline of now + 3s.
//
// A value of 0 disables the timeout overrides completely and returns to each retry call using the
// parent `context.Deadline`.
//
// Note that when this is enabled, any DeadlineExceeded errors that are propagated up will be retried.
func WithPerRetryTimeout(timeout time.Duration) CallOption {
	return CallOption{applyFunc: func(o *options) {
		o.perCallTimeout = timeout
	}}
}

type options struct {
	max             uint
	perCallTimeout  time.Duration
	includeHeader   bool
	codes           []codes.Code
	backoffFunc     BackoffFunc
	onRetryCallback OnRetryCallback
}

// CallOption is a grpc.CallOption that is local to grpc_retry.
type CallOption struct {
	grpc.EmptyCallOption // make sure we implement private after() and before() fields so we don't panic.
	applyFunc            func(opt *options)
}

func newWithCallOptions(opt *options, callOptions []CallOption) *options {
	optCopy := &options{}
	*optCopy = *opt
	for _, f := range callOptions {
		f.applyFunc(optCopy)
	}
	return optCopy
}

func filterCallOptions(callOptions []grpc.CallOption) (grpcOptions []grpc.CallOption, retryOptions []CallOption) {
	for _, opt := range callOptions {
		if co, ok := opt.(CallOption); ok {
			retryOptions = append(retryOptions, co)
		} else {
			grpcOptions = append(grpcOptions, opt)
		}
	}
	return grpcOptions, retryOptions
}
