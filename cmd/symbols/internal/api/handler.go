package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sourcegraph/go-ctags"
	logger "github.com/sourcegraph/log"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/proto"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	internaltypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"

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
	var params search.SymbolsParameters
	params.FromProto(r)

	result, err := s.searchFunc(ctx, params)
	if err != nil {
		arguments := fmt.Sprintf("%+v", params)

		s.logger.Error("search failed",
			logger.String("arguments", arguments),
			logger.Error(err),
		)

		response := search.SymbolsResponse{Err: err.Error()}
		return response.ToProto(), nil
	}

	response := search.SymbolsResponse{Symbols: result}
	return response.ToProto(), nil
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

	// Initialize the gRPC server
	grpcServer := grpc.NewServer()
	grpcServer.RegisterService(&proto.Symbols_ServiceDesc, &grpcService{
		searchFunc:   searchFuncWrapper,
		readFileFunc: readFileFunc,
		ctagsBinary:  ctagsBinary,
		logger:       logger.Scoped("symbolsGRPCServer", "symbols service grpc server"),
	})
	reflection.Register(grpcServer)

	// Initialize the legacy JSON API server
	mux := http.NewServeMux()
	mux.HandleFunc("/search", handleSearchWith(searchFuncWrapper))
	mux.HandleFunc("/healthz", handleHealthCheck)
	mux.HandleFunc("/list-languages", handleListLanguages(ctagsBinary))
	addHandlers(mux, searchFunc, readFileFunc)
	if handleStatus != nil {
		mux.HandleFunc("/status", handleStatus)
	}

	return internalgrpc.MultiplexHandlers(grpcServer, mux)
}

func handleSearchWith(searchFunc types.SearchFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var args search.SymbolsParameters
		if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := searchFunc(r.Context(), args)
		if err != nil {
			// Ignore reporting errors where client disconnected
			if r.Context().Err() == context.Canceled && errors.Is(err, context.Canceled) {
				return
			}

			log15.Error("Symbol search failed", "args", args, "error", err)
			if err := json.NewEncoder(w).Encode(search.SymbolsResponse{Err: err.Error()}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if err := json.NewEncoder(w).Encode(search.SymbolsResponse{Symbols: result}); err != nil {
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

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte("OK")); err != nil {
		log15.Error("failed to write response to health check, err: %s", err)
	}
}
