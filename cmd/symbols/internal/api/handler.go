package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/proto"
	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/search"
	internaltypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type grpcServer struct {
	searchFunc types.SearchFunc
	proto.UnimplementedSymbolsServer
}

func (s *grpcServer) Search(ctx context.Context, r *proto.SearchRequest) (*proto.SymbolsResponse, error) {
	args := searchParamsFromProto(r)

	// TODO@ggilmore: do better than just copying this logic.
	if args.First < 0 || args.First > maxNumSymbolResults {
		args.First = maxNumSymbolResults
	}

	result, err := s.searchFunc(ctx, args)
	if err != nil {
		// How do we handle client disconnections? Can we test this?

		log15.Error("Symbol search failed", "args", args, "error", err) // straight up copying this for now, find another abstraction
		return symbolsResponseToProto(&search.SymbolsResponse{Err: err.Error()}), nil

	}
	return symbolsResponseToProto(&search.SymbolsResponse{Symbols: result}), nil
}

func NewHandler(
	searchFunc types.SearchFunc,
	readFileFunc func(context.Context, internaltypes.RepoCommitPath) ([]byte, error),
	handleStatus func(http.ResponseWriter, *http.Request),
	ctagsBinary string,
) http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("/search", handleSearchWith(searchFunc))
	mux.HandleFunc("/healthz", handleHealthCheck)
	mux.HandleFunc("/list-languages", handleListLanguages(ctagsBinary))
	addHandlers(mux, searchFunc, readFileFunc)
	if handleStatus != nil {
		mux.HandleFunc("/status", handleStatus)
	}
	return mux
}

const maxNumSymbolResults = 500

func handleSearchWith(searchFunc types.SearchFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var args search.SymbolsParameters
		if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if args.First < 0 || args.First > maxNumSymbolResults {
			args.First = maxNumSymbolResults
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

// searchParamsFromProto converts a proto.SearchParameters to a search.SymbolsParameters.
// TODO: Find a better place for this logic to live.
func searchParamsFromProto(r *proto.SearchRequest) search.SymbolsParameters {
	return search.SymbolsParameters{
		Repo:            api.RepoName(r.Repo), // TODO@ggilmore: This api.RepoName is just a go type alias - is it worth creating a new message type just for this?
		CommitID:        api.CommitID(r.CommitId),
		Query:           r.Query,
		IsRegExp:        r.IsRegExp,
		IsCaseSensitive: r.IsCaseSensitive,
		IncludePatterns: r.IncludePatterns,
		ExcludePattern:  r.ExcludePattern,
		First:           int(r.First),
		Timeout:         int(r.Timeout),
	}
}

func symbolsResponseToProto(r *search.SymbolsResponse) *proto.SymbolsResponse {
	if r.Err != "" {
		return &proto.SymbolsResponse{Error: &r.Err}
	}

	var symbolsList []*proto.Symbol

	for _, r := range r.Symbols {
		symbolsList = append(symbolsList, &proto.Symbol{
			Name:        r.Name,
			Path:        r.Path,
			Line:        int32(r.Line),
			Character:   int32(r.Character),
			Kind:        r.Kind,
			Language:    r.Language,
			Parent:      r.Parent,
			ParentKind:  r.ParentKind,
			Signature:   r.Signature,
			FileLimited: r.FileLimited,
		})
	}

	return &proto.SymbolsResponse{
		Symbols: &proto.SymbolsList{Symbols: symbolsList},
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
