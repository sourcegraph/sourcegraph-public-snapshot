// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// This file contains helper functions regarding compression/decompression for confighttp.

package confighttp // import "go.opentelemetry.io/collector/config/confighttp"

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"net/http"

	"github.com/golang/snappy"
	"github.com/klauspost/compress/zstd"

	"go.opentelemetry.io/collector/config/configcompression"
)

type compressRoundTripper struct {
	rt              http.RoundTripper
	compressionType configcompression.Type
	compressor      *compressor
}

var availableDecoders = map[string]func(body io.ReadCloser) (io.ReadCloser, error){
	"": func(io.ReadCloser) (io.ReadCloser, error) {
		// Not a compressed payload. Nothing to do.
		return nil, nil
	},
	"gzip": func(body io.ReadCloser) (io.ReadCloser, error) {
		gr, err := gzip.NewReader(body)
		if err != nil {
			return nil, err
		}
		return gr, nil
	},
	"zstd": func(body io.ReadCloser) (io.ReadCloser, error) {
		zr, err := zstd.NewReader(
			body,
			// Concurrency 1 disables async decoding. We don't need async decoding, it is pointless
			// for our use-case (a server accepting decoding http requests).
			// Disabling async improves performance (I benchmarked it previously when working
			// on https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23257).
			zstd.WithDecoderConcurrency(1),
		)
		if err != nil {
			return nil, err
		}
		return zr.IOReadCloser(), nil
	},
	"zlib": func(body io.ReadCloser) (io.ReadCloser, error) {
		zr, err := zlib.NewReader(body)
		if err != nil {
			return nil, err
		}
		return zr, nil
	},
	"snappy": func(body io.ReadCloser) (io.ReadCloser, error) {
		sr := snappy.NewReader(body)
		sb := new(bytes.Buffer)
		_, err := io.Copy(sb, sr)
		if err != nil {
			return nil, err
		}
		if err = body.Close(); err != nil {
			return nil, err
		}
		return io.NopCloser(sb), nil
	},
}

func newCompressRoundTripper(rt http.RoundTripper, compressionType configcompression.Type) (*compressRoundTripper, error) {
	encoder, err := newCompressor(compressionType)
	if err != nil {
		return nil, err
	}
	return &compressRoundTripper{
		rt:              rt,
		compressionType: compressionType,
		compressor:      encoder,
	}, nil
}

func (r *compressRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get(headerContentEncoding) != "" {
		// If the header already specifies a content encoding then skip compression
		// since we don't want to compress it again. This is a safeguard that normally
		// should not happen since CompressRoundTripper is not intended to be used
		// with http clients which already do their own compression.
		return r.rt.RoundTrip(req)
	}

	// Compress the body.
	buf := bytes.NewBuffer([]byte{})
	if err := r.compressor.compress(buf, req.Body); err != nil {
		return nil, err
	}

	// Create a new request since the docs say that we cannot modify the "req"
	// (see https://golang.org/pkg/net/http/#RoundTripper).
	cReq, err := http.NewRequestWithContext(req.Context(), req.Method, req.URL.String(), buf)
	if err != nil {
		return nil, err
	}

	// Clone the headers and add the encoding header.
	cReq.Header = req.Header.Clone()
	cReq.Header.Add(headerContentEncoding, string(r.compressionType))

	return r.rt.RoundTrip(cReq)
}

type decompressor struct {
	errHandler         func(w http.ResponseWriter, r *http.Request, errorMsg string, statusCode int)
	base               http.Handler
	decoders           map[string]func(body io.ReadCloser) (io.ReadCloser, error)
	maxRequestBodySize int64
}

// httpContentDecompressor offloads the task of handling compressed HTTP requests
// by identifying the compression format in the "Content-Encoding" header and re-writing
// request body so that the handlers further in the chain can work on decompressed data.
// It supports gzip and deflate/zlib compression.
func httpContentDecompressor(h http.Handler, maxRequestBodySize int64, eh func(w http.ResponseWriter, r *http.Request, errorMsg string, statusCode int), enableDecoders []string, decoders map[string]func(body io.ReadCloser) (io.ReadCloser, error)) http.Handler {
	errHandler := defaultErrorHandler
	if eh != nil {
		errHandler = eh
	}

	enabled := map[string]func(body io.ReadCloser) (io.ReadCloser, error){}
	for _, dec := range enableDecoders {
		enabled[dec] = availableDecoders[dec]

		if dec == "deflate" {
			enabled["deflate"] = availableDecoders["zlib"]
		}
	}

	d := &decompressor{
		maxRequestBodySize: maxRequestBodySize,
		errHandler:         errHandler,
		base:               h,
		decoders:           enabled,
	}

	for key, dec := range decoders {
		d.decoders[key] = dec
	}

	return d
}

func (d *decompressor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	newBody, err := d.newBodyReader(r)
	if err != nil {
		d.errHandler(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	if newBody != nil {
		defer newBody.Close()
		// "Content-Encoding" header is removed to avoid decompressing twice
		// in case the next handler(s) have implemented a similar mechanism.
		r.Header.Del("Content-Encoding")
		// "Content-Length" is set to -1 as the size of the decompressed body is unknown.
		r.Header.Del("Content-Length")
		r.ContentLength = -1
		r.Body = http.MaxBytesReader(w, newBody, d.maxRequestBodySize)
	}
	d.base.ServeHTTP(w, r)
}

func (d *decompressor) newBodyReader(r *http.Request) (io.ReadCloser, error) {
	encoding := r.Header.Get(headerContentEncoding)
	decoder, ok := d.decoders[encoding]
	if !ok {
		return nil, fmt.Errorf("unsupported %s: %s", headerContentEncoding, encoding)
	}
	return decoder(r.Body)
}

// defaultErrorHandler writes the error message in plain text.
func defaultErrorHandler(w http.ResponseWriter, _ *http.Request, errMsg string, statusCode int) {
	http.Error(w, errMsg, statusCode)
}
