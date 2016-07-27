// Package langp implements Language Processor utilities.
package langp

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

// Translator is an HTTP handler which translates from the Language Processor
// REST API (defined in proto.go) directly to Sourcegraph LSP batch requests.
type Translator struct {
	// Addr is the address of the LSP server which translation should occur
	// against.
	Addr string

	// WorkDir is where workspaces are created by cloning repositories and
	// dependencies.
	WorkDir string

	// Prepare is called when the language processor should prepare the given
	// workspace directory with the given repository at the given commit.
	//
	// This is where language processors should perform language-specific tasks
	// like downloading dependencies via 'go get', etc. into the workspace
	// directory.
	//
	// If an error is returned, it is returned directly to the person who made
	// the API request which triggered the preperation of the workspace.
	Prepare func(workspace, repo, commit string) error

	// FileURI, if non-nil, is called to form the file URI which is sent to a
	// language server. Provided is the repo and commit, and a file URI which
	// is relative to the repository root.
	//
	// FileURI should return a file URI which is relative to the previously
	// prepared workspace directory.
	FileURI func(repo, commit, file string) string
}

// ServeHTTP implements the http.Handler interface.
func (t *Translator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: use a mux in the future.
	err := t.serveHover(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		err2 := json.NewEncoder(w).Encode(&Error{
			ErrorMsg: err.Error(),
		})
		if err2 != nil {
			// TODO: configurable logging
			log.Println(err2)
		}
	}
}

// prepareWorkspace prepares a new workspace for the given repository and
// revision.
func (t *Translator) prepareWorkspace(rootDir, repo, commit string) error {
	// Ensure the workspace directory exists.
	_, err := os.Stat(rootDir)
	if !os.IsNotExist(err) {
		// Workspace exists already.
		return nil
	}

	if err := os.MkdirAll(rootDir, 0700); err != nil {
		return err
	}
	if err := t.Prepare(rootDir, repo, commit); err != nil {
		// Preparing the workspace has failed, and thus the workspace is
		// incomplete. Remove the directory so that the next request causes
		// preparation again (this is our best chance at keeping the workspace
		// in a working state).
		if err2 := os.RemoveAll(rootDir); err2 != nil {
			log.Println(err2)
		}
		return err
	}
	return nil
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

	// Determine the root path for the workspace and prepare it.
	rootPath := filepath.Join(t.WorkDir, pos.Repo, pos.Commit)
	err = t.prepareWorkspace(rootPath, pos.Repo, pos.Commit)
	if err != nil {
		return err
	}

	// Build the LSP requests.
	reqInit := jsonrpc2.Request{
		ID:     "0",
		Method: "initialize",
	}
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
	fileURI := pos.File
	if t.FileURI != nil {
		fileURI = t.FileURI(pos.Repo, pos.Commit, pos.File)
	}
	reqHover.SetParams(&lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: fileURI},
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
	if hoverResp.Error != nil {
		return hoverResp.Error
	}
	var respHover lsp.Hover
	if err := json.Unmarshal(*hoverResp.Result, &respHover); err != nil {
		return err
	}

	// Encode our response.
	final := &Hover{
		Contents: make([]HoverContent, len(respHover.Contents)),
	}
	for i, marked := range respHover.Contents {
		final.Contents[i] = HoverContent{
			Type:  marked.Language,
			Value: marked.Value,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(final)
}

// ExpandSGPath expands the $SGPATH variable in the given string, except it
// uses ~/.sourcegraph as the default if $SGPATH is not set.
func ExpandSGPath(s string) (string, error) {
	sgpath := os.Getenv("SGPATH")
	if sgpath == "" {
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		sgpath = filepath.Join(u.HomeDir, ".sourcegraph")
	}
	return strings.Replace(s, "$SGPATH", sgpath, -1), nil
}
