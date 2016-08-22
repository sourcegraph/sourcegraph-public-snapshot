package golang

import (
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

// ListenAddrs allows the services/langserver pkg connect to this langserver.
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
			if err := jsonrpc2.Serve(lis, &Handler{}); err != nil {
				log.Fatal("lang/golang: Serve:", err)
			}
		}()
	})
	return listenAddr
}

// Handler implements jsonrpc2.BatchHandler for golang
type Handler struct {
	// These are just used for Handle
	hOnce sync.Once
	h     *Session
}

// Handle is run against the unique session for the Handler. Note this will
// not handle concurrent sessions.
func (h *Handler) Handle(req *jsonrpc2.Request) *jsonrpc2.Response {
	h.hOnce.Do(func() {
		h.h = &Session{}
	})
	return h.h.Handle(req)
}

// HandleBatch spins up and shutsdown a Golang LSP handler, and sends all the
// requests in order.
func (h *Handler) HandleBatch(req []*jsonrpc2.Request) []*jsonrpc2.Response {
	session := &Session{}
	resps := make([]*jsonrpc2.Response, 0, len(req))
	for _, req := range req {
		resp := session.Handle(req)
		if resp != nil {
			resps = append(resps, resp)
		}
	}
	return resps
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

// Session represents an LSP server for one user.
type Session struct {
	mu sync.Mutex

	init *lsp.InitializeParams // set by "initialize" req
}

// reset clears all internal state in h.
func (h *Session) reset(init *lsp.InitializeParams) {
	h.init = init
}

func errResp(req *jsonrpc2.Request, err error) *jsonrpc2.Response {
	if req.Notification {
		log.Println("notification handling failed:", err)
		return nil
	}
	log.Println("error response:", err)
	return &jsonrpc2.Response{
		ID:    req.ID,
		Error: &jsonrpc2.Error{Message: err.Error()},
	}
}

func (h *Session) Handle(req *jsonrpc2.Request) (resp *jsonrpc2.Response) {
	// Coarse lock (for now) to protect h's internal state.
	h.mu.Lock()
	defer h.mu.Unlock()

	// Prevent any uncaught panics from taking the entire server down.
	defer func() {
		if r := recover(); r != nil {
			resp = errResp(req, fmt.Errorf("unexpected panic: %v", r))

			// Same as net/http
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("panic serving %v: %v\n%s", req.Method, r, buf)
			return
		}
	}()

	if req.Method != "initialize" && h.init == nil {
		return errResp(req, errors.New("server must be initialized"))
	}

	var (
		result interface{}
		err    error
	)

	switch req.Method {
	case "initialize":
		var params lsp.InitializeParams
		err = json.Unmarshal(*req.Params, &params)
		if err != nil {
			break
		}
		h.reset(&params)
		result = lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				DefinitionProvider:      true,
				HoverProvider:           true,
				ReferencesProvider:      true,
				WorkspaceSymbolProvider: true,
			},
		}

	case "shutdown":
		// Result is undefined, per
		// https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#shutdown-request.
		result = true

	case "textDocument/hover":
		var params lsp.TextDocumentPositionParams
		err = json.Unmarshal(*req.Params, &params)
		if err != nil {
			break
		}
		result, err = h.handleHover(req, params)

	case "textDocument/definition":
		var params lsp.TextDocumentPositionParams
		err = json.Unmarshal(*req.Params, &params)
		if err != nil {
			break
		}
		result, err = h.handleDefinition(req, params)

	case "textDocument/references":
		var params lsp.ReferenceParams
		err = json.Unmarshal(*req.Params, &params)
		if err != nil {
			break
		}
		result, err = h.handleReferences(req, params)

	case "workspace/symbol":
		var params lsp.WorkspaceSymbolParams
		err = json.Unmarshal(*req.Params, &params)
		if err != nil {
			break
		}
		result, err = h.handleSymbol(req, params)
	}

	if err != nil {
		return errResp(req, err)
	}

	if req.Notification {
		return nil
	}

	resp = &jsonrpc2.Response{ID: req.ID}
	err = resp.SetResult(result)
	if err != nil {
		return errResp(req, err)
	}
	return resp
}
