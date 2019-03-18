package httpapi

import (
	"context"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/xlang"
)

// XLangNewClient returns a client that can provide certain cross-repository code intelligence
// capabilities, exposed through a standard LSP interface. By default it returns a vanilla LSP
// client, but external packages should overwrite it to extend the behavior.
var XLangNewClient = func() (XLangClient, error) {
	return xlang.UnsafeNewDefaultClient()
}

type XLangClient interface {
	Call(ctx context.Context, method string, params, result interface{}, opt ...jsonrpc2.CallOption) error
	Notify(ctx context.Context, method string, params interface{}, opt ...jsonrpc2.CallOption) error
	Close() error
}
