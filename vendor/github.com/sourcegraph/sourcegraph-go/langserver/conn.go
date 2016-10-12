package langserver

import (
	"context"

	"github.com/sourcegraph/jsonrpc2"
)

// JSONRPC2Conn is a limited interface to jsonrpc2.Conn. When the
// build server wraps the lang server, it provides this limited subset
// of methods. This interface exists to make it possible for the build
// server to provide the lang server with this limited connection
// handle.
type JSONRPC2Conn interface {
	Notify(ctx context.Context, method string, params interface{}, opt ...jsonrpc2.CallOption) error
}
