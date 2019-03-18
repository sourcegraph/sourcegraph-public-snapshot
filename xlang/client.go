package xlang

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/prefixsuffixsaver"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

// DefaultAddr returns the TCP address (host:port) of the default LSP proxy service.
func DefaultAddr() string {
	addr := os.Getenv("LSP_PROXY")
	if addr == "" {
		return "lsp-proxy:4388"
	}
	return addr
}

// UnsafeNewDefaultClient returns a new one-shot connection to the LSP proxy server at
// $LSP_PROXY. This is quick (except TCP connection time) because the LSP proxy
// server retains the same underlying connection.
//
// SECURITY NOTE this does not check the user has permission to read the repo
// of any operation done on the Client. Please ensure the user has access to
// the repo.
func UnsafeNewDefaultClient() (*Client, error) {
	return dialClient(DefaultAddr())
}

func dialClient(addr string) (*Client, error) {
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
	span.LogFields(otlog.Object("request", params))

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
				span.LogFields(otlog.String("response", string(tmpResult)))
			}
			span.LogFields(otlog.Object("response", arr))
		} else {
			span.LogFields(otlog.Object("response", obj))
		}
	} else {
		saver := &prefixsuffixsaver.Writer{N: maxSize}
		io.Copy(saver, bytes.NewReader(tmpResult))
		span.LogFields(otlog.String("response", string(saver.Bytes())))
	}
	if tmpResult != nil {
		if err := json.Unmarshal(tmpResult, result); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Notify(ctx context.Context, method string, params interface{}, opt ...jsonrpc2.CallOption) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "LSP Notify "+method)
	defer c.finishWithError(span, &err)
	span.LogFields(otlog.Object("request", params))

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

// UnsafeOneShotClientRequest performs a one-shot LSP client request to the specified
// method (e.g. "textDocument/definition") and stores the results in the given
// pointer value.
//
// SECURITY NOTE this does not check the user has permission to read the
// repo. Please ensure the user has access to the repo.
func UnsafeOneShotClientRequest(ctx context.Context, mode string, rootURI lsp.DocumentURI, method string, params, results interface{}) error {
	// Connect to the xlang proxy.
	c, err := UnsafeNewDefaultClient()
	if err != nil {
		return errors.Wrap(err, "UnsafeNewDefaultClient")
	}
	defer c.Close()

	// Initialize the connection.
	err = c.Call(ctx, "initialize", lspext.ClientProxyInitializeParams{
		InitializeParams: lsp.InitializeParams{
			RootURI: rootURI,
		},
		InitializationOptions: lspext.ClientProxyInitializationOptions{Mode: mode},
		Mode:                  mode,
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

// RemoteOneShotClientRequest performs a one-shot LSP client request to the specified
// method (e.g. "textDocument/definition") and stores the results in the given
// pointer value. It does this against a remote Sourcegraph
func RemoteOneShotClientRequest(ctx context.Context, remote *url.URL, mode string, rootURI lsp.DocumentURI, method string, params, results interface{}) error {
	payload := []*jsonrpc2.Request{
		// Initialize the connection.
		{
			ID:     jsonrpc2.ID{Num: 0},
			Method: "initialize",
		},

		// Perform the request.
		{
			ID:     jsonrpc2.ID{Num: 1},
			Method: method,
		},

		// Shutdown the connection.
		{
			ID:     jsonrpc2.ID{Num: 2},
			Method: "shutdown",
		},
		{
			ID:     jsonrpc2.ID{Num: 3},
			Method: "exit",
			Notif:  true,
		},
	}

	initIdx := 0
	requestIdx := 1

	// init params
	err := payload[initIdx].SetParams(&lspext.ClientProxyInitializeParams{
		InitializeParams: lsp.InitializeParams{
			RootURI: rootURI,
		},
		InitializationOptions: lspext.ClientProxyInitializationOptions{Mode: mode},
		Mode:                  mode,
	})
	if err != nil {
		return err
	}

	// method params
	err = payload[requestIdx].SetParams(params)
	if err != nil {
		return err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	remote = &(*remote) // copy
	remote.Path = "/.api/xlang/" + method
	url := remote.String()

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sourcegraph-Client", "RemoteOneShotClientRequest")
	req.Header.Set("User-Agent", "Sourcegraph/"+env.Version)
	req = req.WithContext(ctx)

	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req,
		nethttp.OperationName("Remote "+method),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	resp, err := httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return err
	}
	defer resp.Body.Close()

	var respPayload []*jsonrpc2.Response
	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return err
	}

	for _, resp := range respPayload {
		if resp != nil && resp.ID.Num == payload[requestIdx].ID.Num { // the interesting request ID
			if resp.Error != nil {
				return resp.Error
			}
			b, err := resp.Result.MarshalJSON()
			if err != nil {
				return err
			}
			return json.Unmarshal(b, results)
		}
	}

	return errors.New("no response found for request")
}

var httpClient = &http.Client{Transport: &nethttp.Transport{}}
