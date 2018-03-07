package app

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/gobuildserver"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

type graphqlQuery struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// serveGoSymbolURL handles Go symbol URLs (e.g.,
// https://sourcegraph.com/go/github.com/gorilla/mux/-/Vars) by
// redirecting them to the file and line/column URL of the definition.
func serveGoSymbolURL(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid symbol URL path: %q", r.URL.Path)
	}
	mode := parts[0]
	symbolID := strings.Join(parts[1:], "/")

	if mode != "go" {
		return &errcode.HTTPErr{
			Status: http.StatusNotFound,
			Err:    errors.New("invalid mode (only \"go\" is supported"),
		}
	}

	importPath := strings.Split(symbolID, "/-/")[0]
	cloneURL, err := gobuildserver.ResolveImportPathCloneURL(httputil.CachingClient, importPath)
	if err != nil {
		return err
	}

	if cloneURL == "" || !strings.HasPrefix(cloneURL, "https://github.com") {
		return fmt.Errorf("non-github clone URL resolved for import path %s", importPath)
	}

	repoURI := api.RepoURI(strings.TrimSuffix(strings.TrimPrefix(cloneURL, "https://"), ".git"))
	repo, err := backend.Repos.GetByURI(ctx, repoURI)
	if err != nil {
		return err
	}
	if err := backend.Repos.RefreshIndex(ctx, repo); err != nil {
		return err
	}

	commitID, err := backend.Repos.ResolveRev(ctx, repo, "")
	if err != nil {
		return err
	}

	symbols, err := backend.Symbols.List(r.Context(), repo.URI, commitID, mode, lspext.WorkspaceSymbolParams{
		Symbol: lspext.SymbolDescriptor{"id": symbolID},
	})
	if err != nil {
		return err
	}

	if len(symbols) > 0 {
		symbol := symbols[0]
		uri, err := uri.Parse(string(symbol.Location.URI))
		if err != nil {
			return err
		}
		filePath := uri.Fragment
		dest := &url.URL{
			Path:     "/" + path.Join(string(repo.URI), "-/blob", filePath),
			Fragment: fmt.Sprintf("L%d:%d$references", symbol.Location.Range.Start.Line+1, symbol.Location.Range.Start.Character+1),
		}
		http.Redirect(w, r, dest.String(), http.StatusFound)
		return nil
	}

	return &errcode.HTTPErr{
		Status: http.StatusNotFound,
		Err:    errors.New("symbol not found"),
	}
}
