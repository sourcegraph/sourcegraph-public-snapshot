package xlang

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	plspext "github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

func (p *Proxy) newClientProxyConn(ctx context.Context, rwc io.ReadWriteCloser) error {
	var connOpt []jsonrpc2.ConnOpt
	if p.Trace {
		connOpt = append(connOpt, jsonrpc2.LogMessages(log.New(os.Stderr, "", 0)))
	}

	c := &clientProxyConn{
		proxy: p,
		last:  time.Now(),
		id:    nextClientID(),
	}
	c.conn = jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(rwc, jsonrpc2.VSCodeObjectCodec{}), jsonrpc2.HandlerWithError(c.handle), connOpt...)

	p.mu.Lock()
	if p.clients == nil {
		p.mu.Unlock()
		return errors.New("the proxy has been closed")
	}
	p.clients[c] = struct{}{}
	clientConnsGauge.Set(float64(len(p.clients)))
	clientConnsCounter.Inc()
	p.mu.Unlock()
	go func() {
		select {
		case <-c.conn.DisconnectNotify():
		}
		p.removeClientConn(c)
	}()
	return nil
}

var (
	badRevCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "xlang",
		Name:      "client_proxy_bad_rev",
		Help:      "Temporary counter to confirm clients are parsing well-formed revs.",
	})
	clientConnsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "xlang",
		Name:      "open_client_proxy_connections",
		Help:      "Number of open connections to the xlang client proxy.",
	})
	clientConnsCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "xlang",
		Name:      "cumu_client_proxy_connections",
		Help:      "Cumulative number of connections to the xlang client proxy (total of open + previously closed since process startup).",
	})
	proxyRetryCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "xlang",
		Name:      "proxy_retries",
		Help:      "The number of times a client retried a request to a server.",
	}, []string{"mode"})
	proxyRetryFailedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "xlang",
		Name:      "proxy_retry_failed",
		Help:      "Count of how often our transient error retries fails to get a result.",
	}, []string{"mode"})
)

func init() {
	prometheus.MustRegister(badRevCounter)
	prometheus.MustRegister(clientConnsGauge)
	prometheus.MustRegister(clientConnsCounter)
	prometheus.MustRegister(proxyRetryCounter)
	prometheus.MustRegister(proxyRetryFailedCounter)
}

func (p *Proxy) removeClientConn(c *clientProxyConn) {
	p.mu.Lock()
	delete(p.clients, c)
	clientConnsGauge.Set(float64(len(p.clients)))
	p.mu.Unlock()
}

// DisconnectIdleClients shuts down clients whose last communication
// with the proxy (either a request or response) was longer than
// maxIdle ago. The Proxy runs DisconnectIdleClients periodically
// based on p.MaxClientIdle.
func (p *Proxy) DisconnectIdleClients(maxIdle time.Duration) error {
	cutoff := time.Now().Add(-1 * maxIdle)
	errs := &errorList{}
	var wg sync.WaitGroup
	p.mu.Lock()
	for c := range p.clients {
		c.mu.Lock()
		idle := c.last.Before(cutoff)
		c.mu.Unlock()
		if idle {
			wg.Add(1)
			go func(c *clientProxyConn) {
				defer wg.Done()
				p.removeClientConn(c)
				if err := c.conn.Close(); err != nil {
					errs.add(err)
				}
			}(c)
		}
	}
	// Only hold lock during fast loop iter, not while waiting to
	// close each idle connection (otherwise we could block p.mu for a
	// long time if closing blocks).
	p.mu.Unlock()

	wg.Wait()
	return errs.error()
}

// contextID identifies a client's session by the minimal information
// necessary to reinitialize it. Two client connections can have
// identical contextInfo, in which case they will share lang/build
// servers. This happens frequently, e.g. in the case when two
// anonymous clients are accessing the same repository at the same
// commit.
type contextID struct {
	rootPath uri.URI // the rootPath in the initialize request (typically the repo clone URL + "?REV")
	mode     string  // the mode (i.e., "go" or "typescript")
}

func (id contextID) String() string {
	return fmt.Sprintf("context(%s mode=%s)", id.rootPath.String(), id.mode)
}

// clientID is used to uniquely identify a client connection in this
// process. The main use is to tie-back jsonrpc2 responses from
// servers with the client it was proxied on behalf of.
type clientID uint64

// clientIDSeq is used to generate clientIDs. Do not use this directly,
// instead use nextClientID.
var clientIDSeq uint64

// nextClientID returns a new clientID which is unique to this process.
func nextClientID() clientID {
	return clientID(atomic.AddUint64(&clientIDSeq, 1))
}

// clientRequestID helps tie back a jsonrpc2 request id to a server_proxy back
// to a client request.
type clientRequestID struct {
	// RID is ID of the original client request
	RID jsonrpc2.ID
	// CID is the client we are proxying on behalf of
	CID clientID
}

type clientProxyConn struct {
	proxy *Proxy         // the proxy that accepted this conn
	conn  *jsonrpc2.Conn // the LSP JSON-RPC 2.0 connection to the client
	id    clientID       // unique id for this connection

	mu       sync.Mutex
	context  contextID
	init     *ClientProxyInitializeParams
	last     time.Time // max(last request received, last response sent), used to evict idle clients
	shutdown bool      // whether this connection has received an LSP "shutdown"
}

// ClientProxyInitializeParams are sent by the client to the proxy in
// the "initialize" request. It has a non-standard field "mode", which
// is the name of the language (using vscode terminology); "go" or
// "typescript", for example.
type ClientProxyInitializeParams struct {
	lsp.InitializeParams
	InitializationOptions ClientProxyInitializationOptions `json:"initializationOptions"`

	// Mode is DEPRECATED; it was moved to the subfield
	// initializationOptions.Mode. It is still here for backward
	// compatibility until the xlang service is upgraded.
	Mode string `json:"mode,omitempty"`
}

// ClientProxyInitializationOptions is the "initializationOptions"
// field of the "initialize" request params sent from the client to
// the LSP client proxy.
type ClientProxyInitializationOptions struct {
	Mode string `json:"mode"`
}

// LogTrackedErrors, if true, causes errors to be logged if they are
// related to language analysis.
var LogTrackedErrors = true

type trackedError struct {
	RootPath string
	Mode     string
	Method   string
	Params   interface{}
	Error    string
}

// handleFromClient receives requests from the client, modifies them,
// sends them to the appropriate lang/build server(s), modifies the
// responses, and returns them to the client.
//
// It modifies the request to rewrite paths (such as initialize's
// rootPath and textDocument/definition's textDocument.uri fields) to
// point to file system paths, checking out the repo to that file
// system path if necessary.
//
// Certain operations (such as workspace/symbols) must be called on
// all build/lang servers, in which case the results are merged
// transparently to the client.
func (c *clientProxyConn) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	c.updateLastTime()
	defer c.updateLastTime()

	// Try to get our parent span context from the JSON-RPC request
	// from the LSP client.
	opName := "LSP client proxy: " + req.Method
	var span opentracing.Span
	var carrier opentracing.TextMapCarrier
	if req.Meta != nil {
		if err := json.Unmarshal(*req.Meta, &carrier); err != nil {
			return nil, err
		}
	}
	if clientCtx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, carrier); err == nil {
		span = opentracing.StartSpan(opName, ext.RPCServerOption(clientCtx))
		ctx = opentracing.ContextWithSpan(ctx, span)
	} else if err != opentracing.ErrSpanContextNotFound {
		return nil, err
	} else {
		span, ctx = opentracing.StartSpanFromContext(ctx, opName)
	}
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogEvent(fmt.Sprintf("error: %v", err))
		}
		span.Finish()
	}()

	c.mu.Lock()
	shutdown := c.shutdown
	c.mu.Unlock()
	if shutdown && req.Method != "exit" {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidRequest,
			Message: fmt.Sprintf("invalid LSP request %q received while client proxy is shutting down (only \"exit\" is allowed)", req.Method),
		}
	}

	// ensureInitialized should be used below methods that require the
	// client to have already sent an "initialize" request.
	ensureInitialized := func() error {
		c.mu.Lock()
		initialized := c.init != nil
		c.mu.Unlock()
		if !initialized {
			return &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInvalidRequest,
				Message: fmt.Sprintf("LSP client must send \"initialize\" request before sending %q", req.Method),
			}
		}
		return nil
	}

	switch req.Method {
	case "initialize":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params struct {
			ClientProxyInitializeParams
			RootURI string `json:"rootUri"`
		}
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		// DEPRECATED: Handle clients that send initialization params with the old Mode field.
		if params.InitializationOptions.Mode == "" {
			params.InitializationOptions.Mode = params.Mode
			params.Mode = ""
		}

		// SECURITY: LSP introduced a rootUri field on
		// InitializeParams and deprecated rootPath. Until we
		// implement rootUri, ensure that it is not being sent to us,
		// to avoid confusion. This also helps avoid an exploit where
		// we check access to the rootPath but provide access to the
		// rootUri, which would allow an attacker to bypass our access
		// check.
		if params.RootURI != "" {
			return nil, fmt.Errorf("rootUri field is not yet supported (use rootPath only): got value %q", params.RootURI)
		}

		rootPathURI, err := uri.Parse(params.RootPath)
		if err != nil {
			return nil, fmt.Errorf("invalid rootPath: %s", err)
		}
		if params.InitializationOptions.Mode == "" {
			return nil, fmt.Errorf(`client must send a "mode" in the initialize request to specify the language`)
		}
		if len(rootPathURI.Rev()) != 40 {
			return nil, fmt.Errorf("absolute commit ID required (40 hex chars) in rootPath %q", rootPathURI)
		}
		if len(rootPathURI.Rev()) != 40 {
			badRevCounter.Inc()
			log.Printf("non absolute rev for %s %s", params.RootPath, traceutil.SpanURL(span))
		}

		c.mu.Lock()
		defer c.mu.Unlock()
		if c.init != nil {
			// This would only happen if the client is misbehaving (if
			// it sends 2 "initialize" requests).
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidRequest, Message: fmt.Sprintf("client proxy handler is already initialized")}
		}
		c.init = &params.ClientProxyInitializeParams
		c.context.rootPath = *rootPathURI
		c.context.mode = c.init.InitializationOptions.Mode
		kind := lsp.TDSKFull
		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				TextDocumentSync:   lsp.TextDocumentSyncOptionsOrKind{Kind: &kind},
				ReferencesProvider: true,
				DefinitionProvider: true,
				HoverProvider:      true,
			},
		}, nil

	case "initialized":
		if err := ensureInitialized(); err != nil {
			return nil, err
		}

		// This is a valid notification to send, but given in our
		// situation we can have many clients for a single language
		// server, it doesn't make sense to pass on.
		return nil, nil

	case "textDocument/definition", "textDocument/xdefinition", "textDocument/hover", "textDocument/references", "textDocument/documentSymbol", "workspace/symbol", "workspace/xreferences", "workspace/xdependencies", "workspace/xpackages":
		if err := ensureInitialized(); err != nil {
			return nil, err
		}
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		// Background modes only ever do one request against them
		// (currently workspace/xreferences). As such we do not need to
		// keep the workspace open.
		if strings.HasSuffix(c.context.mode, "_bg") {
			defer func() {
				go func() {
					id := serverID{contextID: c.context, pathPrefix: ""}
					err := c.proxy.shutDownServer(context.Background(), id)
					if err != nil {
						log.Printf("error shutting down background server: %s", err)
					}
				}()
			}()
		}

		var respObj interface{}
		if err := c.callServer(ctx, req.ID, req.Method, req.Params, &respObj); err != nil {
			// Machine parseable to assist us finding most common errors
			msg, _ := json.Marshal(trackedError{
				RootPath: c.context.rootPath.String(),
				Mode:     c.context.mode,
				Method:   req.Method,
				Params:   req.Params,
				Error:    err.Error(),
			})
			if LogTrackedErrors {
				log.Printf("tracked error: %s", string(msg))
			}
			return nil, err
		}
		return respObj, nil

	case "textDocument/didOpen", "textDocument/didChange", "textDocument/didClose", "textDocument/didSave":
		if err := ensureInitialized(); err != nil {
			return nil, err
		}

		if proxySaveDiagnostics && req.Method == "textDocument/didOpen" {
			var params lsp.DidOpenTextDocumentParams
			if req.Params == nil {
				return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
			}
			if err := json.Unmarshal(*req.Params, &params); err != nil {
				return nil, err
			}

			// When the client opens a document, send over any diagnostics
			// we received in the past.
			relURI, err := relWorkspaceURI(c.context.rootPath, params.TextDocument.URI)
			if err != nil {
				return nil, err
			}
			if diags := c.proxy.getSavedDiagnostics(serverID{contextID: c.context, pathPrefix: ""}, relURI.String()); diags != nil {
				diagnosticsParams := lsp.PublishDiagnosticsParams{
					URI:         params.TextDocument.URI,
					Diagnostics: diags,
				}
				if err := conn.Notify(ctx, "textDocument/publishDiagnostics", diagnosticsParams); err != nil {
					log15.Error("LSP client proxy: error sending saved diagnostics", "context", c.context, "uri", params.TextDocument.URI, "err", err)
				}
			}
		}

		// Specifically forbid these methods so we don't accidentally
		// allow them through. If we did, then any user of a shared
		// workspace could modify the files used for analysis for all
		// users.
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidRequest, Message: fmt.Sprintf("client proxy handler: text document modifications not allowed by client (%s)", req.Method)}

	case "shutdown":
		c.mu.Lock()
		c.shutdown = true
		c.mu.Unlock()
		return nil, nil

	case "exit":
		c.mu.Lock()
		c.shutdown = true
		c.mu.Unlock()
		c.proxy.removeClientConn(c)
		if err := c.conn.Close(); err != jsonrpc2.ErrClosed { // ignore if already closed
			return nil, err
		}
		return nil, nil

	default:
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("client proxy handler: method not found: %q", req.Method)}
	}
}

// handleFromServer is called by associated server proxy connections
// when they receive requests that should be propagated to the client.
func (c *clientProxyConn) handleFromServer(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	c.updateLastTime()
	defer c.updateLastTime()

	c.mu.Lock()
	shutdown := c.shutdown
	c.mu.Unlock()
	if shutdown {
		return nil, nil
	}

	switch req.Method {
	case "textDocument/publishDiagnostics":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var paramsObj lsp.PublishDiagnosticsParams
		if err := json.Unmarshal(*req.Params, &paramsObj); err != nil {
			return nil, err
		}

		// Rewrite paths from server->client and send rewritten
		// notification to client.
		var walkErr error
		lspext.WalkURIFields(&paramsObj, nil, func(uriStr string) string {
			newURI, err := absWorkspaceURI(c.context.rootPath, uriStr)
			if err != nil {
				walkErr = err
				return ""
			}
			return newURI.String()
		})
		if walkErr != nil {
			return nil, walkErr
		}
		if err := conn.Notify(ctx, req.Method, paramsObj); err != nil {
			if err == jsonrpc2.ErrClosed || strings.Contains(err.Error(), "use of closed network connection") {
				err = nil // suppress worthless "notification handling error" log messages when the client has hung up
			}
			return nil, err
		}
		return nil, nil

	case "$/partialResult":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		// Initially just unmarshal the ID, since we may return
		// early. This helps us avoid unmarshalling a potentially
		// large []lsp.Location.
		idOnly := struct {
			ID jsonrpc2.ID `json:"id"`
		}{}
		if err := json.Unmarshal(*req.Params, &idOnly); err != nil {
			return nil, err
		}
		srid := serverRequestIDFromString(idOnly.ID.Str)
		if srid.CRID.CID != c.id {
			// This partialResult's clientID does not match our
			// clientID. This partialResult is not for us so we
			// ignore it.
			return nil, nil
		}

		var paramsObj plspext.PartialResultParams
		if err := json.Unmarshal(*req.Params, &paramsObj); err != nil {
			return nil, err
		}

		// Rewrite ID so it is the same as the originating request
		paramsObj.ID = lsp.ID{
			Str:      srid.CRID.RID.Str,
			Num:      srid.CRID.RID.Num,
			IsString: srid.CRID.RID.IsString,
		}

		// Rewrite paths from server->client and send rewritten
		// notification to client.
		var walkErr error
		lspext.WalkURIFields(paramsObj.Patch, nil, func(uriStr string) string {
			newURI, err := absWorkspaceURI(c.context.rootPath, uriStr)
			if err != nil {
				walkErr = err
				return ""
			}
			return newURI.String()
		})
		if walkErr != nil {
			return nil, walkErr
		}
		if err := conn.Notify(ctx, req.Method, paramsObj); err != nil {
			if err == jsonrpc2.ErrClosed || strings.Contains(err.Error(), "use of closed network connection") {
				err = nil // suppress worthless "notification handling error" log messages when the client has hung up
			}
			return nil, err
		}
		return nil, nil

	default:
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("client handler for propagating server messages: method not found: %q", req.Method)}
	}
}

// callServer sends the LSP request to the server chosen based on the
// client's context and the file URI specified (e.g., for a ".go"
// file, it will choose a Go lang/build server). It rewrites any file
// URIs to refer to file paths in the virtual workspace, not the
// repository clone URL.
func (c *clientProxyConn) callServer(ctx context.Context, rid jsonrpc2.ID, method string, params, result interface{}) error {
	pb, err := json.Marshal(params)
	if err != nil {
		return err
	}
	params = nil
	if err := json.Unmarshal(pb, &params); err != nil {
		return err
	}
	var uris []string
	lspext.WalkURIFields(params, func(uri string) {
		uris = append(uris, uri)
	}, nil)
	if len(uris) != 1 && method != "workspace/symbol" && method != "workspace/xreferences" && method != "workspace/xdependencies" && method != "workspace/xpackages" {
		return fmt.Errorf("expected exactly 1 document URI (got %d) in LSP params object %s", len(uris), pb)
	}

	// Now that we know the prefix of the workspace, rewrite the paths
	// in the LSP params object.
	var walkErr error
	lspext.WalkURIFields(params, nil, func(uriStr string) string {
		newURI, err := relWorkspaceURI(c.context.rootPath, uriStr)
		if err != nil {
			walkErr = err
			return ""
		}
		return newURI.String()
	})
	if walkErr != nil {
		return walkErr
	}

	id := serverID{contextID: c.context, pathPrefix: ""}
	crid := clientRequestID{RID: rid, CID: c.id}

	// We try upto 3 times if we encounter ephemeral errors
	backoffs := []time.Duration{0, 100 * time.Millisecond, 1 * time.Second}
	for _, b := range backoffs {
		if b != 0 {
			// We are retrying. Add in jitter to our backoff sleep
			time.Sleep(b + time.Duration(rand.Int63n(50)-25)*time.Millisecond)
			proxyRetryCounter.WithLabelValues(id.mode).Inc()
		}
		err = c.proxy.callServer(ctx, crid, id, method, params, result)
		if err == nil {
			break
		}
		if !isTemporary(err) && err != io.EOF && err != io.ErrUnexpectedEOF {
			return err
		}
	}
	if err != nil {
		proxyRetryFailedCounter.WithLabelValues(id.mode).Inc()
		return err
	}

	// Convert the URIs back.
	result2, err := json.Marshal(result)
	if err != nil {
		return err
	}
	var resultObj interface{}
	if err := json.Unmarshal(result2, &resultObj); err != nil {
		return err
	}
	lspext.WalkURIFields(resultObj, nil, func(uriStr string) string {
		newURI, err := absWorkspaceURI(c.context.rootPath, uriStr)
		if err != nil {
			walkErr = err
			return ""
		}
		return newURI.String()
	})
	if walkErr != nil {
		return walkErr
	}
	result2, err = json.Marshal(resultObj)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(result2, result); err != nil {
		return err
	}
	return nil
}

func (c *clientProxyConn) updateLastTime() {
	c.mu.Lock()
	c.last = time.Now()
	c.mu.Unlock()
}

func isTemporary(err error) bool {
	te, ok := err.(interface {
		Temporary() bool
	})
	return ok && te.Temporary()
}
