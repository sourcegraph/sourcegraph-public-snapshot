//go:build cgo

package api

import (
	"context"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/squirrel"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	internaltypes "github.com/sourcegraph/sourcegraph/internal/types"
)

// addHandlers adds handlers that require cgo.
func addHandlers(
	mux *http.ServeMux,
	searchFunc types.SearchFunc,
	readFileFunc func(context.Context, internaltypes.RepoCommitPath) ([]byte, error),
) {
	mux.HandleFunc("/localCodeIntel", squirrel.LocalCodeIntelHandler(readFileFunc))
	mux.HandleFunc("/debugLocalCodeIntel", squirrel.DebugLocalCodeIntelHandler)
	mux.HandleFunc("/symbolInfo", squirrel.NewSymbolInfoHandler(searchFunc, readFileFunc))
}
