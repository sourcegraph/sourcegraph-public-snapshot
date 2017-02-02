package zap

import (
	"context"

	"github.com/sourcegraph/jsonrpc2"
)

// RegisterServerExtension registers a new JSON-RPC 2.0
// request/notification method and handler on the server. It should be
// called at init time and must not be called after any Zap server has
// started.
//
// Server extension methods are completely unrelated to editor
// extensions (other than both having the word "extension" in their
// name).
func RegisterServerExtension(method string, handler ServerExtensionHandler) {
	if serverExtensions == nil {
		serverExtensions = map[string]ServerExtensionHandler{}
	}
	if _, present := serverExtensions[method]; present {
		panic("an extension handler already exists for method " + method)
	}
	serverExtensions[method] = handler
}

// ServerExtensionHandler is an extension function called by the
// server when it receives a JSON-RPC 2.0 request/notification whose
// method is registered to a server extension.
type ServerExtensionHandler func(context.Context, *jsonrpc2.Conn, *jsonrpc2.Request) (interface{}, error)

// serverExtensions holds handler functions registered by
// RegisterServerExtension.
var serverExtensions map[string]ServerExtensionHandler
