package httputil

import (
	"bytes"
	"net/http"
	"strconv"
)

type ResponseBuffer struct {
	buf    bytes.Buffer
	Status int
	header http.Header

	// WriteContentLength is whether to write a Content-Length
	// header. This should be false this ResponseBuffer will be
	// written to an encoded ResponseWriter (e.g., gzipped), because
	// Content-Length must be the encoded length, not the unencoded
	// length.
	WriteContentLength bool
}

func (rb *ResponseBuffer) Write(p []byte) (int, error) {
	return rb.buf.Write(p)
}

func (rb *ResponseBuffer) WriteHeader(status int) {
	rb.Status = status
}

func (rb *ResponseBuffer) Header() http.Header {
	if rb.header == nil {
		rb.header = make(http.Header)
	}
	return rb.header
}

func (rb *ResponseBuffer) ContentLength() int {
	return rb.buf.Len()
}

func (rb *ResponseBuffer) WriteTo(w http.ResponseWriter) error {
	for k, v := range rb.header {
		if !rb.WriteContentLength && http.CanonicalHeaderKey(k) == "Content-Length" {
			continue
		}
		w.Header()[k] = v
	}
	if rb.WriteContentLength {
		if l := rb.ContentLength(); l > 0 {
			w.Header().Set("Content-Length", strconv.Itoa(l))
		}
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
