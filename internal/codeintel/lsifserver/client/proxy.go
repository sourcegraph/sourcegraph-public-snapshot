package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

func (c *Client) RawRequest(ctx context.Context, req *http.Request) (_ *http.Response, err error) {
	router := mux.NewRouter()
	router.Path("/uploads/{id:[0-9]+}").Methods("GET").HandlerFunc(c.server.handleGetUploadByID)
	router.Path("/uploads/{id:[0-9]+}").Methods("DELETE").HandlerFunc(c.server.handleDeleteUploadByID)
	router.Path("/uploads/repository/{id:[0-9]+}").Methods("GET").HandlerFunc(c.server.handleGetUploadsByRepo)
	router.Path("/exists").Methods("GET").HandlerFunc(c.server.handleExists)
	router.Path("/definitions").Methods("GET").HandlerFunc(c.server.handleDefinitions)
	router.Path("/references").Methods("GET").HandlerFunc(c.server.handleReferences)
	router.Path("/hover").Methods("GET").HandlerFunc(c.server.handleHover)
	router.Path("/upload").Methods("POST").HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := c.rawRequest(ctx, req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return &http.Response{
		StatusCode: rec.Code,
		Header:     rec.Header(),
		Body:       ioutil.NopCloser(bytes.NewReader(rec.Body.Bytes())),
	}, nil
}

func (c *Client) rawRequest(ctx context.Context, req *http.Request) (_ *http.Response, err error) {
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
