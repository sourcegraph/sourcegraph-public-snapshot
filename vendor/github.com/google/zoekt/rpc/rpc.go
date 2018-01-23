// Package rpc provides a zoekt.Searcher over RPC.
package rpc

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"net/rpc"
	"sync"
	"time"

	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
	"github.com/google/zoekt/rpc/internal/srv"
)

// DefaultRPCPath is the normal path used by zoekt
const DefaultRPCPath = "/rpc"

// Server returns an http.Handler for searcher which is the server side of the
// RPC calls.
func Server(searcher zoekt.Searcher) http.Handler {
	registerGob()
	server := rpc.NewServer()
	server.Register(&srv.Searcher{Searcher: searcher})
	return server
}

// Client connects to a Searcher HTTP RPC server at address (host:port) using
// DefaultRPCPath path.
func Client(address string) zoekt.Searcher {
	return ClientAtPath(address, DefaultRPCPath)
}

// ClientAtPath connects to a Searcher HTTP RPC server at address and path
// (http://host:port/path).
func ClientAtPath(address, path string) zoekt.Searcher {
	registerGob()
	return &client{addr: address, path: path}
}

type client struct {
	addr, path string

	mu  sync.Mutex // protects client and gen
	cl  *rpc.Client
	gen int // incremented each time we dial
}

func (c *client) Search(ctx context.Context, q query.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	var reply srv.SearchReply
	err := c.call(ctx, "Searcher.Search", &srv.SearchArgs{Q: q, Opts: opts}, &reply)
	return reply.Result, err
}

func (c *client) List(ctx context.Context, q query.Q) (*zoekt.RepoList, error) {
	var reply srv.ListReply
	err := c.call(ctx, "Searcher.List", &srv.ListArgs{Q: q}, &reply)
	return reply.List, err
}

func (c *client) call(ctx context.Context, serviceMethod string, args interface{}, reply interface{}) (err error) {
	ctxDone := ctx.Done()
	callDone := make(chan *rpc.Call, 1)

	// We try twice. If we fail to dial or fail to call the function we try
	// again after 100ms. Unrolled to make logic clear
	cl, gen, err := c.getRPCClient(0)
	if err == nil {
		call := cl.Go(serviceMethod, args, reply, callDone)
		select {
		case <-ctxDone:
			return ctx.Err()
		case <-callDone:
			if call.Error == rpc.ErrShutdown {
				break
			}
			return call.Error
		}
	}

	select {
	case <-ctxDone:
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
	}

	cl, gen, err = c.getRPCClient(gen)
	if err != nil {
		return err
	}
	call := cl.Go(serviceMethod, args, reply, callDone)
	select {
	case <-ctxDone:
		return ctx.Err()
	case <-callDone:
		return call.Error
	}
}

// getRPCClient gets the rpc client. If gen matches the current generation, we
// redail and increment the generation. This is used to prevent concurrent
// redailing on network failure.
func (c *client) getRPCClient(gen int) (*rpc.Client, int, error) {
	// coarse lock so we only dial once
	c.mu.Lock()
	defer c.mu.Unlock()
	if gen != c.gen {
		return c.cl, c.gen, nil
	}
	cl, err := rpc.DialHTTPPath("tcp", c.addr, c.path)
	if err != nil {
		return nil, c.gen, err
	}
	c.cl = cl
	c.gen++
	return c.cl, c.gen, nil
}

func (c *client) Close() {
	c.cl.Close()
}

func (c *client) String() string {
	return fmt.Sprintf("rpcSearcher(%s/%s)", c.addr, c.path)
}

var once sync.Once

func registerGob() {
	once.Do(func() {
		gob.Register(&query.And{})
		gob.Register(&query.Or{})
		gob.Register(&query.Regexp{})
		gob.Register(&query.Language{})
		gob.Register(&query.Const{})
		gob.Register(&query.Repo{})
		gob.Register(&query.RepoSet{})
		gob.Register(&query.Substring{})
		gob.Register(&query.Not{})
		gob.Register(&query.Branch{})
	})
}
