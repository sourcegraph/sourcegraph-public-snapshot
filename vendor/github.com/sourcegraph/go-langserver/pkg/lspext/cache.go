package lspext

import "encoding/json"

// See https://github.com/sourcegraph/language-server-protocol/pull/14

// CacheGetParams is the input for 'cache/get'. The response is any or null.
type CacheGetParams struct {
	// Key is the cache key. The key namespace is shared for a language
	// server, but with other language servers. For example the PHP
	// language server on different workspaces share the same key
	// namespace, but does not share the namespace with a Go language
	// server.
	Key string `json:"key"`
}

// CacheSetParams is the input for the notification 'cache/set'.
type CacheSetParams struct {
	// Key is the cache key. The key namespace is shared for a language
	// server, but with other language servers. For example the PHP
	// language server on different workspaces share the same key
	// namespace, but does not share the namespace with a Go language
	// server.
	Key string `json:"key"`

	// Value is type any. We use json.RawMessage since we expect caching
	// implementation to cache the raw bytes, and not bother with
	// Unmarshaling/Marshalling.
	Value *json.RawMessage `json:"value"`
}
