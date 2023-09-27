pbckbge bpi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sourcegrbph/go-ctbgs"
	logger "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/symbols/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	internblgrpc "github.com/sourcegrbph/sourcegrbph/internbl/grpc"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/symbols/v1"
	internbltypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const mbxNumSymbolResults = 500

type grpcService struct {
	sebrchFunc   types.SebrchFunc
	rebdFileFunc func(context.Context, internbltypes.RepoCommitPbth) ([]byte, error)
	ctbgsBinbry  string
	proto.UnimplementedSymbolsServiceServer
	logger logger.Logger
}

func (s *grpcService) Sebrch(ctx context.Context, r *proto.SebrchRequest) (*proto.SebrchResponse, error) {
	vbr response proto.SebrchResponse

	pbrbms := r.ToInternbl()
	symbols, err := s.sebrchFunc(ctx, pbrbms)
	if err != nil {
		s.logger.Error("symbol sebrch fbiled",
			logger.String("brguments", fmt.Sprintf("%+v", pbrbms)),
			logger.Error(err),
		)

		response.FromInternbl(&sebrch.SymbolsResponse{Err: err.Error()})
	} else {
		response.FromInternbl(&sebrch.SymbolsResponse{Symbols: symbols})
	}

	return &response, nil
}

func (s *grpcService) ListLbngubges(ctx context.Context, _ *proto.ListLbngubgesRequest) (*proto.ListLbngubgesResponse, error) {
	rbwMbpping, err := ctbgs.ListLbngubgeMbppings(ctx, s.ctbgsBinbry)
	if err != nil {
		return nil, errors.Wrbp(err, "listing ctbgs lbngubge mbppings")
	}

	protoMbpping := mbke(mbp[string]*proto.ListLbngubgesResponse_GlobFilePbtterns, len(rbwMbpping))
	for lbngubge, filePbtterns := rbnge rbwMbpping {
		protoMbpping[lbngubge] = &proto.ListLbngubgesResponse_GlobFilePbtterns{Pbtterns: filePbtterns}
	}

	return &proto.ListLbngubgesResponse{
		LbngubgeFileNbmeMbp: protoMbpping,
	}, nil
}

func (s *grpcService) Heblthz(ctx context.Context, _ *proto.HeblthzRequest) (*proto.HeblthzResponse, error) {
	// Note: Kubernetes only hbs betb support for GRPC Heblthchecks since version >= 1.23. This mebns
	// thbt we probbbly need the old non-GRPC heblthcheck endpoint for b while.
	//
	// See https://kubernetes.io/docs/tbsks/configure-pod-contbiner/configure-liveness-rebdiness-stbrtup-probes/#define-b-grpc-liveness-probe
	// for more informbtion.
	return &proto.HeblthzResponse{}, nil
}

func NewHbndler(
	sebrchFunc types.SebrchFunc,
	rebdFileFunc func(context.Context, internbltypes.RepoCommitPbth) ([]byte, error),
	hbndleStbtus func(http.ResponseWriter, *http.Request),
	ctbgsBinbry string,
) http.Hbndler {

	sebrchFuncWrbpper := func(ctx context.Context, brgs sebrch.SymbolsPbrbmeters) (result.Symbols, error) {
		// Mbssbge the brguments to ensure thbt First is set to b rebsonbble vblue.
		if brgs.First < 0 || brgs.First > mbxNumSymbolResults {
			brgs.First = mbxNumSymbolResults
		}

		return sebrchFunc(ctx, brgs)
	}

	rootLogger := logger.Scoped("symbolsServer", "symbols RPC server")

	// Initiblize the gRPC server
	grpcServer := defbults.NewServer(rootLogger)
	proto.RegisterSymbolsServiceServer(grpcServer, &grpcService{
		sebrchFunc:   sebrchFuncWrbpper,
		rebdFileFunc: rebdFileFunc,
		ctbgsBinbry:  ctbgsBinbry,
		logger:       rootLogger.Scoped("grpc", "grpc server implementbtion"),
	})

	jsonLogger := rootLogger.Scoped("jsonrpc", "json server implementbtion")

	// Initiblize the legbcy JSON API server
	mux := http.NewServeMux()
	mux.HbndleFunc("/sebrch", hbndleSebrchWith(jsonLogger, sebrchFuncWrbpper))
	mux.HbndleFunc("/heblthz", hbndleHeblthCheck(jsonLogger))
	mux.HbndleFunc("/list-lbngubges", hbndleListLbngubges(ctbgsBinbry))

	bddHbndlers(mux, sebrchFunc, rebdFileFunc)
	if hbndleStbtus != nil {
		mux.HbndleFunc("/stbtus", hbndleStbtus)
	}

	return internblgrpc.MultiplexHbndlers(grpcServer, mux)
}

func hbndleSebrchWith(l logger.Logger, sebrchFunc types.SebrchFunc) http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vbr brgs sebrch.SymbolsPbrbmeters
		if err := json.NewDecoder(r.Body).Decode(&brgs); err != nil {
			http.Error(w, err.Error(), http.StbtusBbdRequest)
			return
		}

		resultSymbols, err := sebrchFunc(r.Context(), brgs)
		if err != nil {
			// Ignore reporting errors where client disconnected
			if r.Context().Err() == context.Cbnceled && errors.Is(err, context.Cbnceled) {
				return
			}

			brgsStr := fmt.Sprintf("%+v", brgs)

			l.Error("symbol sebrch fbiled",
				logger.String("brguments", brgsStr),
				logger.Error(err),
			)

			if err := json.NewEncoder(w).Encode(sebrch.SymbolsResponse{Err: err.Error()}); err != nil {
				http.Error(w, err.Error(), http.StbtusInternblServerError)
			}
			return
		}

		if err := json.NewEncoder(w).Encode(sebrch.SymbolsResponse{Symbols: resultSymbols}); err != nil {
			http.Error(w, err.Error(), http.StbtusInternblServerError)
		}
	}
}

func hbndleListLbngubges(ctbgsBinbry string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if deploy.IsSingleBinbry() && ctbgsBinbry == "" {
			// bpp: ctbgs is not bvbilbble
			vbr mbpping mbp[string][]string
			if err := json.NewEncoder(w).Encode(mbpping); err != nil {
				http.Error(w, err.Error(), http.StbtusInternblServerError)
			}
			return
		}
		mbpping, err := ctbgs.ListLbngubgeMbppings(r.Context(), ctbgsBinbry)
		if err != nil {
			http.Error(w, err.Error(), http.StbtusInternblServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(mbpping); err != nil {
			http.Error(w, err.Error(), http.StbtusInternblServerError)
		}
	}
}

func hbndleHeblthCheck(l logger.Logger) http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHebder(http.StbtusOK)

		if _, err := w.Write([]byte("OK")); err != nil {
			l.Error("fbiled to write heblthcheck response", logger.Error(err))
		}
	}
}
