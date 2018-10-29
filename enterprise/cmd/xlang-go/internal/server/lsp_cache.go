package server

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// lspCache implements httpcache.Cache
type lspCache struct {
	ctx  context.Context
	conn jsonrpc2.JSONRPC2
}

// Get implements httpcache.Cache.Get
func (c *lspCache) Get(key string) ([]byte, bool) {
	var b []byte
	ok, err := cacheGet(c.ctx, c.conn, key, &b)
	if err != nil {
		log15.Warn("failed to execute lspCache get", "cmd", "GET", "error", err)
		return nil, false
	}
	return b, ok
}

// Set implements httpcache.Cache.Set
func (c *lspCache) Set(key string, responseBytes []byte) {
	err := cacheSet(c.ctx, c.conn, key, responseBytes)
	if err != nil {
		excerpt := responseBytes
		if max := 25; len(excerpt) > max {
			excerpt = excerpt[:max]
		}
		log15.Warn("failed to execute lspCache set", "cmd", "SET", "error", err, "valueExcerpt", string(excerpt))
	}
}

// Delete implements httpcache.Cache.Delete
func (c *lspCache) Delete(key string) {
	// The lsp proxy doesn't expose a delete operation,
	// but setting the key to null should be equivalent
	// for our purposes
	err := cacheSet(c.ctx, c.conn, key, nil)
	if err != nil {
		log15.Warn("failed to execute lspCache delete", "error", err)
	}
}

// cacheGet will do a xcache/get request for key, and unmarshal the result
// into v on a cache hit. If it is a cache miss, false is returned.
func cacheGet(ctx context.Context, conn jsonrpc2.JSONRPC2, key string, v interface{}) (bool, error) {
	var r json.RawMessage
	err := conn.Call(ctx, "xcache/get", lspext.CacheGetParams{Key: key}, &r)
	if err != nil {
		return false, err
	}
	b := []byte(r)
	if bytes.Equal(b, []byte("null")) {
		return false, nil
	}
	err = json.Unmarshal(b, &v)
	return err == nil, err
}

// cacheSet will do a xcache/set request for key and a marshalled value.
func cacheSet(ctx context.Context, conn jsonrpc2.JSONRPC2, key string, v interface{}) error {
	b, _ := json.Marshal(v)
	m := json.RawMessage(b)
	return conn.Notify(ctx, "xcache/set", lspext.CacheSetParams{Key: key, Value: &m})
}
