package httpsnoop

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

// HeaderFunc is part of the http.ResponseWriter interface.
type HeaderFunc func() http.Header

// WriteHeaderFunc is part of the http.ResponseWriter interface.
type WriteHeaderFunc func(int)

// WriteFunc is part of the http.ResponseWriter interface.
type WriteFunc func([]byte) (int, error)

// FlushFunc is part of the http.Flusher interface.
type FlushFunc func()

// CloseNotifyFunc is part of the http.CloseNotifier interface.
type CloseNotifyFunc func() <-chan bool

// HijackFunc is part of the http.Hijacker interface.
type HijackFunc func() (net.Conn, *bufio.ReadWriter, error)

// ReadFromFunc is part of the io.ReaderFrom interface.
type ReadFromFunc func(src io.Reader) (int64, error)

// Hooks defines a set of method interceptors for methods included in
// http.ResponseWriter as well as some others. You can think of them as
// middleware for the function calls they target. See Wrap for more details.
type Hooks struct {
	Header      func(HeaderFunc) HeaderFunc
	Write       func(WriteFunc) WriteFunc
	WriteHeader func(WriteHeaderFunc) WriteHeaderFunc
	Flush       func(FlushFunc) FlushFunc
	CloseNotify func(CloseNotifyFunc) CloseNotifyFunc
	ReadFrom    func(ReadFromFunc) ReadFromFunc
	Hijack      func(HijackFunc) HijackFunc
}

// Wrap returns a wrapped version of w that provides the exact same interface
// as w. Specifically if w implements any combination of http.Hijacker,
// http.Flusher, http.CloseNotifier or io.ReaderFrom, the wrapped version will
// implement the exact same combination. If no hooks are set, the wrapped
// version also behaves exactly as w. Hooks targeting methods not supported by
// w are ignored. Any other hooks will intercept the method they target and may
// modify the call's arguments and/or return values. The CaptureMetrics
// implementation serves as a working example for how the hooks can be used.
func Wrap(w http.ResponseWriter, hooks Hooks) http.ResponseWriter {
	// TODO(fg) Go 1.7 has reflect.StructOf which could possibly replace the
	// unfortunate abomination below. However, for now I care about Go 1.6
	// support, and the performance impact of using reflect may also be
	// considerable.
	// See https://github.com/golang/go/issues/16726 if you're interested in
	// why I can't use Go 1.7 until my new MBP arrives ...

	rw := &rw{w: w, h: hooks}
	_, h := w.(http.Hijacker)
	_, f := w.(http.Flusher)
	_, cn := w.(http.CloseNotifier)
	_, rf := w.(io.ReaderFrom)

	switch {
	case h && f && cn && rf:
		return struct {
			http.ResponseWriter
			http.Hijacker
			http.Flusher
			http.CloseNotifier
			io.ReaderFrom
		}{rw, rw, rw, rw, rw}
	case h && f && cn && !rf:
		return struct {
			http.ResponseWriter
			http.Hijacker
			http.Flusher
			http.CloseNotifier
		}{rw, rw, rw, rw}
	case h && f && !cn && rf:
		return struct {
			http.ResponseWriter
			http.Hijacker
			http.Flusher
			io.ReaderFrom
		}{rw, rw, rw, rw}
	case h && f && !cn && !rf:
		return struct {
			http.ResponseWriter
			http.Hijacker
			http.Flusher
		}{rw, rw, rw}
	case h && !f && cn && rf:
		return struct {
			http.ResponseWriter
			http.Hijacker
			http.CloseNotifier
			io.ReaderFrom
		}{rw, rw, rw, rw}
	case h && !f && cn && !rf:
		return struct {
			http.ResponseWriter
			http.Hijacker
			http.CloseNotifier
		}{rw, rw, rw}
	case h && !f && !cn && rf:
		return struct {
			http.ResponseWriter
			http.Hijacker
			io.ReaderFrom
		}{rw, rw, rw}
	case h && !f && !cn && !rf:
		return struct {
			http.ResponseWriter
			http.Hijacker
		}{rw, rw}
	case !h && f && cn && rf:
		return struct {
			http.ResponseWriter
			http.Flusher
			http.CloseNotifier
			io.ReaderFrom
		}{rw, rw, rw, rw}
	case !h && f && cn && !rf:
		return struct {
			http.ResponseWriter
			http.Flusher
			http.CloseNotifier
		}{rw, rw, rw}
	case !h && f && !cn && rf:
		return struct {
			http.ResponseWriter
			http.Flusher
			io.ReaderFrom
		}{rw, rw, rw}
	case !h && f && !cn && !rf:
		return struct {
			http.ResponseWriter
			http.Flusher
		}{rw, rw}
	case !h && !f && cn && rf:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
			io.ReaderFrom
		}{rw, rw, rw}
	case !h && !f && cn && !rf:
		return struct {
			http.ResponseWriter
			http.CloseNotifier
		}{rw, rw}
	case !h && !f && !cn && rf:
		return struct {
			http.ResponseWriter
			io.ReaderFrom
		}{rw, rw}
	case !h && !f && !cn && !rf:
		return struct {
			http.ResponseWriter
		}{rw}
	}
	panic("unreachable")
}

type rw struct {
	w http.ResponseWriter
	h Hooks
}

func (w *rw) Header() http.Header {
	fn := w.w.Header
	if w.h.Header != nil {
		fn = w.h.Header(fn)
	}
	return fn()
}

func (w *rw) WriteHeader(code int) {
	fn := w.w.WriteHeader
	if w.h.WriteHeader != nil {
		fn = w.h.WriteHeader(fn)
	}
	fn(code)
}

func (w *rw) Write(b []byte) (int, error) {
	f := w.w.Write
	if w.h.Write != nil {
		f = w.h.Write(f)
	}
	return f(b)
}

func (w *rw) Flush() {
	f := w.w.(http.Flusher).Flush
	if w.h.Flush != nil {
		f = w.h.Flush(f)
	}
	f()
}

func (w *rw) CloseNotify() <-chan bool {
	f := w.w.(http.CloseNotifier).CloseNotify
	if w.h.CloseNotify != nil {
		f = w.h.CloseNotify(f)
	}
	return f()
}

func (w *rw) ReadFrom(src io.Reader) (int64, error) {
	f := w.w.(io.ReaderFrom).ReadFrom
	if w.h.ReadFrom != nil {
		f = w.h.ReadFrom(f)
	}
	return f(src)
}

func (w *rw) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	f := w.w.(http.Hijacker).Hijack
	if w.h.Hijack != nil {
		f = w.h.Hijack(f)
	}
	return f()
}
