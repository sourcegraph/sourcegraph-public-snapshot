// Package lputil implements Language Processor utilities.
package lputil

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

// Translator is an HTTP handler which translates from the Language Processor
// REST API (defined in proto.go) directly to Sourcegraph LSP batch requests.
type Translator struct {
	// Addr is the address of the LSP server which translation should occur
	// against.
	Addr string

	// RootPath is invoked to determine the workspace directly at which all
	// file requests are relative to.
	RootPath func(repo, commit string) string
}

// ServeHTTP implements the http.Handler interface.
func (t *Translator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: use a mux in the future.
	err := t.serveHover(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		err2 := json.NewEncoder(w).Encode(&Error{
			Error: err.Error(),
		})
		if err2 != nil {
			// TODO: configurable logging
			log.Println(err2)
		}
	}
}

func (t *Translator) serveHover(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" || r.URL.Path != "/hover" || len(r.URL.Query()) > 0 {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	// TODO(slimsag): We don't need to create a new JSON RPC 2 connection every
	// time, but we will need reconnection logic and a non-dumb jsonrpc2.Client
	// which can handle concurrency (according to Sourcegraph LSP spec we can
	// and should use one connection for all requests).
	conn, err := net.Dial("tcp", t.Addr)
	if err != nil {
		return err
	}
	cl := jsonrpc2.NewClient(conn)
	defer func() {
		if err := cl.Close(); err != nil {
			// TODO: configurable logging
			log.Println(err)
		}
	}()

	// Decode the user request.
	var pos Position
	if err := json.NewDecoder(r.Body).Decode(&pos); err != nil {
		return err
	}
	if pos.Repo == "" {
		return fmt.Errorf("Repo field must be set")
	}
	if pos.Commit == "" {
		return fmt.Errorf("Commit field must be set")
	}
	if pos.File == "" {
		return fmt.Errorf("File field must be set")
	}

	// Build the LSP requests.
	reqInit := jsonrpc2.Request{
		ID:     "0",
		Method: "initialize",
	}
	rootPath := t.RootPath(pos.Repo, pos.Commit)
	log.Println("hover", rootPath)
	reqInit.SetParams(&lsp.InitializeParams{
		RootPath: rootPath,
	})
	// TODO: should probably check server capabilities before invoking hover,
	// but good enough for now.
	reqHoverID := "1"
	reqHover := jsonrpc2.Request{
		ID:     reqHoverID,
		Method: "textDocument/hover",
	}
	reqHover.SetParams(&lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: pos.File},
		Position: lsp.Position{
			Line:      pos.Line,
			Character: pos.Character,
		},
	})
	reqShutdown := jsonrpc2.Request{ID: "2", Method: "shutdown"}

	// Make the batched LSP request.
	resps, err := cl.RequestBatchAndWaitForAllResponses(
		reqInit,
		reqHover,
		reqShutdown,
	)
	if err != nil {
		return err
	}

	// Unmarshal the LSP responses.
	hoverResp, ok := resps[reqHoverID]
	if !ok {
		return fmt.Errorf("response to hover request from LSP server not found")
	}
	var respHover lsp.Hover
	if hoverResp.Result != nil {
		if err := json.Unmarshal(*hoverResp.Result, &respHover); err != nil {
			return err
		}
	}

	// Encode our response.
	final := &Hover{
		Contents: make([]HoverContent, len(respHover.Contents)),
	}
	for _, marked := range respHover.Contents {
		final.Contents = append(final.Contents, HoverContent{
			Type:  marked.Language,
			Value: marked.Value,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(final)
}
