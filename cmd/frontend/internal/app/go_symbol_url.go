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
	symbols(id: $id, mode: $mode) {
		path
		line
		character
		repository {
		uri
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

	req, err := http.NewRequest("POST", "http://localhost:3080/.api/graphql", bytes.NewReader(data))
	if err != nil {
		return err
	}
	for k := range r.Header {
		req.Header[k] = r.Header[k]
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	symbolResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer symbolResp.Body.Close()

	var resp struct {
		Data struct {
			Symbols []struct {
				Path       string `json:"path"`
				Line       int    `json:"line"`
				Character  int    `json:"character"`
				Repository struct {
					URI string `json:"uri"`
				}
			}
		} `json:"data"`
	}
	if err := json.NewDecoder(symbolResp.Body).Decode(&resp); err != nil {
		return err
	}

	if len(resp.Data.Symbols) > 0 {
		symbol := resp.Data.Symbols[0]
		dest := &url.URL{
			Path:     "/" + path.Join(symbol.Repository.URI, "-/blob", symbol.Path),
			Fragment: fmt.Sprintf("L%d:%d$references", symbol.Line+1, symbol.Character+1),
		}
		http.Redirect(w, r, dest.String(), http.StatusFound)
		return nil
	}

	return &errcode.HTTPErr{
		Status: http.StatusNotFound,
		Err:    errors.New("symbol not found"),
	}
}
