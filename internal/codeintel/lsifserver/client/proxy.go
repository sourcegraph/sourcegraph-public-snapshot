package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

func (c *Client) RawRequest(ctx context.Context, req *http.Request) (_ *http.Response, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "lsifserver.client.do")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	req, ht := nethttp.TraceRequest(
		span.Tracer(),
		req.WithContext(ctx),
		nethttp.OperationName("LSIF client"),
		nethttp.ClientTrace(false),
	)
	defer ht.Finish()

	// Do not use ctxhttp.Do here as it will re-wrap the request
	// with a context and this will causes the ot-headers not to
	// propagate correctly.
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return nil, errors.Wrap(err, "lsif request failed")
	}

	return resp, nil
}

func SelectRandomHost() (string, error) {
	endpoints, err := LSIFURLs().Endpoints()
	if err != nil {
		return "", err
	}

	for host := range endpoints {
		return host, nil
	}

	return "", fmt.Errorf("no endpoints available")
}
