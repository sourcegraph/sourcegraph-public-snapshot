package squirrel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/inconshreveable/log15"
	symbolsTypes "github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// LocalCodeIntelHandler responds to /localCodeIntel
func LocalCodeIntelHandler(readFile readFileFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the args from the request body.
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log15.Error("failed to read request body", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var args types.RepoCommitPath
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&args); err != nil {
			log15.Error("failed to decode request body", "err", err, "body", string(body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		squirrel := New(readFile, nil)
		defer squirrel.Close()

		// Compute the local code intel payload.
		payload, err := squirrel.LocalCodeIntel(r.Context(), args)
		if payload != nil && os.Getenv("SQUIRREL_DEBUG") == "true" {
			debugStringBuilder := &strings.Builder{}
			fmt.Fprintln(debugStringBuilder, "👉 /localCodeIntel repo:", args.Repo, "commit:", args.Commit, "path:", args.Path)
			contents, err := readFile(r.Context(), args)
			if err != nil {
				log15.Error("failed to read file from gitserver", "err", err)
			} else {
				prettyPrintLocalCodeIntelPayload(debugStringBuilder, *payload, string(contents))
				fmt.Fprintln(debugStringBuilder, "✅ /localCodeIntel repo:", args.Repo, "commit:", args.Commit, "path:", args.Path)

				fmt.Println(" ")
				fmt.Println(bracket(debugStringBuilder.String()))
				fmt.Println(" ")
			}
		}
		if err != nil {
			_ = json.NewEncoder(w).Encode(nil)

			// Log the error if it's not an unrecognized file extension or unsupported language error.
			if !errors.Is(err, UnrecognizedFileExtensionError) && !errors.Is(err, UnsupportedLanguageError) {
				log15.Error("failed to generate local code intel payload", "err", err)
			}

			return
		}

		// Write the response.
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(payload)
		if err != nil {
			log15.Error("failed to write response: %s", "error", err)
			http.Error(w, fmt.Sprintf("failed to generate local code intel payload: %s", err), http.StatusInternalServerError)
			return
		}
	}
}

// NewSymbolInfoHandler responds to /symbolInfo
func NewSymbolInfoHandler(symbolSearch symbolsTypes.SearchFunc, readFile readFileFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the args from the request body.
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log15.Error("failed to read request body", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var args types.RepoCommitPathPoint
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&args); err != nil {
			log15.Error("failed to decode request body", "err", err, "body", string(body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Find the symbol.
		squirrel := New(readFile, symbolSearch)
		defer squirrel.Close()
		fmt.Println("Calling SymbolInfo from NewSymbolInfoHandler")
		result, err := squirrel.SymbolInfo(r.Context(), args)
		if os.Getenv("SQUIRREL_DEBUG") == "true" {
			debugStringBuilder := &strings.Builder{}
			fmt.Fprintln(debugStringBuilder, "👉 /symbolInfo repo:", args.Repo, "commit:", args.Commit, "path:", args.Path, "row:", args.Row, "column:", args.Column)
			squirrel.breadcrumbs.pretty(debugStringBuilder, readFile)
			if result == nil {
				fmt.Fprintln(debugStringBuilder, "❌ no definition found")
			} else {
				fmt.Fprintln(debugStringBuilder, "✅ /symbolInfo", *result)
			}

			fmt.Println(" ")
			fmt.Println(bracket(debugStringBuilder.String()))
			fmt.Println(" ")
		}
		if err != nil {
			_ = json.NewEncoder(w).Encode(nil)
			log15.Error("failed to get definition", "err", err)
			return
		}

		// Write the response.
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			log15.Error("failed to write response: %s", "error", err)
			http.Error(w, fmt.Sprintf("failed to get definition: %s", err), http.StatusInternalServerError)
			return
		}
	}
}
