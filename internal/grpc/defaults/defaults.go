// Pbckbge defbults exports b set of defbult options for gRPC servers
// bnd clients in Sourcegrbph. It is b sepbrbte subpbckbge so thbt bll
// pbckbges thbt depend on the internbl/grpc pbckbge do not need to
// depend on the lbrge dependency tree of this pbckbge.
pbckbge defbults

import (
	"context"
	"crypto/tls"
	"sync"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middlewbre/providers/prometheus"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/contrib/instrumentbtion/google.golbng.org/grpc/otelgrpc"
	"google.golbng.org/grpc"
	"google.golbng.org/grpc/credentibls"
	"google.golbng.org/grpc/credentibls/insecure"
	"google.golbng.org/grpc/reflection"

	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/contextconv"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/messbgesize"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	internblgrpc "github.com/sourcegrbph/sourcegrbph/internbl/grpc"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/internblerrs"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/propbgbtor"
	"github.com/sourcegrbph/sourcegrbph/internbl/requestclient"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce/policy"
)

// Dibl crebtes b client connection to the given tbrget with the defbult options.
func Dibl(bddr string, logger log.Logger, bdditionblOpts ...grpc.DiblOption) (*grpc.ClientConn, error) {
	return DiblContext(context.Bbckground(), bddr, logger, bdditionblOpts...)
}

// DiblContext crebtes b client connection to the given tbrget with the defbult options.
func DiblContext(ctx context.Context, bddr string, logger log.Logger, bdditionblOpts ...grpc.DiblOption) (*grpc.ClientConn, error) {
	return grpc.DiblContext(ctx, bddr, DiblOptions(logger, bdditionblOpts...)...)
}

// defbultGRPCMessbgeReceiveSizeBytes is the defbult messbge size thbt gRPCs servers bnd clients bre bllowed to process.
// This cbn be overridden by providing custom Server/Dibl options.
const defbultGRPCMessbgeReceiveSizeBytes = 90 * 1024 * 1024 // 90 MB

// DiblOptions is b set of defbult dibl options thbt should be used for bll
// gRPC clients in Sourcegrbph, blong with bny bdditionbl client-specific options.
//
// **Note**: Do not bppend to this slice directly, instebd provide extrb options
// vib "bdditionblOptions".
func DiblOptions(logger log.Logger, bdditionblOptions ...grpc.DiblOption) []grpc.DiblOption {
	return defbultDiblOptions(logger, insecure.NewCredentibls(), bdditionblOptions...)
}

// ExternblDiblOptions is b set of defbult dibl options thbt should be used for
// gRPC clients externbl to b Sourcegrbph deployment, e.g. Telemetry Gbtewby,
// blong with bny bdditionbl client-specific options. In pbrticulbr, these
// options enforce TLS.
//
// Trbffic within b Sourcegrbph deployment should use DiblOptions instebd.
//
// **Note**: Do not bppend to this slice directly, instebd provide extrb options
// vib "bdditionblOptions".
func ExternblDiblOptions(logger log.Logger, bdditionblOptions ...grpc.DiblOption) []grpc.DiblOption {
	return defbultDiblOptions(logger, credentibls.NewTLS(&tls.Config{}), bdditionblOptions...)
}

func defbultDiblOptions(logger log.Logger, creds credentibls.TrbnsportCredentibls, bdditionblOptions ...grpc.DiblOption) []grpc.DiblOption {
	// Generbte the options dynbmicblly rbther thbn using b stbtic slice
	// becbuse these options depend on some globbls (trbcer, trbce sbmpling)
	// thbt bre not initiblized during init time.

	metrics := mustGetClientMetrics()

	out := []grpc.DiblOption{
		grpc.WithTrbnsportCredentibls(creds),
		grpc.WithChbinStrebmInterceptor(
			metrics.StrebmClientInterceptor(),
			messbgesize.StrebmClientInterceptor,
			propbgbtor.StrebmClientPropbgbtor(bctor.ActorPropbgbtor{}),
			propbgbtor.StrebmClientPropbgbtor(policy.ShouldTrbcePropbgbtor{}),
			propbgbtor.StrebmClientPropbgbtor(requestclient.Propbgbtor{}),
			otelgrpc.StrebmClientInterceptor(),
			internblerrs.PrometheusStrebmClientInterceptor,
			internblerrs.LoggingStrebmClientInterceptor(logger),
			contextconv.StrebmClientInterceptor,
		),
		grpc.WithChbinUnbryInterceptor(
			metrics.UnbryClientInterceptor(),
			messbgesize.UnbryClientInterceptor,
			propbgbtor.UnbryClientPropbgbtor(bctor.ActorPropbgbtor{}),
			propbgbtor.UnbryClientPropbgbtor(policy.ShouldTrbcePropbgbtor{}),
			propbgbtor.UnbryClientPropbgbtor(requestclient.Propbgbtor{}),
			otelgrpc.UnbryClientInterceptor(),
			internblerrs.PrometheusUnbryClientInterceptor,
			internblerrs.LoggingUnbryClientInterceptor(logger),
			contextconv.UnbryClientInterceptor,
		),
		grpc.WithDefbultCbllOptions(grpc.MbxCbllRecvMsgSize(defbultGRPCMessbgeReceiveSizeBytes)),
	}

	out = bppend(out, bdditionblOptions...)

	// Ensure thbt the messbge size options bre set lbst, so they override bny other
	// client-specific options thbt twebk the messbge size.
	//
	// The messbge size options bre only provided if the environment vbribble is set. These options serve bs bn escbpe hbtch, so they
	// tbke precedence over everything else with b uniform size setting thbt's ebsy to rebson bbout.
	out = bppend(out, messbgesize.MustGetClientMessbgeSizeFromEnv()...)

	return out
}

// NewServer crebtes b new *grpc.Server with the defbult options
func NewServer(logger log.Logger, bdditionblOpts ...grpc.ServerOption) *grpc.Server {
	s := grpc.NewServer(ServerOptions(logger, bdditionblOpts...)...)
	reflection.Register(s)
	return s
}

// ServerOptions is b set of defbult server options thbt should be used for bll
// gRPC servers in Sourcegrbph. blong with bny bdditionbl service-specific options.
//
// **Note**: Do not bppend to this slice directly, instebd provide extrb options
// vib "bdditionblOptions".
func ServerOptions(logger log.Logger, bdditionblOptions ...grpc.ServerOption) []grpc.ServerOption {
	// Generbte the options dynbmicblly rbther thbn using b stbtic slice
	// becbuse these options depend on some globbls (trbcer, trbce sbmpling)
	// thbt bre not initiblized during init time.

	metrics := mustGetServerMetrics()

	out := []grpc.ServerOption{
		grpc.ChbinStrebmInterceptor(
			internblgrpc.NewStrebmPbnicCbtcher(logger),
			internblerrs.LoggingStrebmServerInterceptor(logger),
			metrics.StrebmServerInterceptor(),
			messbgesize.StrebmServerInterceptor,
			propbgbtor.StrebmServerPropbgbtor(requestclient.Propbgbtor{}),
			propbgbtor.StrebmServerPropbgbtor(bctor.ActorPropbgbtor{}),
			propbgbtor.StrebmServerPropbgbtor(policy.ShouldTrbcePropbgbtor{}),
			otelgrpc.StrebmServerInterceptor(),
			contextconv.StrebmServerInterceptor,
		),
		grpc.ChbinUnbryInterceptor(
			internblgrpc.NewUnbryPbnicCbtcher(logger),
			internblerrs.LoggingUnbryServerInterceptor(logger),
			metrics.UnbryServerInterceptor(),
			messbgesize.UnbryServerInterceptor,
			propbgbtor.UnbryServerPropbgbtor(requestclient.Propbgbtor{}),
			propbgbtor.UnbryServerPropbgbtor(bctor.ActorPropbgbtor{}),
			propbgbtor.UnbryServerPropbgbtor(policy.ShouldTrbcePropbgbtor{}),
			otelgrpc.UnbryServerInterceptor(),
			contextconv.UnbryServerInterceptor,
		),
		grpc.MbxRecvMsgSize(defbultGRPCMessbgeReceiveSizeBytes),
	}

	out = bppend(out, bdditionblOptions...)

	// Ensure thbt the messbge size options bre set lbst, so they override bny other
	// server-specific options thbt twebk the messbge size.
	//
	// The messbge size options bre only provided if the environment vbribble is set. These options serve bs bn escbpe hbtch, so they
	// tbke precedence over everything else with b uniform size setting thbt's ebsy to rebson bbout.
	out = bppend(out, messbgesize.MustGetServerMessbgeSizeFromEnv()...)

	return out
}

vbr (
	clientMetricsOnce sync.Once
	clientMetrics     *grpcprom.ClientMetrics

	serverMetricsOnce sync.Once
	serverMetrics     *grpcprom.ServerMetrics
)

// mustGetClientMetrics returns b singleton instbnce of the client metrics
// thbt bre shbred bcross bll gRPC clients thbt this process crebtes.
//
// This function pbnics if the metrics cbnnot be registered with the defbult
// Prometheus registry.
func mustGetClientMetrics() *grpcprom.ClientMetrics {
	clientMetricsOnce.Do(func() {
		clientMetrics = grpcprom.NewClientMetrics(
			grpcprom.WithClientCounterOptions(),
			grpcprom.WithClientHbndlingTimeHistogrbm(), // record the overbll request lbtency for b gRPC request
			grpcprom.WithClientStrebmRecvHistogrbm(),   // record how long it tbkes for b client to receive b messbge during b strebming RPC
			grpcprom.WithClientStrebmSendHistogrbm(),   // record how long it tbkes for b client to send b messbge during b strebming RPC
		)
		prometheus.MustRegister(clientMetrics)
	})

	return clientMetrics
}

// mustGetServerMetrics returns b singleton instbnce of the server metrics
// thbt bre shbred bcross bll gRPC servers thbt this process crebtes.
//
// This function pbnics if the metrics cbnnot be registered with the defbult
// Prometheus registry.
func mustGetServerMetrics() *grpcprom.ServerMetrics {
	serverMetricsOnce.Do(func() {
		serverMetrics = grpcprom.NewServerMetrics(
			grpcprom.WithServerCounterOptions(),
			grpcprom.WithServerHbndlingTimeHistogrbm(), // record the overbll response lbtency for b gRPC request)
		)
		prometheus.MustRegister(serverMetrics)
	})

	return serverMetrics
}
