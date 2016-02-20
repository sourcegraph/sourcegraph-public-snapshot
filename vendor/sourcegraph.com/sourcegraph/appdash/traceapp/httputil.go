// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package traceapp

import (
	"bytes"
	"net/http"
	"strconv"
)

type responseBuffer struct {
	buf    bytes.Buffer
	Status int
	header http.Header
}

func (rb *responseBuffer) Write(p []byte) (int, error) {
	return rb.buf.Write(p)
}

func (rb *responseBuffer) WriteHeader(status int) {
	rb.Status = status
}

func (rb *responseBuffer) Header() http.Header {
	if rb.header == nil {
		rb.header = make(http.Header)
	}
	return rb.header
}

func (rb *responseBuffer) ContentLength() int {
	return rb.buf.Len()
}

func (rb *responseBuffer) WriteTo(w http.ResponseWriter) error {
	for k, v := range rb.header {
		w.Header()[k] = v
	}
	if l := rb.ContentLength(); l > 0 {
		w.Header().Set("Content-Length", strconv.Itoa(l))
	}
	if rb.Status != 0 {
		w.WriteHeader(rb.Status)
	}
	if rb.buf.Len() > 0 {
		if _, err := w.Write(rb.buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}
