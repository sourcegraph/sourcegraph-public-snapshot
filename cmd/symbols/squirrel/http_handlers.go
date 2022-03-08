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
)

// Responds to /localCodeIntel
func LocalCodeIntelHandler(w http.ResponseWriter, r *http.Request) {
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

	// Compute the local code intel payload.
	payload, err := localCodeIntel(r.Context(), args, readFileFromGitserver)
	if payload != nil && os.Getenv("SQUIRREL_DEBUG") == "true" {
		debugStringBuilder := &strings.Builder{}
		fmt.Fprintln(debugStringBuilder, "👉 /localCodeIntel repo:", args.Repo, "commit:", args.Commit, "path:", args.Path)
		contents, err := readFileFromGitserver(r.Context(), args)
		if err != nil {
			log15.Error("failed to read file from gitserver", "err", err)
		} else {
			prettyPrintLocalCodeIntelPayload(debugStringBuilder, args, *payload, string(contents))
			fmt.Fprintln(debugStringBuilder, "✅ /localCodeIntel repo:", args.Repo, "commit:", args.Commit, "path:", args.Path)

			fmt.Println(" ")
			fmt.Println(bracket(debugStringBuilder.String()))
			fmt.Println(" ")
		}
	}
	if err != nil {
		_ = json.NewEncoder(w).Encode(nil)
		log15.Error("failed to generate local code intel payload", "err", err)
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

// Responds to /symbolInfo
func NewSymbolInfoHandler(symbolSearch symbolsTypes.SearchFunc) func(w http.ResponseWriter, r *http.Request) {
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
		squirrel := NewSquirrelService(readFileFromGitserver, symbolSearch)
		result, err := squirrel.symbolInfo(r.Context(), args)
		if os.Getenv("SQUIRREL_DEBUG") == "true" {
			debugStringBuilder := &strings.Builder{}
			fmt.Fprintln(debugStringBuilder, "👉 /symbolInfo repo:", args.Repo, "commit:", args.Commit, "path:", args.Path, "row:", args.Row, "column:", args.Column)
			prettyPrintBreadcrumbs(debugStringBuilder, squirrel.breadcrumbs, readFileFromGitserver)
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
