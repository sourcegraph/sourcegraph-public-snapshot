package xlang

import (
	"context"
	"log"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/neelance/parallel"
)

// NewProxy creates a new LSP proxy.
func NewProxy() *Proxy {
	return &Proxy{
		MaxClientIdle: 120 * time.Second,
		MaxServerIdle: 300 * time.Second,

		closed: make(chan struct{}),

		clients:          map[*clientProxyConn]struct{}{},
		servers:          map[*serverProxyConn]struct{}{},
		serverNewConnMus: map[serverID]*sync.Mutex{},
	}
}

// Proxy proxies LSP JSON-RPC 2.0 connections, sitting between the
// client (typically a user's browser, via our HTTP API) and
// lang/build servers.
type Proxy struct {
	MaxClientIdle time.Duration // disconnect idle clients after this duration
	MaxServerIdle time.Duration // shut down idle servers after this duration

	Trace bool // print traces of all requests/responses between proxy and client

	closed chan struct{} // a channel that is closed when (*Proxy).Close is called

	mu               sync.Mutex
	clients          map[*clientProxyConn]struct{} // open connections from clients
	servers          map[*serverProxyConn]struct{} // open connections to lang/build servers
	serverNewConnMus map[serverID]*sync.Mutex      // new pending connections to lang/build servers
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
			select {
			case <-done:
				return // stop when the listener is closed
			case <-time.After(p.MaxServerIdle / 2):
				ctx, cancel := context.WithTimeout(context.Background(), p.MaxServerIdle/2)
				if err := p.ShutDownIdleServers(ctx, p.MaxServerIdle); err != nil {
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
		p.newClientProxyConn(ctx, nc)
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
	// ShutDownIdleServers that are blocked on p.mu (which we hold) do
	// not attempt to double-close any client/server conns (thereby
	// causing a panic).
	p.clients = nil
	p.servers = nil

	// Only hold lock during fast loop iter; no need to wait for the
	// shutdowns/disconnects to complete.
	p.mu.Unlock()

	return par.Wait()
}
