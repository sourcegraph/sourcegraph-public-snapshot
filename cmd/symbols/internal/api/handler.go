package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	proto "github.com/sourcegraph/sourcegraph/internal/symbols/v1"
	internaltypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const maxNumSymbolResults = 500

type grpcService struct {
	searchFunc   types.SearchFunc
	readFileFunc func(context.Context, internaltypes.RepoCommitPath) ([]byte, error)
	ctagsBinary  string
	proto.UnimplementedSymbolsServiceServer
	logger logger.Logger
}

func (s *grpcService) Search(ctx context.Context, r *proto.SearchRequest) (*proto.SearchResponse, error) {
	var response proto.SearchResponse

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

func (s *grpcService) Healthz(ctx context.Context, _ *proto.HealthzRequest) (*proto.HealthzResponse, error) {
	// Note: Kubernetes only has beta support for GRPC Healthchecks since version >= 1.23. This means
	// that we probably need the old non-GRPC healthcheck endpoint for a while.
	//
	// See https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-grpc-liveness-probe
	// for more information.
	return &proto.HealthzResponse{}, nil
}

func NewHandler(
	searchFunc types.SearchFunc,
	readFileFunc func(context.Context, internaltypes.RepoCommitPath) ([]byte, error),
	handleStatus func(http.ResponseWriter, *http.Request),
	ctagsBinary string,
) http.Handler {

	searchFuncWrapper := func(ctx context.Context, args search.SymbolsParameters) (result.Symbols, error) {
		// Massage the arguments to ensure that First is set to a reasonable value.
		if args.First < 0 || args.First > maxNumSymbolResults {
			args.First = maxNumSymbolResults
		}

		return searchFunc(ctx, args)
	}

	rootLogger := logger.Scoped("symbolsServer")

	// Initialize the gRPC server
	grpcServer := defaults.NewServer(rootLogger)
	proto.RegisterSymbolsServiceServer(grpcServer, &grpcService{
		searchFunc:   searchFuncWrapper,
		readFileFunc: readFileFunc,
		ctagsBinary:  ctagsBinary,
		logger:       rootLogger.Scoped("grpc"),
	})

	jsonLogger := rootLogger.Scoped("jsonrpc")

	// Initialize the legacy JSON API server
	mux := http.NewServeMux()
	mux.HandleFunc("/search", handleSearchWith(jsonLogger, searchFuncWrapper))
	mux.HandleFunc("/healthz", handleHealthCheck(jsonLogger))

	addHandlers(mux, searchFunc, readFileFunc)
	if handleStatus != nil {
		mux.HandleFunc("/status", handleStatus)
	}

	return internalgrpc.MultiplexHandlers(grpcServer, mux)
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

func handleHealthCheck(l logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte("OK")); err != nil {
			l.Error("failed to write healthcheck response", logger.Error(err))
		}
	}
}
