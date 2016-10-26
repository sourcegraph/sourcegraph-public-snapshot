package xlang

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/prefixsuffixsaver"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// NewDefaultClient returns a new one-shot connection to the LSP proxy server at
// $LSP_PROXY. This is quick (except TCP connection time) because the LSP proxy
// server retains the same underlying connection.
func NewDefaultClient() (*Client, error) {
	addr := os.Getenv("LSP_PROXY")
	if addr == "" {
		return nil, errors.New("no LSP_PROXY env var set (need address to LSP proxy)")
	}

	dialCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	conn, err := DialProxy(dialCtx, addr, nil)
	if err != nil {
		return nil, err
	}
	return &Client{
		Conn: conn,
	}, nil
}

// Client wraps a JSON RPC 2 connection and instruments it using OpenTracing
// from the context.
type Client struct {
	Conn *jsonrpc2.Conn
}

func (c *Client) withMeta(span opentracing.Span, opt []jsonrpc2.CallOption) []jsonrpc2.CallOption {
	carrier := opentracing.TextMapCarrier{}
	if err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.TextMap, carrier); err != nil {
		log.Println("warning: withMeta failed to inject global tracer", err)
		return opt
	}
	return append(opt, jsonrpc2.Meta(carrier))
}

func (c *Client) Call(ctx context.Context, method string, params, result interface{}, opt ...jsonrpc2.CallOption) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "LSP Call "+method)
	defer c.finishWithError(span, &err)
	span.LogEventWithPayload("request", params)

	// Store the result into tmpResult so that we can forward it onward to the
	// OpenTracing tracer. This doesn't support streaming, but neither does
	// Call internally, so this has zero overhead currently.
	var tmpResult json.RawMessage
	err = c.Conn.Call(ctx, method, params, &tmpResult, c.withMeta(span, opt)...)
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}

	// 1 KB is a good, safe choice for medium-to-high throughput traces.
	maxSize := 1 * 1024
	if len(tmpResult) < maxSize {
		// A small object, so no need to use prefix suffix saver. Unmarshal the
		// object into a map so that Lightstep formats it nicely.
		obj := make(map[string]interface{})
		if err2 := json.Unmarshal(tmpResult, &obj); err2 != nil {
			// Try an array, then.
			arr := make([]interface{}, 0)
			if err2 := json.Unmarshal(tmpResult, &arr); err2 != nil {
				span.LogEventWithPayload("response", string(tmpResult))
			}
			span.LogEventWithPayload("response", arr)
		} else {
			span.LogEventWithPayload("response", obj)
		}
	} else {
		saver := &prefixsuffixsaver.Writer{N: maxSize}
		io.Copy(saver, bytes.NewReader(tmpResult))
		span.LogEventWithPayload("response", string(saver.Bytes()))
	}
	return json.Unmarshal(tmpResult, result)
}

func (c *Client) Notify(ctx context.Context, method string, params interface{}, opt ...jsonrpc2.CallOption) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "LSP Notify "+method)
	defer c.finishWithError(span, &err)
	span.LogEventWithPayload("request", params)

	return c.Conn.Notify(ctx, method, params, c.withMeta(span, opt)...)
}

func (c *Client) Close() (err error) { return c.Conn.Close() }

func (c *Client) finishWithError(span opentracing.Span, err *error) {
	e := *err
	if e != nil {
		ext.Error.Set(span, true)
		span.SetTag("err", e.Error())
	}
	span.Finish()
}

// OneShotClientRequest performs a one-shot LSP client request to the specified
// method (e.g. "textDocument/definition") and stores the results in the given
// pointer value.
func OneShotClientRequest(ctx context.Context, mode, rootPath, method string, params, results interface{}) error {
	// Connect to the xlang proxy.
	c, err := NewDefaultClient()
	if err != nil {
		return errors.Wrap(err, "NewDefaultClient")
	}
	defer c.Close()

	// Initialize the connection.
	err = c.Call(ctx, "initialize", ClientProxyInitializeParams{
		InitializeParams: lsp.InitializeParams{
			RootPath: rootPath,
		},
		Mode: mode,
	}, nil)
	if err != nil {
		return errors.Wrap(err, "LSP initialize")
	}

	// Perform the request.
	err = c.Call(ctx, method, params, results)
	if err != nil {
		return errors.Wrap(err, "LSP "+method)
	}

	// Shutdown the connection.
	err = c.Call(ctx, "shutdown", nil, nil)
	if err != nil {
		return errors.Wrap(err, "LSP shutdown")
	}
	err = c.Notify(ctx, "exit", nil)
	if err != nil {
		return errors.Wrap(err, "LSP exit")
	}
	return nil
}
