package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
)

type graphqlQuery struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// serveGoSymbolURL handles Go symbol URLs (e.g.,
// https://sourcegraph.com/go/github.com/gorilla/mux/-/Vars) by
// redirecting them to the file and line/column URL of the definition.
func serveGoSymbolURL(w http.ResponseWriter, r *http.Request) error {
	// Make a standard HTTP request to our GraphQL endpoint. This
	// works because we only care about symbol URLs for Go defs in
	// public repos.

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid symbol URL path: %q", r.URL.Path)
	}
	mode := parts[0]
	symbolID := strings.Join(parts[1:], "/")

	body := graphqlQuery{
		Query: `query Workbench($id: String, $mode: String) {
  root {
    symbols(id: $id, mode: $mode) {
      path
      line
      character
      repository {
        uri
      }
    }
  }
}`,
		Variables: map[string]interface{}{
			"id":   symbolID,
			"mode": mode,
			"rev":  "",
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	u := conf.AppURL.ResolveReference(&url.URL{Path: "/.api/graphql"}).String()
	req, err := http.Post(u, "application/json; charset=utf-8", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer req.Body.Close()

	var resp struct {
		Data struct {
			Root struct {
				Symbols []struct {
					Path       string `json:"path"`
					Line       int    `json:"line"`
					Character  int    `json:"character"`
					Repository struct {
						URI string `json:"uri"`
					}
				}
			} `json:"root"`
		} `json:"data"`
	}
	if err := json.NewDecoder(req.Body).Decode(&resp); err != nil {
		return err
	}

	if len(resp.Data.Root.Symbols) > 0 {
		symbol := resp.Data.Root.Symbols[0]
		dest := &url.URL{
			Path:     "/" + path.Join(symbol.Repository.URI, "-/blob", symbol.Path),
			Fragment: fmt.Sprintf("L%d:%d", symbol.Line+1, symbol.Character+1),
		}
		http.Redirect(w, r, dest.String(), http.StatusFound)
		return nil
	}

	return &errcode.HTTPErr{
		Status: http.StatusNotFound,
		Err:    errors.New("symbol not found"),
	}
}
