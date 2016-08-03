// Package langp implements Language Processor utilities.
package langp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

	// PrepareRepo is called when the language processor should clone the given
	// repository into the specified workspace at a subdirectory desired by the
	// language.
	//
	// If an error is returned, it is returned directly to the person who made
	// the API request which triggered the preperation of the workspace.
	PrepareRepo func(workspace, repo, commit string) error

	// PrepareDeps is called when the language processor should prepare the
	// dependencies for the given workspace/repo/commit.
	//
	// This is where language processors should perform language-specific tasks
	// like downloading dependencies via 'go get', etc. into the workspace
	// directory.
	//
	// If an error is returned, it is returned directly to the person who made
	// the API request which triggered the preperation of the workspace.
	PrepareDeps func(workspace, repo, commit string) error

	// FileURI, if non-nil, is called to form the file URI which is sent to a
	// language server. Provided is the repo and commit, and a file URI which
	// is relative to the repository root.
	//
	// FileURI should return a file URI which is relative to the previously
	// prepared workspace directory.
	FileURI func(repo, commit, file string) string
}

// New creates a new HTTP handler which translates from the Language Processor
// REST API (defined in proto.go) directly to Sourcegraph LSP batch requests
// using the specified options.
func New(opts *Translator) http.Handler {
	t := &translator{
		Translator:     opts,
		preparingRepos: newPending(),
		preparingDeps:  newPending(),
	}
	mux := http.NewServeMux()
	mux.Handle("/definition", t.handler("/definition", t.serveDefinition))
	mux.Handle("/external-refs", t.handler("/external-refs", t.serveExternalRefs))
	mux.Handle("/hover", t.handler("/hover", t.serveHover))
	return mux
}

type pending struct {
	*sync.Mutex
	m map[string]bool
}

func newPending() *pending {
	return &pending{
		Mutex: &sync.Mutex{},
		m:     make(map[string]bool),
	}
}

// acquire acquires ownership of preparing k. If k is already being prepared
// by someone else, this methods blocks until preparation of k is finished
// and handled=true is returned.
func (p *pending) acquire(k string, timeout time.Duration) (wasTimeout, handled bool) {
	// If nobody is preparing k, mark ownership and return:
	p.Lock()
	if _, pending := p.m[k]; !pending {
		p.m[k] = true
		p.Unlock()
		handled = false
		return
	}
	p.Unlock()

	// Someone is preparing k, wait for completion.
	start := time.Now()
	for {
		p.Lock()
		_, pending := p.m[k]
		p.Unlock()
		if !pending {
			handled = true
			return
		}
		if time.Since(start) > timeout {
			wasTimeout = true
			return
		}
		// TODO: timeout request
		time.Sleep(1 * time.Millisecond)
	}
}

// done is called to signal completion for preparation previously acquired by
// calling acquire.
func (p *pending) done(k string) {
	p.Lock()
	_, pending := p.m[k]
	if !pending {
		panic("pending: done() called for non-acquired k")
	}
	delete(p.m, k)
	p.Unlock()
}

type translator struct {
	*Translator
	mux                           *http.ServeMux
	preparingRepos, preparingDeps *pending
}

type handlerFunc func(body []byte) (interface{}, error)

// handler returns an HTTP handler which serves the given method at the
// specified path.
func (t *translator) handler(path string, m handlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Confirm the path because http.ServeMux is fuzzy.
		if r.URL.Path != path {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// All Language Processor methods are POST and have no query
		// parameters.
		if r.Method != "POST" || len(r.URL.Query()) > 0 {
			resp := &Error{ErrorMsg: "expected POST with no query parameters"}
			t.writeResponse(w, http.StatusBadRequest, resp, path, nil)
			return
		}

		// Handle the request.
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			resp := &Error{ErrorMsg: err.Error()}
			t.writeResponse(w, http.StatusBadRequest, resp, path, body)
			return
		}
		resp, err := m(body)
		if err != nil {
			resp := &Error{ErrorMsg: err.Error()}
			t.writeResponse(w, http.StatusBadRequest, resp, path, body)
			return
		}
		t.writeResponse(w, http.StatusOK, resp, path, body)
	})
}

func (t *translator) writeResponse(w http.ResponseWriter, status int, v interface{}, path string, body []byte) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	respBody, err := json.Marshal(v)
	if err != nil {
		// TODO: configurable logging
		log.Println(err)
	}
	_, err = io.Copy(w, bytes.NewReader(respBody))
	if err != nil {
		// TODO: configurable logging
		log.Println(err)
	}

	// TODO: configurable logging
	log.Printf("POST %s -> %d %s\n\tBody:     %s\n\tResponse: %s\n", path, status, http.StatusText(status), string(body), string(respBody))
}

// prepareWorkspace prepares a new workspace for the given repository and
// revision.
func (t *translator) prepareWorkspace(rootDir, repo, commit string) error {
	// Ensure the workspace directory exists.
	_, err := os.Stat(rootDir)
	if !os.IsNotExist(err) {
		// Workspace exists already.
		return nil
	}

	// Acquire ownership of repository preparation.
	//
	// TODO(slimsag): use a smaller timeout which propagates nicely to the
	// frontend.
	timeout, handled := t.preparingRepos.acquire(rootDir, 1*time.Hour)
	if timeout || handled {
		// A different request prepared the repository.
		return nil
	}
	defer t.preparingRepos.done(rootDir)

	// Prepare the workspace by creating the directory and cloning the
	// repository.
	if err := os.MkdirAll(rootDir, 0700); err != nil {
		return err
	}
	if err := t.PrepareRepo(rootDir, repo, commit); err != nil {
		// Preparing the workspace has failed, and thus the workspace is
		// incomplete. Remove the directory so that the next request causes
		// preparation again (this is our best chance at keeping the workspace
		// in a working state).
		log.Println("preparing workspace repo:", err)
		if err2 := os.RemoveAll(rootDir); err2 != nil {
			log.Println(err2)
		}
		return err
	}

	// Acquire ownership of dependency preparation.
	timeout, handled = t.preparingDeps.acquire(rootDir, 0*time.Second)
	if timeout || handled {
		// A different request is preparing the dependencies.
		return nil
	}

	// Prepare the dependencies asynchronously.
	go func() {
		defer t.preparingDeps.done(rootDir)
		if err := t.PrepareDeps(rootDir, repo, commit); err != nil {
			// Preparing the workspace has failed, and thus the workspace is
			// incomplete. Remove the directory so that the next request causes
			// preparation again (this is our best chance at keeping the workspace
			// in a working state).
			log.Println("preparing workspace deps:", err)
			if err2 := os.RemoveAll(rootDir); err2 != nil {
				log.Println(err2)
			}
		}
	}()
	return nil
}

func (t *translator) serveDefinition(body []byte) (interface{}, error) {
	// TODO(slimsag): We don't need to create a new JSON RPC 2 connection every
	// time, but we will need reconnection logic and a non-dumb jsonrpc2.Client
	// which can handle concurrency (according to Sourcegraph LSP spec we can
	// and should use one connection for all requests).
	conn, err := net.Dial("tcp", t.Addr)
	if err != nil {
		return nil, err
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
	if err := json.Unmarshal(body, &pos); err != nil {
		return nil, err
	}
	if pos.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if pos.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	if pos.File == "" {
		return nil, fmt.Errorf("File field must be set")
	}

	// Determine the root path for the workspace and prepare it.
	rootPath := filepath.Join(t.WorkDir, pos.Repo, pos.Commit)
	err = t.prepareWorkspace(rootPath, pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}

	// Build the LSP requests.
	reqInit := &jsonrpc2.Request{
		ID:     "0",
		Method: "initialize",
	}
	if err := reqInit.SetParams(&lsp.InitializeParams{
		RootPath: rootPath,
	}); err != nil {
		return nil, err
	}
	// TODO: should probably check server capabilities before invoking hover,
	// but good enough for now.
	reqDefID := "1"
	reqDef := &jsonrpc2.Request{
		ID:     reqDefID,
		Method: "textDocument/definition",
	}
	p := pos.LSP()
	if t.FileURI != nil {
		p.TextDocument.URI = t.FileURI(pos.Repo, pos.Commit, pos.File)
	}
	if err := reqDef.SetParams(p); err != nil {
		return nil, err
	}
	reqShutdown := &jsonrpc2.Request{ID: "2", Method: "shutdown"}

	// Make the batched LSP request.
	resps, err := cl.RequestBatchAndWaitForAllResponses(
		reqInit,
		reqDef,
		reqShutdown,
	)
	if err != nil {
		return nil, err
	}

	// Unmarshal the LSP responses.
	defResp, ok := resps[reqDefID]
	if !ok {
		return nil, fmt.Errorf("response to def request from LSP server not found")
	}
	if defResp.Error != nil {
		return nil, defResp.Error
	}
	// TODO: according to spec this could be lsp.Location OR []lsp.Location
	var respDef []lsp.Location
	if err := json.Unmarshal(*defResp.Result, &respDef); err != nil {
		return nil, err
	}
	return Range{
		Repo:           pos.Repo,
		Commit:         pos.Commit,
		File:           respDef[0].URI,
		StartLine:      respDef[0].Range.Start.Line,
		StartCharacter: respDef[0].Range.Start.Character,
		EndLine:        respDef[0].Range.End.Line,
		EndCharacter:   respDef[0].Range.End.Character,
	}, nil
}

func (t *translator) serveHover(body []byte) (interface{}, error) {
	// TODO(slimsag): We don't need to create a new JSON RPC 2 connection every
	// time, but we will need reconnection logic and a non-dumb jsonrpc2.Client
	// which can handle concurrency (according to Sourcegraph LSP spec we can
	// and should use one connection for all requests).
	conn, err := net.Dial("tcp", t.Addr)
	if err != nil {
		return nil, err
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
	if err := json.Unmarshal(body, &pos); err != nil {
		return nil, err
	}
	if pos.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if pos.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	if pos.File == "" {
		return nil, fmt.Errorf("File field must be set")
	}

	// Determine the root path for the workspace and prepare it.
	rootPath := filepath.Join(t.WorkDir, pos.Repo, pos.Commit)
	err = t.prepareWorkspace(rootPath, pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}

	// Build the LSP requests.
	reqInit := &jsonrpc2.Request{
		ID:     "0",
		Method: "initialize",
	}
	if err := reqInit.SetParams(&lsp.InitializeParams{
		RootPath: rootPath,
	}); err != nil {
		return nil, err
	}
	// TODO: should probably check server capabilities before invoking hover,
	// but good enough for now.
	reqHoverID := "1"
	reqHover := &jsonrpc2.Request{
		ID:     reqHoverID,
		Method: "textDocument/hover",
	}
	p := pos.LSP()
	if t.FileURI != nil {
		p.TextDocument.URI = t.FileURI(pos.Repo, pos.Commit, pos.File)
	}
	if err := reqHover.SetParams(p); err != nil {
		return nil, err
	}
	reqShutdown := &jsonrpc2.Request{ID: "2", Method: "shutdown"}

	// Make the batched LSP request.
	resps, err := cl.RequestBatchAndWaitForAllResponses(
		reqInit,
		reqHover,
		reqShutdown,
	)
	if err != nil {
		return nil, err
	}

	// Unmarshal the LSP responses.
	hoverResp, ok := resps[reqHoverID]
	if !ok {
		return nil, fmt.Errorf("response to hover request from LSP server not found")
	}
	if hoverResp.Error != nil {
		return nil, hoverResp.Error
	}
	var respHover lsp.Hover
	if err := json.Unmarshal(*hoverResp.Result, &respHover); err != nil {
		return nil, err
	}
	return HoverFromLSP(respHover), nil
}

func (t *translator) serveExternalRefs(body []byte) (interface{}, error) {
	// TODO(slimsag): We don't need to create a new JSON RPC 2 connection every
	// time, but we will need reconnection logic and a non-dumb jsonrpc2.Client
	// which can handle concurrency (according to Sourcegraph LSP spec we can
	// and should use one connection for all requests).
	conn, err := net.Dial("tcp", t.Addr)
	if err != nil {
		return nil, err
	}
	cl := jsonrpc2.NewClient(conn)
	defer func() {
		if err := cl.Close(); err != nil {
			// TODO: configurable logging
			log.Println(err)
		}
	}()

	// Decode the user request.
	var r RepoRev
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	if r.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if r.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}

	// Determine the root path for the workspace and prepare it.
	rootPath := filepath.Join(t.WorkDir, r.Repo, r.Commit)
	err = t.prepareWorkspace(rootPath, r.Repo, r.Commit)
	if err != nil {
		return nil, err
	}

	// Build the LSP requests.
	reqInit := &jsonrpc2.Request{
		ID:     "0",
		Method: "initialize",
	}
	if err := reqInit.SetParams(&lsp.InitializeParams{
		RootPath: rootPath,
	}); err != nil {
		return nil, err
	}
	// TODO: should probably check server capabilities before invoking symbol,
	// but good enough for now.
	reqSymbolID := "1"
	reqSymbol := &jsonrpc2.Request{
		ID:     reqSymbolID,
		Method: "workspace/symbol",
	}
	p := lsp.WorkspaceSymbolParams{
		// TODO(keegancsmith) this is go specific
		Query: "external " + r.Repo + "/...",
	}
	if err := reqSymbol.SetParams(p); err != nil {
		return nil, err
	}
	reqShutdown := &jsonrpc2.Request{ID: "2", Method: "shutdown"}

	// Make the batched LSP request.
	resps, err := cl.RequestBatchAndWaitForAllResponses(
		reqInit,
		reqSymbol,
		reqShutdown,
	)
	if err != nil {
		return nil, err
	}

	// Unmarshal the LSP responses.
	symbolResp, ok := resps[reqSymbolID]
	if !ok {
		return nil, fmt.Errorf("response to symbol request from LSP server not found")
	}
	if symbolResp.Error != nil {
		return nil, symbolResp.Error
	}
	var respSymbol []lsp.SymbolInformation
	if err := json.Unmarshal(*symbolResp.Result, &respSymbol); err != nil {
		return nil, err
	}

	var defs []DefSpec
	// TODO(keegancsmith) go specific
	for _, s := range respSymbol {
		// TODO(keegancsmith) we should inspect the workspace to find
		// out the repo of the dependency
		commit := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		pkgParts := strings.Split(s.ContainerName, "/")
		var repo, unit string
		if len(pkgParts) < 3 {
			// Hack for stdlib
			repo = "github.com/golang/go"
			unit = s.ContainerName
		} else {
			repo = strings.Join(pkgParts[:3], "/")
			unit = strings.Join(pkgParts[3:], "/")
		}
		defs = append(defs, DefSpec{
			Repo:     repo,
			Commit:   commit,
			UnitType: "GoPackage",
			Unit:     unit,
			Path:     s.Name,
		})
	}
	return ExternalRefs{Defs: defs}, nil
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
