package httpapi

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"

	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
)

func serveRepoDefLanding(w http.ResponseWriter, r *http.Request) error {
	repo, repoRev, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		return errors.Wrap(err, "GetRepoAndRev")
	}

	// Parse query parameters.
	file := r.URL.Query().Get("file")
	line, err := strconv.Atoi(r.URL.Query().Get("line"))
	if err != nil {
		return errors.Wrap(err, "parsing line query param")
	}
	character, err := strconv.Atoi(r.URL.Query().Get("character"))
	if err != nil {
		return errors.Wrap(err, "parsing character query param")
	}

	// TODO: figure out how to handle other languages here.
	language := "go"

	// Lookup the symbol's information by performing textDocument/definition
	// and then looking through workspace/symbol results for the definition.
	rootPath := "git://" + repo.URI + "?" + repoRev.CommitID
	var locations []lsp.Location
	err = xlang.OneShotClientRequest(r.Context(), language, rootPath, "textDocument/definition", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: rootPath + "#" + file},
		Position:     lsp.Position{Line: line, Character: character},
	}, &locations)
	if len(locations) == 0 {
		return fmt.Errorf("textDocument/definition returned zero locations")
	}
	uri, err := url.Parse(locations[0].URI)
	if err != nil {
		return errors.Wrap(err, "parsing definition URL")
	}

	// Query workspace symbols.
	withoutFile := *uri
	withoutFile.Fragment = ""
	var symbols []lsp.SymbolInformation
	err = xlang.OneShotClientRequest(r.Context(), language, withoutFile.String(), "workspace/symbol", lsp.WorkspaceSymbolParams{
		// TODO(slimsag): before merge, performance for golang/go here is not
		// good. Allow specifying file URIs as a query filter. Sucks a bit that
		// textDocument/definition won't give us the Name/ContainerName that we
		// need!
		Query: "", // all symbols
	}, &symbols)

	// Find the matching symbol.
	var symbol *lsp.SymbolInformation
	for _, sym := range symbols {
		if sym.Location.URI != locations[0].URI {
			continue
		}
		if sym.Location.Range.Start.Line != locations[0].Range.Start.Line {
			continue
		}
		if sym.Location.Range.Start.Character != locations[0].Range.Start.Character {
			continue
		}
		symbol = &sym
		break
	}
	if symbol == nil {
		return fmt.Errorf("could not finding matching symbol info")
	}

	legacyURL, err := router.Rel.URLToLegacyDefLanding(*symbol)
	if err != nil {
		return errors.Wrap(err, "legacyDefLandingURL")
	}

	w.Header().Set("cache-control", "private, max-age=60")
	return writeJSON(w, &struct {
		URL string
	}{
		URL: legacyURL,
	})
}
