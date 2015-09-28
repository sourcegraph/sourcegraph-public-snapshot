// Copyright 2014 The transport Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package transport provides a general-purpose wapper for
// http.RoundTripper. It implements the pattern of taking a
// request, modifying a copy, and passing the modified copy to an
// underlying RoundTripper, including bookkeeping necessary to
// cancel in-flight requests.
package transport

import (
	"errors"
	"io"
	"net/http"
	"sync"
)

// Wrapper is an http.RoundTripper that makes HTTP requests,
// wrapping a base RoundTripper and altering every outgoing
// request in some way.
type Wrapper struct {
	// Modify alters the request as needed.
	Modify func(*http.Request) error

	// Base is the base RoundTripper used to make HTTP requests.
	// If nil, http.DefaultTransport is used.
	Base http.RoundTripper

	mu     sync.Mutex                      // guards modReq
	modReq map[*http.Request]*http.Request // original -> modified
}

// RoundTrip provides a copy of req
// to the underlying RoundTripper,
// altered in some way by Modify.
func (t *Wrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Modify == nil {
		return nil, errors.New("transport: Wrapper's Modify is nil")
	}

	req2 := cloneRequest(req) // per RoundTripper contract
	err := t.Modify(req2)
	if err != nil {
		return nil, err
	}
	t.setModReq(req, req2)
	res, err := t.base().RoundTrip(req2)
	if err != nil {
		t.setModReq(req, nil)
		return nil, err
	}
	res.Body = &onEOFReader{
		rc: res.Body,
		fn: func() { t.setModReq(req, nil) },
	}
	return res, nil
}

// CancelRequest cancels an in-flight request by closing its connection.
func (t *Wrapper) CancelRequest(req *http.Request) {
	type canceler interface {
		CancelRequest(*http.Request)
	}
	if cr, ok := t.base().(canceler); ok {
		t.mu.Lock()
		modReq := t.modReq[req]
		delete(t.modReq, req)
		t.mu.Unlock()
		cr.CancelRequest(modReq)
	}
}

func (t *Wrapper) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

func (t *Wrapper) setModReq(orig, mod *http.Request) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.modReq == nil {
		t.modReq = make(map[*http.Request]*http.Request)
	}
	if mod == nil {
		delete(t.modReq, orig)
	} else {
		t.modReq[orig] = mod
	}
}

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}

type onEOFReader struct {
	rc io.ReadCloser
	fn func()
}

func (r *onEOFReader) Read(p []byte) (n int, err error) {
	n, err = r.rc.Read(p)
	if err == io.EOF {
		r.runFunc()
	}
	return
}

func (r *onEOFReader) Close() error {
	err := r.rc.Close()
	r.runFunc()
	return err
}

func (r *onEOFReader) runFunc() {
	if fn := r.fn; fn != nil {
		fn()
		r.fn = nil
	}
}
