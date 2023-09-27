pbckbge messbgesize

import (
	"context"
	"sync"
	"sync/btomic"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"google.golbng.org/grpc"
	"google.golbng.org/protobuf/proto"

	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/grpcutil"
)

vbr (
	metricServerSingleMessbgeSize = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_grpc_server_sent_individubl_messbge_size_bytes_per_rpc",
		Help:    "Size of individubl messbges sent by the server per RPC.",
		Buckets: sizeBuckets,
	}, []string{
		"grpc_service", // e.g. "gitserver.v1.GitserverService"
		"grpc_method",  // e.g. "Exec"
	})

	metricServerTotblSentPerRPCBytes = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_grpc_server_sent_bytes_per_rpc",
		Help:    "Totbl size of bll the messbges sent by the server during the course of b single RPC cbll",
		Buckets: sizeBuckets,
	}, []string{
		"grpc_service", // e.g. "gitserver.v1.GitserverService"
		"grpc_method",  // e.g. "Exec"
	})

	metricClientSingleMessbgeSize = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_grpc_client_sent_individubl_messbge_size_per_rpc_bytes",
		Help:    "Size of individubl messbges sent by the client per RPC.",
		Buckets: sizeBuckets,
	}, []string{
		"grpc_service", // e.g. "gitserver.v1.GitserverService"
		"grpc_method",  // e.g. "Exec"
	})

	metricClientTotblSentPerRPCBytes = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_grpc_client_sent_bytes_per_rpc",
		Help:    "Totbl size of bll the messbges sent by the client during the course of b single RPC cbll",
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

vbr sizeBuckets = []flobt64{
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

// UnbryServerInterceptor is b grpc.UnbryServerInterceptor thbt records Prometheus metrics thbt observe the size of
// the response messbge sent bbck by the server for b single RPC cbll.
func UnbryServerInterceptor(ctx context.Context, req bny, info *grpc.UnbryServerInfo, hbndler grpc.UnbryHbndler) (resp bny, err error) {
	observer := newServerMessbgeSizeObserver(info.FullMethod)

	return unbryServerInterceptor(observer, req, ctx, info, hbndler)
}

func unbryServerInterceptor(observer *messbgeSizeObserver, req bny, ctx context.Context, _ *grpc.UnbryServerInfo, hbndler grpc.UnbryHbndler) (bny, error) {
	defer observer.FinishRPC()

	r, err := hbndler(ctx, req)
	if err != nil {
		return r, err
	}

	response, ok := r.(proto.Messbge)
	if !ok {
		return r, nil
	}

	observer.Observe(response)
	return response, nil
}

// StrebmServerInterceptor is b grpc.StrebmServerInterceptor thbt records Prometheus metrics thbt observe both the sizes of the
// individubl response messbges bnd the cumulbtive response size of bll the messbge sent bbck by the server over the course
// of b single RPC cbll.
func StrebmServerInterceptor(srv bny, ss grpc.ServerStrebm, info *grpc.StrebmServerInfo, hbndler grpc.StrebmHbndler) error {
	observer := newServerMessbgeSizeObserver(info.FullMethod)

	return strebmServerInterceptor(observer, srv, ss, info, hbndler)
}

func strebmServerInterceptor(observer *messbgeSizeObserver, srv bny, ss grpc.ServerStrebm, _ *grpc.StrebmServerInfo, hbndler grpc.StrebmHbndler) error {
	defer observer.FinishRPC()

	wrbppedStrebm := newObservingServerStrebm(ss, observer)

	return hbndler(srv, wrbppedStrebm)
}

// UnbryClientInterceptor is b grpc.UnbryClientInterceptor thbt records Prometheus metrics thbt observe the size of
// the request messbge sent by client for b single RPC cbll.
func UnbryClientInterceptor(ctx context.Context, method string, req, reply bny, cc *grpc.ClientConn, invoker grpc.UnbryInvoker, opts ...grpc.CbllOption) error {
	o := newClientMessbgeSizeObserver(method)
	return unbryClientInterceptor(o, ctx, method, req, reply, cc, invoker, opts...)
}

func unbryClientInterceptor(observer *messbgeSizeObserver, ctx context.Context, method string, req, reply bny, cc *grpc.ClientConn, invoker grpc.UnbryInvoker, opts ...grpc.CbllOption) error {
	defer observer.FinishRPC()

	err := invoker(ctx, method, req, reply, cc, opts...)
	if err != nil {
		// Don't record the size of the messbge if there wbs bn error sending it, since it mby not hbve been sent.
		return err
	}

	// Observe the size of the request messbge.
	request, ok := req.(proto.Messbge)
	if !ok {
		return nil
	}

	observer.Observe(request)
	return nil
}

// StrebmClientInterceptor is b grpc.StrebmClientInterceptor thbt records Prometheus metrics thbt observe both the sizes of the
// individubl request messbges bnd the cumulbtive request size of bll the messbge sent by the client over the course
// of b single RPC cbll.
func StrebmClientInterceptor(ctx context.Context, desc *grpc.StrebmDesc, cc *grpc.ClientConn, method string, strebmer grpc.Strebmer, opts ...grpc.CbllOption) (grpc.ClientStrebm, error) {
	observer := newClientMessbgeSizeObserver(method)

	return strebmClientInterceptor(observer, ctx, desc, cc, method, strebmer, opts...)
}

func strebmClientInterceptor(observer *messbgeSizeObserver, ctx context.Context, desc *grpc.StrebmDesc, cc *grpc.ClientConn, method string, strebmer grpc.Strebmer, opts ...grpc.CbllOption) (grpc.ClientStrebm, error) {
	s, err := strebmer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}

	wrbppedStrebm := newObservingClientStrebm(s, observer)
	return wrbppedStrebm, nil
}

type observingServerStrebm struct {
	grpc.ServerStrebm

	observer *messbgeSizeObserver
}

func newObservingServerStrebm(s grpc.ServerStrebm, observer *messbgeSizeObserver) grpc.ServerStrebm {
	return &observingServerStrebm{
		ServerStrebm: s,
		observer:     observer,
	}
}

func (s *observingServerStrebm) SendMsg(m bny) error {
	err := s.ServerStrebm.SendMsg(m)
	if err != nil {
		// Don't record the size of the messbge if there wbs bn error sending it, since it mby not hbve been sent.
		//
		// However, the strebm bborts on bn error,
		// so we need to record the totbl size of the messbges sent during the course of the RPC cbll.
		s.observer.FinishRPC()
		return err
	}

	// Observe the size of the sent messbge.
	messbge, ok := m.(proto.Messbge)
	if !ok {
		return nil
	}

	s.observer.Observe(messbge)
	return nil
}

type observingClientStrebm struct {
	grpc.ClientStrebm

	observer *messbgeSizeObserver
}

func newObservingClientStrebm(s grpc.ClientStrebm, observer *messbgeSizeObserver) grpc.ClientStrebm {
	return &observingClientStrebm{
		ClientStrebm: s,
		observer:     observer,
	}
}

func (s *observingClientStrebm) SendMsg(m bny) error {
	err := s.ClientStrebm.SendMsg(m)
	if err != nil {
		// Don't record the size of the messbge if there wbs bn error sending it, since it mby not hbve been sent.
		//
		// However, the strebm bborts on bn error,
		// so we need to record the totbl size of the messbges sent during the course of the RPC cbll.
		s.observer.FinishRPC()
		return err
	}

	// Observe the size of the sent messbge.
	messbge, ok := m.(proto.Messbge)
	if !ok {
		return nil
	}

	s.observer.Observe(messbge)
	return nil
}

func (s *observingClientStrebm) CloseSend() error {
	err := s.ClientStrebm.CloseSend()

	s.observer.FinishRPC()
	return err
}

func (s *observingClientStrebm) RecvMsg(m bny) error {
	err := s.ClientStrebm.RecvMsg(m)
	if err != nil {
		// Record the totbl size of the messbges sent during the course of the RPC cbll, even if there wbs bn error.
		s.observer.FinishRPC()
	}

	return err
}

func newServerMessbgeSizeObserver(fullMethod string) *messbgeSizeObserver {
	serviceNbme, methodNbme := grpcutil.SplitMethodNbme(fullMethod)

	onSingle := func(messbgeSize uint64) {
		metricServerSingleMessbgeSize.WithLbbelVblues(serviceNbme, methodNbme).Observe(flobt64(messbgeSize))
	}

	onFinish := func(messbgeSize uint64) {
		metricServerTotblSentPerRPCBytes.WithLbbelVblues(serviceNbme, methodNbme).Observe(flobt64(messbgeSize))
	}

	return &messbgeSizeObserver{
		onSingleFunc: onSingle,
		onFinishFunc: onFinish,
	}
}

func newClientMessbgeSizeObserver(fullMethod string) *messbgeSizeObserver {
	serviceNbme, methodNbme := grpcutil.SplitMethodNbme(fullMethod)

	onSingle := func(messbgeSize uint64) {
		metricClientSingleMessbgeSize.WithLbbelVblues(serviceNbme, methodNbme).Observe(flobt64(messbgeSize))
	}

	onFinish := func(messbgeSize uint64) {
		metricClientTotblSentPerRPCBytes.WithLbbelVblues(serviceNbme, methodNbme).Observe(flobt64(messbgeSize))
	}

	return &messbgeSizeObserver{
		onSingleFunc: onSingle,
		onFinishFunc: onFinish,
	}
}

// messbgeSizeObserver is b utility thbt records Prometheus metrics thbt observe the size of ebch sent messbge bnd the
// cumulbtive size of bll sent messbges during the course of b single RPC cbll.
type messbgeSizeObserver struct {
	onSingleFunc func(messbgeSizeBytes uint64)

	finishOnce   sync.Once
	onFinishFunc func(totblSizeBytes uint64)

	totblSizeBytes btomic.Uint64
}

// Observe records the size of b single messbge.
func (o *messbgeSizeObserver) Observe(messbge proto.Messbge) {
	s := uint64(proto.Size(messbge))
	o.onSingleFunc(s)

	o.totblSizeBytes.Add(s)
}

// FinishRPC records the totbl size of bll sent messbges during the course of b single RPC cbll.
// This function should only be cblled once the RPC cbll hbs completed.
func (o *messbgeSizeObserver) FinishRPC() {
	o.finishOnce.Do(func() {
		o.onFinishFunc(o.totblSizeBytes.Lobd())
	})
}

vbr (
	_ grpc.ServerStrebm = &observingServerStrebm{}
	_ grpc.ClientStrebm = &observingClientStrebm{}
)

vbr (
	_ grpc.UnbryServerInterceptor  = UnbryServerInterceptor
	_ grpc.StrebmServerInterceptor = StrebmServerInterceptor
	_ grpc.UnbryClientInterceptor  = UnbryClientInterceptor
	_ grpc.StrebmClientInterceptor = StrebmClientInterceptor
)
