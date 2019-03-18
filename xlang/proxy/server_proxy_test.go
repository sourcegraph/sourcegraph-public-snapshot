package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/rcache"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/go-lsp/lspext"
	"github.com/sourcegraph/jsonrpc2"
)

func TestCache(t *testing.T) {
	rcache.SetupForTest(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup fake "language server". connC will have a connection we can
	// use to do the test on
	var stream jsonrpc2.ObjectStream
	connC := make(chan *jsonrpc2.Conn, 1)
	{
		handler := func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
			if req.Method == "initialize" {
				connC <- conn
				return lsp.InitializeResult{}, nil
			}
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
		}
		a, b := InMemoryPeerConns()
		jsonrpc2.NewConn(ctx, a, jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError(handler)))
		stream = b
	}
	defer stream.Close()

	c := &serverProxyConn{
		id: serverID{contextID: contextID{mode: "cache-test"}},
	}
	c.conn = jsonrpc2.NewConn(ctx, stream, jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError(c.handle)))
	defer c.conn.Close()
	c.initOnce.Do(func() {
		_, err := c.lspInitialize(ctx)
		if err != nil {
			t.Fatal(err)
		}
	})

	// Everything above this is boilerplate that we should make easier to
	// avoid in the future

	conn := <-connC
	defer conn.Close()
	get := func(k string) (string, bool) {
		var v json.RawMessage
		err := conn.Call(ctx, "xcache/get", lspext.CacheGetParams{Key: k}, &v)
		if err != nil {
			t.Fatalf("get for %v failed: %v", k, err)
		}
		if bytes.Equal([]byte(v), []byte("null")) {
			return "", false
		}
		var s string
		_ = json.Unmarshal([]byte(v), &s)
		return s, true
	}
	set := func(k, v string) {
		b, _ := json.Marshal(v)
		m := json.RawMessage(b)
		err := conn.Notify(ctx, "xcache/set", lspext.CacheSetParams{Key: k, Value: &m})
		if err != nil {
			t.Fatalf("set for %v=%v failed: %v", k, v, err)
		}
	}

	if _, ok := get("x"); ok {
		t.Fatal("expected first get to fail")
	}
	set("y", "foo")
	if _, ok := get("x"); ok {
		t.Fatal("expected second get to fail")
	}
	for _, v := range []string{"hello", "", "world"} {
		set("x", v)
		// Sleep since notifications don't block
		time.Sleep(50 * time.Millisecond)
		got, ok := get("x")
		if !ok {
			t.Fatalf("expected get(x) = %v, got cache miss", v)
		}
		if got != v {
			t.Fatalf("expected get(x) = %v, got %v", v, got)
		}
	}
	got, ok := get("y")
	if !ok {
		t.Fatalf("expected get(y) = %v, got cache miss", "foo")
	}
	if got != "foo" {
		t.Fatalf("expected get(y) = %v, got %v", "foo", got)
	}
}
