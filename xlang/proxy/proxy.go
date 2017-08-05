package proxy

import (
	"context"
	"log"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/neelance/parallel"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

const (
	// CodeModeNotFound is the JSON-RPC 2.0 error code that indicates that the client
	// requested to initialize a session for a mode (language ID, such as "go") that no
	// servers are registered to handle.
	CodeModeNotFound = -32000
)

// New creates a new LSP proxy.
func New() *Proxy {
	return &Proxy{
		MaxClientIdle:   30 * time.Minute,
		MaxServerIdle:   300 * time.Second,
		MaxServerUnused: 30 * time.Second,

		closed: make(chan struct{}),

		clients: map[*clientProxyConn]struct{}{},
		servers: map[*serverProxyConn]struct{}{},
	}
}

// Proxy proxies LSP JSON-RPC 2.0 connections, sitting between the
// client (typically a user's browser, via our HTTP API) and
// lang/build servers.
type Proxy struct {
	MaxClientIdle   time.Duration // disconnect idle clients after this duration
	MaxServerIdle   time.Duration // shut down idle servers after this duration
	MaxServerUnused time.Duration // shut down unused servers after this duration

	Trace bool // print traces of all requests/responses between proxy and client

	closed chan struct{} // a channel that is closed when (*Proxy).Close is called

	mu      sync.Mutex
	clients map[*clientProxyConn]struct{} // open connections from clients
	servers map[*serverProxyConn]struct{} // open connections to lang/build servers
}

// Serve accepts incoming client connections on the listener l.
//
// The client should send an LSP initialize request immediately after
// connecting.
//
// Serve always returns a non-nil error.
func (p *Proxy) Serve(ctx context.Context, lis net.Listener) error {
	// Run background goroutines to disconnect idle clients and
	// terminate idle servers.
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case <-done:
				return // stop when the listener is closed
			case <-time.After(p.MaxClientIdle / 2):
				if err := p.DisconnectIdleClients(p.MaxClientIdle); err != nil {
					log.Printf("LSP proxy: disconnecting idle clients: %s", err)
				}
			}
		}
	}()
	go func() {
		for {
			d := p.MaxServerIdle
			if p.MaxServerUnused < d {
				d = p.MaxServerUnused
			}
			select {
			case <-done:
				return // stop when the listener is closed
			case <-time.After(d / 2):
				ctx, cancel := context.WithTimeout(context.Background(), d)
				idleCutoff := time.Now().Add(-1 * p.MaxServerIdle)
				unusedCutoff := time.Now().Add(-1 * p.MaxServerUnused)
				filter := func(s *serverProxyConn) bool {
					s.mu.Lock()
					last := s.stats.Last
					unused := s.stats.TotalCount == 0
					// If the only request has been for workspace/xreference,
					// expire now. If workspace/xreferences is present it is
					// usually the only request done to a server.
					isShortLived := s.stats.TotalCount == 1 && s.stats.TotalFinishedCount == 1 && s.stats.Counts["workspace/xreferences"] == 1
					s.mu.Unlock()
					return last.Before(idleCutoff) || isShortLived || (unused && last.Before(unusedCutoff))
				}
				if err := p.shutdownServers(ctx, filter); err != nil {
					log.Printf("LSP proxy: shutting down idle servers: %s", err)
				}
				cancel()
			}
		}
	}()

	for {
		nc, err := lis.Accept()
		if err != nil {
			return err
		}
		err = p.newClientProxyConn(ctx, nc)
		if err != nil {
			return err
		}
	}
}

// Close shuts down all build/language servers and closes all client
// and server connections. It does NOT stop any listeners passed to
// Serve; those must be closed prior to calling Close.
//
// TODO(sqs): consider returning from Serve or printing a log message
// if this Close is called but there are still active listeners.
func (p *Proxy) Close(ctx context.Context) error {
	par := parallel.NewRun(runtime.GOMAXPROCS(0))
	p.mu.Lock()
	for c := range p.clients {
		par.Acquire()
		go func(c *clientProxyConn) {
			defer par.Release()
			if err := c.conn.Close(); err != nil {
				par.Error(err)
			}
		}(c)
	}
	for s := range p.servers {
		par.Acquire()
		go func(s *serverProxyConn) {
			defer par.Release()
			if err := s.shutdownAndExit(ctx); err != nil {
				par.Error(err)
			}
			if err := s.conn.Close(); err != nil {
				par.Error(err)
			}
		}(s)
	}

	// Set to nil so that calls to DisconnectIdleClients and
	// shutdownIdleServers that are blocked on p.mu (which we hold) do
	// not attempt to double-close any client/server conns (thereby
	// causing a panic).
	p.clients = nil
	p.servers = nil

	// Only hold lock during fast loop iter; no need to wait for the
	// shutdowns/disconnects to complete.
	p.mu.Unlock()

	return par.Wait()
}

func (p *Proxy) getSavedDiagnostics(id serverID, documentURI lsp.DocumentURI) []lsp.Diagnostic {
	var c *serverProxyConn
	p.mu.Lock()
	for cc := range p.servers {
		if cc.id == id {
			c = cc
			break
		}
	}
	p.mu.Unlock()
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.diagnostics[diagnosticsKey{serverID: c.id, documentURI: documentURI}]
}

// getSavedMessages returns the saved messages for the specified
// server proxy. The slice returned should not be mutated.
func (p *Proxy) getSavedMessages(id serverID) []lsp.ShowMessageParams {
	var c *serverProxyConn
	p.mu.Lock()
	for cc := range p.servers {
		if cc.id == id {
			c = cc
			break
		}
	}
	p.mu.Unlock()
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	// return a shallow copy as the slice may be mutated by other goroutines
	return c.messages[:]
}
