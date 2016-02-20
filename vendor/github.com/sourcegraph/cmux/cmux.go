package cmux

import (
	"fmt"
	"io"
	"net"
)

// Matcher matches a connection based on its content.
type Matcher func(r io.Reader) (ok bool)

// ErrorHandler handles an error and returns whether
// the mux should continue serving the listener.
type ErrorHandler func(err error) (ok bool)

// ErrNotMatched is returned whenever a connection is not matched by any of
// the matchers registered in the multiplexer.
type ErrNotMatched struct {
	c net.Conn
}

func (e ErrNotMatched) Error() string {
	return fmt.Sprintf("mux: connection %v not matched by an matcher",
		e.c.RemoteAddr())
}

func (e ErrNotMatched) Temporary() bool { return true }
func (e ErrNotMatched) Timeout() bool   { return false }

type errListenerClosed string

func (e errListenerClosed) Error() string   { return string(e) }
func (e errListenerClosed) Temporary() bool { return false }
func (e errListenerClosed) Timeout() bool   { return false }

var (
	ErrListenerClosed = errListenerClosed("mux: listener closed")
)

// New instantiates a new connection multiplexer.
func New(l net.Listener) CMux {
	// if !flag.Parsed() {
	// 	flag.Parse()
	// }
	return &cMux{
		root:   l,
		bufLen: 1024,
		errh:   func(err error) bool { return true },
	}
}

// CMux is a multiplexer for network connections.
type CMux interface {
	// Match returns a net.Listener that sees (i.e., accepts) only
	// the connections matched by at least one of the matcher.
	//
	// The order used to call Match determines the priority of matchers.
	Match(matchers ...Matcher) net.Listener
	// Serve starts multiplexing the listener. Serve blocks and perhaps
	// should be invoked concurrently within a go routine.
	Serve() error
	// HandleError registers an error handler that handles listener errors.
	HandleError(h ErrorHandler)
}

type matchersListener struct {
	ss []Matcher
	l  muxListener
}

type cMux struct {
	root   net.Listener
	bufLen int
	errh   ErrorHandler
	sls    []matchersListener
}

func (m *cMux) Match(matchers ...Matcher) (l net.Listener) {
	ml := muxListener{
		Listener: m.root,
		connc:    make(chan net.Conn, m.bufLen),
		donec:    make(chan struct{}),
	}
	m.sls = append(m.sls, matchersListener{ss: matchers, l: ml})
	return ml
}

func (m *cMux) Serve() error {
	defer func() {
		for _, sl := range m.sls {
			close(sl.l.donec)
		}
	}()

	for {
		c, err := m.root.Accept()
		if err != nil {
			if !m.handleErr(err) {
				return err
			}
			continue
		}

		go m.serve(c)
	}
}

func (m *cMux) serve(c net.Conn) {
	muc := newMuxConn(c)
	matched := false
	for _, sl := range m.sls {
		for _, s := range sl.ss {
			matched = s(muc.sniffer())
			muc.reset()
			if matched {
				select {
				// TODO(soheil): threre is a possiblity of having unclosed connection.
				case sl.l.connc <- muc:
				case <-sl.l.donec:
					c.Close()
				}
				return
			}
		}
	}

	if !matched {
		c.Close()
		err := ErrNotMatched{c: c}
		if !m.handleErr(err) {
			m.root.Close()
		}
	}
}

func (m *cMux) HandleError(h ErrorHandler) {
	m.errh = h
}

func (m *cMux) handleErr(err error) bool {
	if !m.errh(err) {
		return false
	}

	if ne, ok := err.(net.Error); ok {
		return ne.Temporary()
	}

	return false
}

type muxListener struct {
	net.Listener
	connc chan net.Conn
	donec chan struct{}
}

func (l muxListener) Accept() (c net.Conn, err error) {
	c, ok := <-l.connc
	if !ok {
		return nil, ErrListenerClosed
	}
	return c, nil
}

type MuxConn struct {
	net.Conn
	buf buffer
}

func newMuxConn(c net.Conn) *MuxConn {
	return &MuxConn{
		Conn: c,
	}
}

func (m *MuxConn) Read(b []byte) (n int, err error) {
	if n, err = m.buf.Read(b); err == nil {
		return
	}

	n, err = m.Conn.Read(b)
	return
}

func (m *MuxConn) sniffer() io.Reader {
	return io.MultiReader(&m.buf, io.TeeReader(m.Conn, &m.buf))
}

func (m *MuxConn) reset() {
	m.buf.resetRead()
}
