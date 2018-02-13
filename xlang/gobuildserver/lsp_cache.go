package gobuildserver

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
func (c *lspCache) Get(key string) (responseBytes []byte, ok bool) {
	var r json.RawMessage
	err := c.conn.Call(c.ctx, "xcache/get", lspext.CacheGetParams{Key: key}, &r)

	if err != nil {
		log15.Warn("failed to execute lspCache get", "cmd", "GET", "error", err)
		ok = false
		return
	}
	responseBytes = []byte(r)
	ok = !bytes.Equal(responseBytes, []byte("null"))
	return
}

// Set implements httpcache.Cache.Set
func (c *lspCache) Set(key string, responseBytes []byte) {
	m := json.RawMessage(responseBytes)
	err := c.conn.Notify(c.ctx, "xcache/set", lspext.CacheSetParams{Key: key, Value: &m})

	if err != nil {
		log15.Warn("failed to execute lspCache set", "cmd", "SET", "error", err)
	}
	return
}

// Delete implements httpcache.Cache.Delete
func (c *lspCache) Delete(key string) {
	// The lsp proxy doesn't expose a delete operation,
	// but setting the key to null should be equivalent
	// for our purposes
	c.Set(key, []byte("null"))
}
