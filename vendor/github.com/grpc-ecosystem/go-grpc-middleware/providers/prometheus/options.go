// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// FromError returns a grpc status if error code is a valid grpc status.
func FromError(err error) (s *status.Status, ok bool) {
	return status.FromError(err)

	// TODO: @yashrsharma44 - discuss if we require more error handling from the previous package
}

// A CounterOption lets you add options to Counter metrics using With* funcs.
type CounterOption func(*prometheus.CounterOpts)

type counterOptions []CounterOption

func (co counterOptions) apply(o prometheus.CounterOpts) prometheus.CounterOpts {
	for _, f := range co {
		f(&o)
	}
	return o
}

// WithConstLabels allows you to add ConstLabels to Counter metrics.
func WithConstLabels(labels prometheus.Labels) CounterOption {
	return func(o *prometheus.CounterOpts) {
		o.ConstLabels = labels
	}
}

// A HistogramOption lets you add options to Histogram metrics using With*
// funcs.
type HistogramOption func(*prometheus.HistogramOpts)

type histogramOptions []HistogramOption

func (ho histogramOptions) apply(o prometheus.HistogramOpts) prometheus.HistogramOpts {
	for _, f := range ho {
		f(&o)
	}
	return o
}

// WithHistogramBuckets allows you to specify custom bucket ranges for histograms if EnableHandlingTimeHistogram is on.
func WithHistogramBuckets(buckets []float64) HistogramOption {
	return func(o *prometheus.HistogramOpts) { o.Buckets = buckets }
}

// WithHistogramConstLabels allows you to add custom ConstLabels to
// histograms metrics.
func WithHistogramConstLabels(labels prometheus.Labels) HistogramOption {
	return func(o *prometheus.HistogramOpts) {
		o.ConstLabels = labels
	}
}

func typeFromMethodInfo(mInfo *grpc.MethodInfo) grpcType {
	if !mInfo.IsClientStream && !mInfo.IsServerStream {
		return Unary
	}
	if mInfo.IsClientStream && !mInfo.IsServerStream {
		return ClientStream
	}
	if !mInfo.IsClientStream && mInfo.IsServerStream {
		return ServerStream
	}
	return BidiStream
}

// An Option lets you add options to prometheus interceptors using With* funcs.
type Option func(*config)

type config struct {
	exemplarFn exemplarFromCtxFn
}

func (c *config) apply(opts []Option) {
	for _, o := range opts {
		o(c)
	}
}

// WithExemplarFromContext sets function that will be used to deduce exemplar for all counter and histogram metrics.
func WithExemplarFromContext(exemplarFn exemplarFromCtxFn) Option {
	return func(o *config) {
		o.exemplarFn = exemplarFn
	}
}
