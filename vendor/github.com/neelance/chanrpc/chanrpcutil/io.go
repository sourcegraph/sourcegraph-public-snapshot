package chanrpcutil

import "io"

type reader struct {
	c   <-chan []byte
	buf []byte
}

func NewReader(c <-chan []byte) io.Reader {
	return &reader{c: c}
}

func (r *reader) Read(p []byte) (n int, err error) {
	for len(r.buf) == 0 {
		var ok bool
		r.buf, ok = <-r.c
		if !ok {
			return 0, io.EOF
		}
	}

	n = copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}

type writer struct {
	c   chan<- []byte
	buf []byte
}

func (w *writer) Write(p []byte) (int, error) {
	n := len(p)
	if len(w.buf)+n > cap(w.buf) {
		w.c <- w.buf
		w.buf = make([]byte, 0, cap(w.buf))
	}
	w.buf = append(w.buf, p...)
	return n, nil
}

func (w *writer) Close() error {
	if len(w.buf) != 0 {
		w.c <- w.buf
	}
	close(w.c)
	w.buf = nil
	return nil
}

func NewWriter() (<-chan []byte, io.WriteCloser) {
	return NewWriterSize(1024 * 1024)
}

func NewWriterSize(bufSize int) (<-chan []byte, io.WriteCloser) {
	c := make(chan []byte, 10)
	return c, &writer{c: c, buf: make([]byte, 0, bufSize)}
}
