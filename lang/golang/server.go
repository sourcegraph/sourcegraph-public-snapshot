package golang

import (
	"encoding/json"
	"errors"
	"go/token"
	"log"
	"net"
	"sync"

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

// Session represents an LSP server for one user.
type Session struct {
	mu sync.Mutex

	init         *lsp.InitializeParams // set by "initialize" req
	overlayFiles map[string][]byte     // set by "textDocument/did{Open,Change,Close}" reqs
	fset         *token.FileSet
}

// reset clears all internal state in h.
func (h *Session) reset(init *lsp.InitializeParams) {
	h.init = init
	h.overlayFiles = nil
	h.fset = nil
}

func errResp(req *jsonrpc2.Request, err error) *jsonrpc2.Response {
	if req.Notification {
		log.Println("notification handling failed:", err)
		return nil
	}
	return &jsonrpc2.Response{
		ID:    req.ID,
		Error: &jsonrpc2.Error{Message: err.Error()},
	}
}

func (h *Session) Handle(req *jsonrpc2.Request) *jsonrpc2.Response {
	// Coarse lock (for now) to protect h's internal state.
	h.mu.Lock()
	defer h.mu.Unlock()

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
				TextDocumentSync:   lsp.TDSKFull,
				HoverProvider:      true,
				DefinitionProvider: true,
				ReferencesProvider: true,
			},
		}

	case "textDocument/didOpen":
		var params DidOpenTextDocumentParams
		err = json.Unmarshal(*req.Params, &params)
		if err != nil {
			break
		}
		h.addOverlayFile(params.TextDocument.URI, []byte(params.TextDocument.Text))

	case "textDocument/didChange":
		var params DidChangeTextDocumentParams
		err = json.Unmarshal(*req.Params, &params)
		if err != nil {
			break
		}
		contents, found := h.readOverlayFile(params.TextDocument.URI)
		if !found {
			log.Println("received textDocument/didChange for unknown file:", params.TextDocument.URI)
			break
		}
		for _, change := range params.ContentChanges {
			switch {
			case change.Range == nil && change.RangeLength == 0:
				contents = []byte(change.Text) // new full content

			default:
				log.Println("incremental updates in textDocument/didChange not supported:", params.TextDocument.URI)
			}
		}
		h.addOverlayFile(params.TextDocument.URI, contents)

	case "textDocument/didClose":
		var params DidCloseTextDocumentParams
		err = json.Unmarshal(*req.Params, &params)
		if err != nil {
			break
		}
		h.removeOverlayFile(params.TextDocument.URI)

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
	}

	if err != nil {
		return errResp(req, err)
	}

	if req.Notification {
		return nil
	}

	resp := &jsonrpc2.Response{ID: req.ID}
	resp.SetResult(result)
	return resp
}

// TODO(sqs): move these to package lsp

type DidOpenTextDocumentParams struct {
	TextDocument lsp.TextDocumentItem `json:"textDocument"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   lsp.VersionedTextDocumentIdentifier `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent    `json:"contentChanges"`
}

type TextDocumentContentChangeEvent struct {
	Range       *lsp.Range `json:"range,omitEmpty"`
	RangeLength uint       `json:"rangeLength,omitEmpty"`
	Text        string     `json:"text"`
}

type DidCloseTextDocumentParams struct {
	TextDocument lsp.TextDocumentIdentifier `json:"textDocument"`
}
