package golang

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/cmdutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

var listenAddrOnce sync.Once
var listenAddr string

// ListenAddr allows the services/langserver pkg connect to this langserver.
// TODO(keegancsmith) services/langserver is a hack and should not be coupled
// to this package
func ListenAddr() string {
	listenAddrOnce.Do(func() {
		lis, err := net.Listen("tcp", ":0")
		if err != nil {
			log.Fatal("lang/golang: Listen:", err)
		}
		listenAddr = lis.Addr().String()
		log.Println("Go language server listening on", listenAddr)
		go func() {
			h := &Handler{}
			if err := jsonrpc2.Serve(context.Background(), lis, jsonrpc2.HandlerWithError(h.Handle)); err != nil {
				log.Fatal("lang/golang: Serve:", err)
			}
		}()
	})
	return listenAddr
}

// cmdOutput is a helper around c.Output which logs the command, how long it
// took to run, and a nice error in the event of failure.
//
// The specified env is set as cmd.Env (because we do this at ALL callsites
// today anyway).
func cmdOutput(env []string, c *exec.Cmd) ([]byte, error) {
	c.Env = env
	start := time.Now()
	stdout, err := cmdutil.Output(c)
	log.Printf("TIME: %v '%s'\n", time.Since(start), strings.Join(c.Args, " "))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return stdout, nil
}

// Handler represents an LSP handler for one user.
type Handler struct {
	mu sync.Mutex

	init *lsp.InitializeParams // set by "initialize" req
}

// reset clears all internal state in h.
func (h *Handler) reset(init *lsp.InitializeParams) {
	h.init = init
}

func (h *Handler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	// Coarse lock (for now) to protect h's internal state.
	h.mu.Lock()
	defer h.mu.Unlock()

	// Prevent any uncaught panics from taking the entire server down.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unexpected panic: %v", r)

			// Same as net/http
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("panic serving %v: %v\n%s", req.Method, r, buf)
			return
		}
	}()

	if req.Method != "initialize" && h.init == nil {
		return nil, errors.New("server must be initialized")
	}

	switch req.Method {
	case "initialize":
		var params lsp.InitializeParams
		if err = json.Unmarshal(*req.Params, &params); err != nil {
			return
		}
		if params.RootPath == "" {
			params.RootPath = "/"
		}
		h.reset(&params)
		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				DefinitionProvider:      true,
				HoverProvider:           true,
				ReferencesProvider:      true,
				WorkspaceSymbolProvider: true,
			},
		}, nil

	case "shutdown":
		// Result is undefined, per
		// https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#shutdown-request.
		return nil, nil

	case "textDocument/hover":
		var params lsp.TextDocumentPositionParams
		if err = json.Unmarshal(*req.Params, &params); err != nil {
			return
		}
		return h.handleHover(req, params)

	case "textDocument/definition":
		var params lsp.TextDocumentPositionParams
		if err = json.Unmarshal(*req.Params, &params); err != nil {
			return
		}
		return h.handleDefinition(req, params)

	case "textDocument/references":
		var params lsp.ReferenceParams
		if err = json.Unmarshal(*req.Params, &params); err != nil {
			break
		}
		return h.handleReferences(req, params)

	case "workspace/symbol":
		var params lsp.WorkspaceSymbolParams
		if err = json.Unmarshal(*req.Params, &params); err != nil {
			break
		}
		return h.handleSymbol(req, params)
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}
