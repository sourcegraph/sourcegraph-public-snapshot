// Package rpc provides a search.Searcher over RPC.
package rpc

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/keegancsmith/rpc"
	"github.com/sourcegraph/sourcegraph/pkg/search"
	"github.com/sourcegraph/sourcegraph/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/pkg/search/rpc/internal/srv"
)

// DefaultRPCPath is the rpc path
const DefaultRPCPath = "/rpc"

// Server returns an http.Handler for searcher which is the server side of the
// RPC calls.
func Server(searcher search.Searcher) (http.Handler, error) {
	registerGob()
	server := rpc.NewServer()
	err := server.Register(&srv.Searcher{Searcher: searcher})
	return server, err
}

// Client connects to a Searcher HTTP RPC server at address (host:port) using
// DefaultRPCPath path.
func Client(address string) search.Searcher {
	return ClientAtPath(address, DefaultRPCPath)
}

// ClientAtPath connects to a Searcher HTTP RPC server at address and path
// (http://host:port/path).
func ClientAtPath(address, path string) search.Searcher {
	registerGob()
	return &client{addr: address, path: path}
}

type client struct {
	addr, path string

	mu  sync.Mutex // protects client and gen
	cl  *rpc.Client
	gen int // incremented each time we dial
}

func (c *client) Search(ctx context.Context, q query.Q, opts *search.Options) (*search.Result, error) {
	var reply srv.SearchReply
	err := c.call(ctx, "Searcher.Search", &srv.SearchArgs{Q: q, Opts: opts}, &reply)
	return reply.Result, err
}

func (c *client) call(ctx context.Context, serviceMethod string, args interface{}, reply interface{}) error {
	// We try twice. If we fail to dial or fail to call the function we try
	// again after 100ms. Unrolled to make logic clear
	cl, gen, err := c.getRPCClient(ctx, 0)
	if err == nil {
		err = cl.Call(ctx, serviceMethod, args, reply)
		if err != rpc.ErrShutdown {
			return err
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
	}

	cl, _, err = c.getRPCClient(ctx, gen)
	if err != nil {
		return err
	}
	return cl.Call(ctx, serviceMethod, args, reply)
}

// getRPCClient gets the rpc client. If gen matches the current generation, we
// redail and increment the generation. This is used to prevent concurrent
// redailing on network failure.
func (c *client) getRPCClient(ctx context.Context, gen int) (*rpc.Client, int, error) {
	// coarse lock so we only dial once
	c.mu.Lock()
	defer c.mu.Unlock()
	if gen != c.gen {
		return c.cl, c.gen, nil
	}
	var timeout time.Duration
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}
	cl, err := rpc.DialHTTPPathTimeout("tcp", c.addr, c.path, timeout)
	if err != nil {
		return nil, c.gen, err
	}
	c.cl = cl
	c.gen++
	return c.cl, c.gen, nil
}

func (c *client) Close() {
	c.mu.Lock()
	cl := c.cl
	c.mu.Unlock()
	if cl != nil {
		cl.Close()
	}
}

func (c *client) String() string {
	return fmt.Sprintf("rpcSearcher(%s/%s)", c.addr, c.path)
}

var once sync.Once

func registerGob() {
	once.Do(func() {
		gob.RegisterName("*sgquery.And", &query.And{})
		gob.RegisterName("*sgquery.Ref", &query.Ref{})
		gob.RegisterName("*sgquery.Const", &query.Const{})
		gob.RegisterName("*sgquery.Language", &query.Language{})
		gob.RegisterName("*sgquery.Not", &query.Not{})
		gob.RegisterName("*sgquery.Or", &query.Or{})
		gob.RegisterName("*sgquery.Regexp", &query.Regexp{})
		gob.RegisterName("*sgquery.RepoSet", &query.RepoSet{})
		gob.RegisterName("*sgquery.Repo", &query.Repo{})
		gob.RegisterName("*sgquery.Substring", &query.Substring{})
		gob.RegisterName("*sgquery.Type", &query.Type{})
		gob.RegisterName("*sgsearch.Repository", &search.Repository{})
		gob.RegisterName("*sgsearch.FileMatch", &search.FileMatch{})
		gob.RegisterName("*sgsearch.LineFragmentMatch", &search.LineFragmentMatch{})
		gob.RegisterName("*sgsearch.LineMatch", &search.LineMatch{})
		gob.RegisterName("*sgsearch.Options", &search.Options{})
		gob.RegisterName("*sgsearch.RepositoryStatus", &search.RepositoryStatus{})
		gob.RegisterName("*sgsearch.Result", &search.Result{})
		gob.RegisterName("*sgsearch.Stats", &search.Stats{})
	})
}
