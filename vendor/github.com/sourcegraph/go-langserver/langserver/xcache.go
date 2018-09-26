package langserver

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
)

// cacheGet will do a xcache/get request for key, and unmarshal the result
// into v on a cache hit. If it is a cache miss, false is returned. If the
// client is not a XCacheProvider, cacheGet will always return false.
func (h *LangHandler) cacheGet(ctx context.Context, conn jsonrpc2.JSONRPC2, key string, v interface{}) bool {
	if !h.init.Capabilities.XCacheProvider {
		return false
	}

	var r json.RawMessage
	err := conn.Call(ctx, "xcache/get", lspext.CacheGetParams{Key: key}, &r)
	if err != nil {
		return false
	}
	b := []byte(r)
	if bytes.Equal(b, []byte("null")) {
		return false
	}
	err = json.Unmarshal(b, &v)
	return err == nil
}

// cacheSet will do a xcache/set request for key and a marshalled value. If
// the client is not a XCacheProvider, cacheSet will do nothing.
func (h *LangHandler) cacheSet(ctx context.Context, conn jsonrpc2.JSONRPC2, key string, v interface{}) {
	if !h.init.Capabilities.XCacheProvider {
		return
	}

	b, _ := json.Marshal(v)
	m := json.RawMessage(b)
	_ = conn.Notify(ctx, "xcache/set", lspext.CacheSetParams{Key: key, Value: &m})
}
