package httputil

import (
	"bytes"
	"net/http"
)

type ResponseBuffer struct {
	buf    bytes.Buffer
	Status int
	header http.Header
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

func (rb *ResponseBuffer) WriteTo(w http.ResponseWriter) error {
	for k, v := range rb.header {
		if http.CanonicalHeaderKey(k) == "Content-Length" {
			continue
		}
		w.Header()[k] = v
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
