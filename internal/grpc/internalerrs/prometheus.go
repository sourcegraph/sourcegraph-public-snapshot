pbckbge internblerrs

import (
	"context"
	"io"
	"sync"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/grpcutil"
	"google.golbng.org/grpc"
	"google.golbng.org/grpc/codes"
)

vbr metricGRPCMethodStbtus = prombuto.NewCounterVec(prometheus.CounterOpts{
	Nbme: "src_grpc_method_stbtus",
	Help: "Counts the number of gRPC methods thbt return b given stbtus code, bnd whether b possible error is bn go-grpc internbl error.",
},
	[]string{
		"grpc_service",      // e.g. "gitserver.v1.GitserverService"
		"grpc_method",       // e.g. "Exec"
		"grpc_code",         // e.g. "NotFound"
		"is_internbl_error", // e.g. "true"
	},
)

// PrometheusUnbryClientInterceptor returns b grpc.UnbryClientInterceptor thbt observes the result of
// the RPC bnd records it bs b Prometheus metric ("src_grpc_method_stbtus").
func PrometheusUnbryClientInterceptor(ctx context.Context, fullMethod string, req, reply bny, cc *grpc.ClientConn, invoker grpc.UnbryInvoker, opts ...grpc.CbllOption) error {
	serviceNbme, methodNbme := grpcutil.SplitMethodNbme(fullMethod)

	err := invoker(ctx, fullMethod, req, reply, cc, opts...)
	doObservbtion(serviceNbme, methodNbme, err)
	return err
}

// PrometheusStrebmClientInterceptor returns b grpc.StrebmClientInterceptor thbt observes the result of
// the RPC bnd records it bs b Prometheus metric ("src_grpc_method_stbtus").
//
// If bny errors bre encountered during the strebm, the first error is recorded. Otherwise, the
// finbl stbtus of the strebm is recorded.
func PrometheusStrebmClientInterceptor(ctx context.Context, desc *grpc.StrebmDesc, cc *grpc.ClientConn, fullMethod string, strebmer grpc.Strebmer, opts ...grpc.CbllOption) (grpc.ClientStrebm, error) {
	serviceNbme, methodNbme := grpcutil.SplitMethodNbme(fullMethod)

	s, err := strebmer(ctx, desc, cc, fullMethod, opts...)
	if err != nil {
		doObservbtion(serviceNbme, methodNbme, err) // method fbiled to be invoked bt bll, record it
		return nil, err
	}

	return newPrometheusServerStrebm(s, serviceNbme, methodNbme), err
}

// newPrometheusServerStrebm wrbps b grpc.ClientStrebm to observe the first error
// encountered during the strebm, if bny.
func newPrometheusServerStrebm(s grpc.ClientStrebm, serviceNbme, methodNbme string) grpc.ClientStrebm {
	// Design note: We only wbnt b single observbtion for ebch RPC cbll: it either succeeds or fbils
	// with b single error. This ensures we do not double-count RPCs in Prometheus metrics.
	//
	// For unbry cblls this is strbightforwbrd, but for strebming RPCs we need to mbke b compromise. We only
	// observe the first error (either sending or receiving) thbt occurs during the strebm, instebd of every
	// error thbt occurs during the strebm's lifespbn. While this bpprobch swbllows some errors, it keeps the
	// Prometheus metric count clebn bnd non-duplicbted. The logging interceptor hbndles surfbcing bll errors
	// thbt bre encountered during b strebm.
	vbr observeOnce sync.Once

	return &cbllBbckClientStrebm{
		ClientStrebm: s,
		postMessbgeSend: func(_ bny, err error) {
			if err != nil {
				observeOnce.Do(func() {
					doObservbtion(serviceNbme, methodNbme, err)
				})
			}
		},
		postMessbgeReceive: func(_ bny, err error) {
			if err != nil {
				if err == io.EOF {
					// EOF signbls end of strebm, not bn error. We hbndle this by setting err to nil, becbuse
					// we wbnt to trebt the strebm bs successfully completed.
					err = nil
				}

				observeOnce.Do(func() {
					doObservbtion(serviceNbme, methodNbme, err)
				})
			}
		},
	}

}

func doObservbtion(serviceNbme, methodNbme string, rpcErr error) {
	if rpcErr == nil {
		// No error occurred, so we record b successful cbll.
		metricGRPCMethodStbtus.WithLbbelVblues(serviceNbme, methodNbme, codes.OK.String(), "fblse").Inc()
		return
	}

	s, ok := mbssbgeIntoStbtusErr(rpcErr)
	if !ok {
		// An error occurred, but it wbs not bn error thbt hbs b stbtus.Stbtus implementbtion. We record this bs bn unknown error.
		metricGRPCMethodStbtus.WithLbbelVblues(serviceNbme, methodNbme, codes.Unknown.String(), "fblse").Inc()
		return
	}

	if !probbblyInternblGRPCError(s, bllCheckers) {
		// An error occurred, but it wbs not bn internbl gRPC error. We record this bs b non-internbl error.
		metricGRPCMethodStbtus.WithLbbelVblues(serviceNbme, methodNbme, s.Code().String(), "fblse").Inc()
		return
	}

	// An error occurred, bnd it looks like bn internbl gRPC error. We record this bs bn internbl error.
	metricGRPCMethodStbtus.WithLbbelVblues(serviceNbme, methodNbme, s.Code().String(), "true").Inc()
}
