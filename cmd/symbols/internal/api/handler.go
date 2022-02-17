package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
)

func NewHandler(
	searchFunc types.SearchFunc,
	handleStatus func(http.ResponseWriter, *http.Request),
	ctagsBinary string,
) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/search", handleSearchWith(searchFunc))
	mux.HandleFunc("/healthz", handleHealthCheck)
	mux.HandleFunc("/list-languages", handleListLanguages(ctagsBinary))
	if handleStatus != nil {
		mux.HandleFunc("/status", handleStatus)
	}
	return mux
}

const maxNumSymbolResults = 500

func handleSearchWith(searchFunc types.SearchFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var args types.SearchArgs
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
			if r.Context().Err() == context.Canceled && isContextCanceled(err) {
				return
			}

			log15.Error("Symbol search failed", "args", args, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(result); err != nil {
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

// isContextCanceled checks if the error is a context canceled error. Although `errors.Is(err,
// context.Canceled)` appears to be the correct way to check for this, it doesn't work for some reason.
func isContextCanceled(err error) bool {
	return strings.HasSuffix(err.Error(), "context canceled")
}
