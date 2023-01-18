package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/go-ctags"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/proto"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/search"
	internaltypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"google.golang.org/grpc/reflection"
)

const maxNumSymbolResults = 500

type grpcServer struct {
	searchFunc   types.SearchFunc
	readFileFunc func(context.Context, internaltypes.RepoCommitPath) ([]byte, error)
	ctagsBinary  string
	proto.UnimplementedSymbolsServer
}

func (s *grpcServer) Search(ctx context.Context, r *proto.SearchRequest) (*proto.SymbolsResponse, error) {
	var params search.SymbolsParameters
	params.FromProto(r)

	result, err := s.searchFunc(ctx, params)
	if err != nil {
		// How do we handle client disconnections? Can we test this?

		log15.Error("Symbol search failed", "args", params, "error", err) // straight up copying this for now, find another abstraction

		response := search.SymbolsResponse{Err: err.Error()}
		return response.ToProto(), nil

	}

	response := search.SymbolsResponse{Symbols: result}
	return response.ToProto(), nil
}

func (s *grpcServer) ListLanguages(ctx context.Context, _ *emptypb.Empty) (*proto.ListLanguagesResponse, error) {
	rawMapping, err := ctags.ListLanguageMappings(ctx, s.ctagsBinary)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list language mappings")
	}

	mapping := make(map[string]*proto.ListLanguagesResponse_GlobFilePatterns, len(rawMapping))
	for language, filePatterns := range rawMapping {
		mapping[language] = &proto.ListLanguagesResponse_GlobFilePatterns{Patterns: filePatterns}
	}

	return &proto.ListLanguagesResponse{
		LanguageFileNameMap: mapping,
	}, nil
}

func (s *grpcServer) Healthz(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
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
	s := grpc.NewServer()
	s.RegisterService(&proto.Symbols_ServiceDesc, &grpcServer{
		searchFunc:   searchFuncWrapper,
		readFileFunc: readFileFunc,
		ctagsBinary:  ctagsBinary,
	})
	reflection.Register(s)

	// Initialize the legacy JSON API server
	mux := http.NewServeMux()
	mux.HandleFunc("/search", handleSearchWith(searchFuncWrapper))
	mux.HandleFunc("/healthz", handleHealthCheck)
	mux.HandleFunc("/list-languages", handleListLanguages(ctagsBinary))
	addHandlers(mux, searchFunc, readFileFunc)
	if handleStatus != nil {
		mux.HandleFunc("/status", handleStatus)
	}

	handler := h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the request is for the gRPC server, serve it.
		if r.ProtoMajor == 2 && r.Header.Get("Content-Type") == "application/grpc" {
			s.ServeHTTP(w, r)
			return
		}

		// Otherwise, serve the JSON API server.
		mux.ServeHTTP(w, r)
	}), &http2.Server{})

	return handler
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
