package messagesize

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/internal/grpc/grpcutil"
)

var (
	metricServerSingleMessageSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_grpc_server_sent_individual_message_size_bytes_per_rpc",
		Help:    "Size of individual messages sent by the server per RPC.",
		Buckets: sizeBuckets,
	}, []string{
		"grpc_service", // e.g. "gitserver.v1.GitserverService"
		"grpc_method",  // e.g. "Exec"
	})

	metricServerTotalSentPerRPCBytes = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_grpc_server_sent_bytes_per_rpc",
		Help:    "Total size of all the messages sent by the server during the course of a single RPC call",
		Buckets: sizeBuckets,
	}, []string{
		"grpc_service", // e.g. "gitserver.v1.GitserverService"
		"grpc_method",  // e.g. "Exec"
	})

	metricClientSingleMessageSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_grpc_client_sent_individual_message_size_per_rpc_bytes",
		Help:    "Size of individual messages sent by the client per RPC.",
		Buckets: sizeBuckets,
	}, []string{
		"grpc_service", // e.g. "gitserver.v1.GitserverService"
		"grpc_method",  // e.g. "Exec"
	})

	metricClientTotalSentPerRPCBytes = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_grpc_client_sent_bytes_per_rpc",
		Help:    "Total size of all the messages sent by the client during the course of a single RPC call",
		Buckets: sizeBuckets,
	}, []string{
		"grpc_service", // e.g. "gitserver.v1.GitserverService"
		"grpc_method",  // e.g. "Exec"
	})
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

var sizeBuckets = []float64{
	0,
	1 * KB,
	10 * KB,
	50 * KB,
	100 * KB,
	500 * KB,
	1 * MB,
	5 * MB,
	10 * MB,
	50 * MB,
	100 * MB,
	500 * MB,
	1 * GB,
	5 * GB,
	10 * GB,
}

// UnaryServerInterceptor is a grpc.UnaryServerInterceptor that records Prometheus metrics that observe the size of
// the response message sent back by the server for a single RPC call.
func UnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	observer := newServerMessageSizeObserver(info.FullMethod)

	return unaryServerInterceptor(observer, req, ctx, info, handler)
}

func unaryServerInterceptor(observer *messageSizeObserver, req any, ctx context.Context, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	defer observer.FinishRPC()

	r, err := handler(ctx, req)
	if err != nil {
		return r, err
	}

	response, ok := r.(proto.Message)
	if !ok {
		return r, nil
	}

	observer.Observe(response)
	return response, nil
}

// StreamServerInterceptor is a grpc.StreamServerInterceptor that records Prometheus metrics that observe both the sizes of the
// individual response messages and the cumulative response size of all the message sent back by the server over the course
// of a single RPC call.
func StreamServerInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	observer := newServerMessageSizeObserver(info.FullMethod)

	return streamServerInterceptor(observer, srv, ss, info, handler)
}

func streamServerInterceptor(observer *messageSizeObserver, srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	defer observer.FinishRPC()

	wrappedStream := newObservingServerStream(ss, observer)

	return handler(srv, wrappedStream)
}

// UnaryClientInterceptor is a grpc.UnaryClientInterceptor that records Prometheus metrics that observe the size of
// the request message sent by client for a single RPC call.
func UnaryClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	o := newClientMessageSizeObserver(method)
	return unaryClientInterceptor(o, ctx, method, req, reply, cc, invoker, opts...)
}

func unaryClientInterceptor(observer *messageSizeObserver, ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	defer observer.FinishRPC()

	err := invoker(ctx, method, req, reply, cc, opts...)
	if err != nil {
		// Don't record the size of the message if there was an error sending it, since it may not have been sent.
		return err
	}

	// Observe the size of the request message.
	request, ok := req.(proto.Message)
	if !ok {
		return nil
	}

	observer.Observe(request)
	return nil
}

// StreamClientInterceptor is a grpc.StreamClientInterceptor that records Prometheus metrics that observe both the sizes of the
// individual request messages and the cumulative request size of all the message sent by the client over the course
// of a single RPC call.
func StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	observer := newClientMessageSizeObserver(method)

	return streamClientInterceptor(observer, ctx, desc, cc, method, streamer, opts...)
}

func streamClientInterceptor(observer *messageSizeObserver, ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}

	wrappedStream := newObservingClientStream(s, observer)
	return wrappedStream, nil
}

type observingServerStream struct {
	grpc.ServerStream

	observer *messageSizeObserver
}

func newObservingServerStream(s grpc.ServerStream, observer *messageSizeObserver) grpc.ServerStream {
	return &observingServerStream{
		ServerStream: s,
		observer:     observer,
	}
}

func (s *observingServerStream) SendMsg(m any) error {
	err := s.ServerStream.SendMsg(m)
	if err != nil {
		// Don't record the size of the message if there was an error sending it, since it may not have been sent.
		//
		// However, the stream aborts on an error,
		// so we need to record the total size of the messages sent during the course of the RPC call.
		s.observer.FinishRPC()
		return err
	}

	// Observe the size of the sent message.
	message, ok := m.(proto.Message)
	if !ok {
		return nil
	}

	s.observer.Observe(message)
	return nil
}

type observingClientStream struct {
	grpc.ClientStream

	observer *messageSizeObserver
}

func newObservingClientStream(s grpc.ClientStream, observer *messageSizeObserver) grpc.ClientStream {
	return &observingClientStream{
		ClientStream: s,
		observer:     observer,
	}
}

func (s *observingClientStream) SendMsg(m any) error {
	err := s.ClientStream.SendMsg(m)
	if err != nil {
		// Don't record the size of the message if there was an error sending it, since it may not have been sent.
		//
		// However, the stream aborts on an error,
		// so we need to record the total size of the messages sent during the course of the RPC call.
		s.observer.FinishRPC()
		return err
	}

	// Observe the size of the sent message.
	message, ok := m.(proto.Message)
	if !ok {
		return nil
	}

	s.observer.Observe(message)
	return nil
}

func (s *observingClientStream) CloseSend() error {
	err := s.ClientStream.CloseSend()

	s.observer.FinishRPC()
	return err
}

func (s *observingClientStream) RecvMsg(m any) error {
	err := s.ClientStream.RecvMsg(m)
	if err != nil {
		// Record the total size of the messages sent during the course of the RPC call, even if there was an error.
		s.observer.FinishRPC()
	}

	return err
}

func newServerMessageSizeObserver(fullMethod string) *messageSizeObserver {
	serviceName, methodName := grpcutil.SplitMethodName(fullMethod)

	onSingle := func(messageSize uint64) {
		metricServerSingleMessageSize.WithLabelValues(serviceName, methodName).Observe(float64(messageSize))
	}

	onFinish := func(messageSize uint64) {
		metricServerTotalSentPerRPCBytes.WithLabelValues(serviceName, methodName).Observe(float64(messageSize))
	}

	return &messageSizeObserver{
		onSingleFunc: onSingle,
		onFinishFunc: onFinish,
	}
}

func newClientMessageSizeObserver(fullMethod string) *messageSizeObserver {
	serviceName, methodName := grpcutil.SplitMethodName(fullMethod)

	onSingle := func(messageSize uint64) {
		metricClientSingleMessageSize.WithLabelValues(serviceName, methodName).Observe(float64(messageSize))
	}

	onFinish := func(messageSize uint64) {
		metricClientTotalSentPerRPCBytes.WithLabelValues(serviceName, methodName).Observe(float64(messageSize))
	}

	return &messageSizeObserver{
		onSingleFunc: onSingle,
		onFinishFunc: onFinish,
	}
}

// messageSizeObserver is a utility that records Prometheus metrics that observe the size of each sent message and the
// cumulative size of all sent messages during the course of a single RPC call.
type messageSizeObserver struct {
	onSingleFunc func(messageSizeBytes uint64)

	finishOnce   sync.Once
	onFinishFunc func(totalSizeBytes uint64)

	totalSizeBytes atomic.Uint64
}

// Observe records the size of a single message.
func (o *messageSizeObserver) Observe(message proto.Message) {
	s := uint64(proto.Size(message))
	o.onSingleFunc(s)

	o.totalSizeBytes.Add(s)
}

// FinishRPC records the total size of all sent messages during the course of a single RPC call.
// This function should only be called once the RPC call has completed.
func (o *messageSizeObserver) FinishRPC() {
	o.finishOnce.Do(func() {
		o.onFinishFunc(o.totalSizeBytes.Load())
	})
}

var (
	_ grpc.ServerStream = &observingServerStream{}
	_ grpc.ClientStream = &observingClientStream{}
)

var (
	_ grpc.UnaryServerInterceptor  = UnaryServerInterceptor
	_ grpc.StreamServerInterceptor = StreamServerInterceptor
	_ grpc.UnaryClientInterceptor  = UnaryClientInterceptor
	_ grpc.StreamClientInterceptor = StreamClientInterceptor
)
