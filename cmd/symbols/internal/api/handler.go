package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sourcegraph/go-ctags"
	logger "github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/symbols/proto"
	internaltypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

const maxNumSymbolResults = 500

type grpcService struct {
	searchFunc   types.SearchFunc
	readFileFunc func(context.Context, internaltypes.RepoCommitPath) ([]byte, error)
	ctagsBinary  string
	proto.UnimplementedSymbolsServer
	logger logger.Logger
}

func (s *grpcService) Search(ctx context.Context, r *proto.SearchRequest) (*proto.SymbolsResponse, error) {
	var response proto.SymbolsResponse

	params := r.ToInternal()
	symbols, err := s.searchFunc(ctx, params)
	if err != nil {
		s.logger.Error("symbol search failed",
			logger.String("arguments", fmt.Sprintf("%+v", params)),
			logger.Error(err),
		)

		response.FromInternal(&search.SymbolsResponse{Err: err.Error()})
	} else {
		response.FromInternal(&search.SymbolsResponse{Symbols: symbols})
	}

	return &response, nil
}

func (s *grpcService) ListLanguages(ctx context.Context, _ *emptypb.Empty) (*proto.ListLanguagesResponse, error) {
	rawMapping, err := ctags.ListLanguageMappings(ctx, s.ctagsBinary)
	if err != nil {
		return nil, errors.Wrap(err, "listing ctags language mappings")
	}

	protoMapping := make(map[string]*proto.ListLanguagesResponse_GlobFilePatterns, len(rawMapping))
	for language, filePatterns := range rawMapping {
		protoMapping[language] = &proto.ListLanguagesResponse_GlobFilePatterns{Patterns: filePatterns}
	}

	return &proto.ListLanguagesResponse{
		LanguageFileNameMap: protoMapping,
	}, nil
}

func (s *grpcService) Healthz(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	// Note: Kubernetes only has beta support for GRPC Healthchecks since version >= 1.23. This means
	// that we probably need the old non-GRPC healthcheck endpoint for a while.
	//
	// See https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-grpc-liveness-probe
	// for more information.
	return &emptypb.Empty{}, nil
}

func NewHandler(
	ctx context.Context,
	searchFunc types.SearchFunc,
	readFileFunc func(context.Context, internaltypes.RepoCommitPath) ([]byte, error),
	handleStatus func(http.ResponseWriter, *http.Request),
	ctagsBinary string,
) (handler http.Handler, startFn func() error) {

	searchFuncWrapper := func(ctx context.Context, args search.SymbolsParameters) (result.Symbols, error) {
		// Massage the arguments to ensure that First is set to a reasonable value.
		if args.First < 0 || args.First > maxNumSymbolResults {
			args.First = maxNumSymbolResults
		}

		return searchFunc(ctx, args)
	}

	rootLogger := logger.Scoped("symbolsServer", "symbols RPC server")

	// Initialize the gRPC server
	grpcServer := grpc.NewServer(
		defaults.ServerOptions(rootLogger)...,
	)
	grpcServer.RegisterService(&proto.Symbols_ServiceDesc, &grpcService{
		searchFunc:   searchFuncWrapper,
		readFileFunc: readFileFunc,
		ctagsBinary:  ctagsBinary,
		logger:       rootLogger.Scoped("grpc", "grpc server implementation"),
	})

	reflection.Register(grpcServer)

	localListener := newInMemoryListener("symbols")

	// setup grpcUI client connection options
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithContextDialer(localListener.ContextDial))
	opts = append(opts, defaults.DialOptions()...)
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	target := localListener.Addr().String()

	grpcUIHandler, grpcUIStartFn := NewGRPCUIHandler(ctx, target, opts...)

	jsonLogger := rootLogger.Scoped("jsonrpc", "json server implementation")

	// Initialize the legacy JSON API server
	mux := http.NewServeMux()
	mux.HandleFunc("/search", handleSearchWith(jsonLogger, searchFuncWrapper))
	mux.HandleFunc("/healthz", handleHealthCheck(jsonLogger))
	mux.HandleFunc("/list-languages", handleListLanguages(ctagsBinary))

	mux.Handle("/grpcui/", http.StripPrefix("/grpcui", grpcUIHandler))

	addHandlers(mux, searchFunc, readFileFunc)
	if handleStatus != nil {
		mux.HandleFunc("/status", handleStatus)
	}

	start := func() error {
		// TODO:@ggilmore: This seems a bit racy? Is there a way to ensure that the server
		// is listening before we start the grpcui server?
		go func() {
			if err := grpcServer.Serve(localListener); err != nil {
				rootLogger.Error("failed to serve HTTP server", logger.Error(err))
			}
		}()

		if err := grpcUIStartFn(); err != nil {
			return errors.Wrap(err, "initializing grpcui")
		}

		return nil
	}

	wrappedHandler := internalgrpc.MultiplexHandlers(grpcServer, mux)

	return wrappedHandler, start
}

func handleSearchWith(l logger.Logger, searchFunc types.SearchFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var args search.SymbolsParameters
		if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resultSymbols, err := searchFunc(r.Context(), args)
		if err != nil {
			// Ignore reporting errors where client disconnected
			if r.Context().Err() == context.Canceled && errors.Is(err, context.Canceled) {
				return
			}

			argsStr := fmt.Sprintf("%+v", args)

			l.Error("symbol search failed",
				logger.String("arguments", argsStr),
				logger.Error(err),
			)

			if err := json.NewEncoder(w).Encode(search.SymbolsResponse{Err: err.Error()}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if err := json.NewEncoder(w).Encode(search.SymbolsResponse{Symbols: resultSymbols}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func handleListLanguages(ctagsBinary string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		mapping, err := ctags.ListLanguageMappings(r.Context(), ctagsBinary)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(mapping); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func handleHealthCheck(l logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte("OK")); err != nil {
			l.Error("failed to write healthcheck response", logger.Error(err))
		}
	}
}
