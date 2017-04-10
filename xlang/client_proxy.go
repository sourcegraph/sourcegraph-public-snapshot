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
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/juju/ratelimit"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	plspext "github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

// repoBlacklist contains repos which we have blacklisted. It is set via the
// environment variable REPO_BLACKLIST.
var repoBlacklist = make(map[string]bool)

// repoBlacklistXReferences contains repos which we have blacklisted only on
// workspace/xreferences. It is set via the environment variable REPO_BLACKLIST_XREFERENCES.
var repoBlacklistXReferences = make(map[string]bool)

var (
	clientLimitReqSec      float64
	clientLimitReqSecBurst int64
)

func init() {
	repos := strings.Fields(env.Get("REPO_BLACKLIST", "", "repos which we should not serve requests for. Seperated by whitespace"))
	for _, r := range repos {
		repoBlacklist[r] = true
	}

	repos = strings.Fields(env.Get("REPO_BLACKLIST_XREFERENCES", "", "repos which we should not serve workspace/xreferences requests for. Seperated by whitespace"))
	for _, r := range repos {
		repoBlacklistXReferences[r] = true
	}

	var err error
	clientLimitReqSec, err = strconv.ParseFloat(env.Get("CLIENT_LIMIT_REQ_SEC", "2", "The allowed requests a second before rate limiting. float"), 64)
	if err != nil {
		log.Fatal("badly formatted CLIENT_LIMIT_REQ_SEC", err)
	}
	clientLimitReqSecBurst, err = strconv.ParseInt(env.Get("CLIENT_LIMIT_REQ_SEC_BURST", "50", "The maximum requests a second before rate limiting. int"), 10, 64)
	if err != nil {
		log.Fatal("badly formatted CLIENT_LIMIT_REQ_SEC_BURST", err)
	}
}

func (p *Proxy) newClientProxyConn(ctx context.Context, rwc io.ReadWriteCloser) error {
	var connOpt []jsonrpc2.ConnOpt
	if p.Trace {
		connOpt = append(connOpt, jsonrpc2.LogMessages(log.New(os.Stderr, "", 0)))
	}

	c := &clientProxyConn{
		proxy: p,
		last:  time.Now(),
		id:    nextClientID(),

		limiter: ratelimit.NewBucketWithRate(clientLimitReqSec, clientLimitReqSecBurst),
	}
	c.conn = jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(rwc, jsonrpc2.VSCodeObjectCodec{}), jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError(c.handle)), connOpt...)

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
	clientRateLimited = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "xlang",
		Name:      "client_rate_limited",
		Help:      "The number of times a client request was rate limited.",
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
	prometheus.MustRegister(clientConnsGauge)
	prometheus.MustRegister(clientConnsCounter)
	prometheus.MustRegister(clientRateLimited)
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

	// session is the unique ID identifying this session, used when it
	// shouldn't be shared by all users viewing the same rootPath and
	// mode (e.g., for Zap and/or when textDocument/didChange, etc.,
	// should be enabled).
	//
	// 🚨 SECURITY: The session isolation that this provides is dependent 🚨
	// on how difficult to guess this value is. Currently it is chosen
	// by the client (and only used for Zap). E.g., if the client
	// picks "foo", then anyone else could probably guess "foo". If
	// the client guesses a long unique string, then nobody will be
	// able to guess it. When we expose isolated session functionality
	// to users, we should guarantee that session is always chosen
	// externally so that sufficiently unguessable values are used.
	session string
}

func (id contextID) String() string {
	return fmt.Sprintf("context(%s mode=%s session=%q)", id.rootPath.String(), id.mode, id.session)
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

	// limiter is used to rate limit a client.
	limiter *ratelimit.Bucket

	mu       sync.Mutex
	context  contextID
	init     *lspext.ClientProxyInitializeParams
	last     time.Time // max(last request received, last response sent), used to evict idle clients
	shutdown bool      // whether this connection has received an LSP "shutdown"
}

// LogTrackedErrors if true causes errors to be logged if they are related to
// language analysis.
var LogTrackedErrors = true

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

	// Enforce rate limiter only for requests. We can send a large amount
	// of notifications in normal operation.
	if !req.Notif && c.limiter.TakeAvailable(1) != 1 {
		clientRateLimited.Inc()
		// This client is misbehaving rate limit wise, so we fail the request.
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidRequest,
			Message: "rate limit for client exceeded",
		}
	}

	c.mu.Lock()
	shutdown := c.shutdown
	c.mu.Unlock()
	if shutdown && !(req.Method == "exit" || req.Method == "shutdown") {
		// vscode may retry a shutdown request. So we ignore it sending shutdown twice.
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
			lspext.ClientProxyInitializeParams
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

		// 🚨 SECURITY: LSP introduced a rootUri field on 🚨
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
		if repoBlacklist[rootPathURI.Repo()] {
			return nil, fmt.Errorf("repo is blacklisted")
		}

		c.mu.Lock()
		if c.init != nil {
			c.mu.Unlock()
			// This would only happen if the client is misbehaving (if
			// it sends 2 "initialize" requests).
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidRequest, Message: fmt.Sprintf("client proxy handler is already initialized")}
		}
		c.init = &params.ClientProxyInitializeParams
		c.context.rootPath = *rootPathURI
		c.context.mode = c.init.InitializationOptions.Mode
		c.context.session = c.init.InitializationOptions.Session
		c.mu.Unlock()

		// PERF: Initialize the server as soon as possible, so that it
		// can start indexing before a user does an explicit request.
		err = c.proxy.initializeServer(ctx, serverID{contextID: c.context})
		if err != nil {
			return nil, err
		}

		kind := lsp.TDSKFull
		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				TextDocumentSync:             lsp.TextDocumentSyncOptionsOrKind{Kind: &kind},
				ReferencesProvider:           true,
				DefinitionProvider:           true,
				HoverProvider:                true,
				DocumentSymbolProvider:       true,
				WorkspaceSymbolProvider:      true,
				XWorkspaceReferencesProvider: true,
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

		if req.Method == "workspace/xreferences" && repoBlacklistXReferences[c.context.rootPath.Repo()] {
			return nil, fmt.Errorf("repo is blacklisted")
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
						logError("Shutting down background server failed: "+err.Error(), c.context, "method", req.Method, "id", req.ID)
					}
				}()
			}()
		}

		var respObj interface{}
		if err := c.callServer(ctx, req.ID, req.Method, req.Notif, false, req.Params, &respObj); err != nil {
			logError(req.Method+" failed: "+err.Error(), c.context, "method", req.Method, "id", req.ID, "error", err.Error())
			return nil, err
		}
		return respObj, nil

	case "textDocument/didOpen":
		if err := ensureInitialized(); err != nil {
			return nil, err
		}

		if proxySaveDiagnostics {
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

		// Only allow isolated sessions (not shared sessions) to
		// modify file contents. The modified file contents will
		// only be visible within this session and will not leak
		// to other sessions.
		if c.allowTextDocumentSync() {
			if err := c.callServer(ctx, req.ID, req.Method, req.Notif, false, req.Params, nil); err != nil {
				return nil, err
			}
			return nil, nil
		}

		// didOpen is sent when a user opens a file. However, didOpen
		// can mutate the workspace of our shared language server. Our
		// only clients should never try to mutate the workspace. So
		// we ignore the body of didOpen and send a fake hover
		// request. We do this since most language servers will cache
		// type information for a file after a request for it. Some
		// language servers are smart about immutable didOpen, but
		// others may trigger overlay code / cache invalidation in the
		// language server.
		var params lsp.DidOpenTextDocumentParams
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		err := json.Unmarshal(*req.Params, &params)
		if err != nil {
			return nil, err
		}
		var respObj interface{}
		hoverParams := lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: params.TextDocument.URI},
			Position: lsp.Position{
				Line:      0,
				Character: 0,
			},
		}
		err = c.callServer(ctx, req.ID, "textDocument/hover", false, true, &hoverParams, &respObj)
		return nil, err // we ignore the hover response (other than err) since the original request was a notif.

	case "textDocument/didClose":
		if err := ensureInitialized(); err != nil {
			return nil, err
		}

		// didClose is harmless to pass along, it does not mutate
		// anything.
		//
		// TODO(john): I noticed while debugging https://github.com/sourcegraph/sourcegraph/issues/4616
		// that this statement is not true. Toggling the diff editor sends a didClose event for the
		// "left" document URI (like this https://cl.ly/1y3M3r1j0W1i) which does seem to have some reverting
		// effect on the langserver.
		// if err := c.callServer(ctx, req.ID, req.Method, req.Notif, false, req.Params, nil); err != nil {
		// 	return nil, err
		// }
		return nil, nil

	case "textDocument/didChange", "textDocument/didSave":
		if err := ensureInitialized(); err != nil {
			return nil, err
		}

		// Only allow isolated sessions (not shared sessions) to
		// modify file contents. The modified file contents will
		// only be visible within this session and will not leak
		// to other sessions.
		if c.allowTextDocumentSync() {
			if err := c.callServer(ctx, req.ID, req.Method, req.Notif, false, req.Params, nil); err != nil {
				return nil, err
			}
			return nil, nil
		}

		// Specifically forbid these methods so we don't accidentally
		// allow them through. If we did, then any user of a shared
		// workspace could modify the files used for analysis for all
		// users.
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidRequest, Message: fmt.Sprintf("client proxy handler: text document modifications not allowed by client (%s)", req.Method)}

	case "$/cancelRequest":
		// We should forward this to all language servers, but this requires writing some logic to rewrite the ids.
		// Since our language servers don't currently support cancellation it is fine to just not forward it for now.
		return nil, nil

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

// allowTextDocumentSync returns true if the client is allowed to send LSP
// requests which can mutate the workspace. We allow mutable workspaces when
// it isn't shared amongst users (such as when a user is editing it via zap).
func (c *clientProxyConn) allowTextDocumentSync() bool {
	return c.context.session != ""
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

var requestOriginatedFromProxy = &jsonrpc2.ID{Str: "special value; never used directly"}

// callServer sends the LSP request to the server chosen based on the
// client's context and the file URI specified (e.g., for a ".go"
// file, it will choose a Go lang/build server). It rewrites any file
// URIs to refer to file paths in the virtual workspace, not the
// repository clone URL.
//
// If notif is true, then rid and result must be zero valued.
//
// If requestOriginatedFromProxy is true, then it indicates that the
// request does not correspond to a proxied request from the client
// (e.g., the "textDocument/hover" prewarm request we send). In this
// case, the proxy MUST generate an ID for the request that is
// guaranteed to be distinct from any request ID generated by
// jsonrpc2.PickID, so that the proxy's request does not get treated
// as a client request (which is possible if the jsonrpc2.ID zero
// value, indicating a numeric ID of 0, is used).
func (c *clientProxyConn) callServer(ctx context.Context, rid jsonrpc2.ID, method string, notif, requestOriginatedFromProxy bool, params, result interface{}) error {
	if notif && (rid != jsonrpc2.ID{} || result != nil) {
		return fmt.Errorf("invalid non-zero ID or result for notification %q", method)
	}

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
	// HOTFIX(keegancsmith) remove retries 2017-02-21
	backoffs := []time.Duration{0} //, 100 * time.Millisecond, 1 * time.Second}
	for _, b := range backoffs {
		if b != 0 {
			// We are retrying. Add in jitter to our backoff sleep
			time.Sleep(b + time.Duration(rand.Int63n(50)-25)*time.Millisecond)
			proxyRetryCounter.WithLabelValues(id.mode).Inc()
		}
		err = c.proxy.callServer(ctx, crid, id, method, notif, requestOriginatedFromProxy, params, result)
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
	if result != nil {
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
