package api

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type inMemoryListener struct {
	connCh  chan net.Conn
	address net.Addr

	mu     sync.RWMutex
	closed bool
}

func newInMemoryListener(name string) *inMemoryListener {
	return &inMemoryListener{
		connCh:  make(chan net.Conn),
		address: inMemoryAddr{name: name},
	}
}

func (l *inMemoryListener) ContextDial(ctx context.Context, _ string) (net.Conn, error) {
	l.mu.RLock()
	closed := l.closed
	l.mu.RUnlock()

	if closed {
		return nil, errInMemoryListenerClosed
	}

	sideA, sideB := net.Pipe()

	select {
	case l.connCh <- sideA: // ready to load one side of the connection
		return sideB, nil
	case <-ctx.Done(): // context canceled before we could load the connection
		sideA.Close()
		sideB.Close()
		return nil, ctx.Err()
	}
}

func (l *inMemoryListener) Accept() (net.Conn, error) {
	l.mu.RLock()
	closed := l.closed
	l.mu.RUnlock()

	if closed {
		return nil, errInMemoryListenerClosed
	}

	conn := <-l.connCh
	return conn, nil
}

func (l *inMemoryListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil
	}

	l.closed = true
	close(l.connCh)

	return nil
}

func (l *inMemoryListener) Addr() net.Addr {
	return l.address
}

type inMemoryAddr struct {
	name string
}

func (l inMemoryAddr) Network() string {
	return "inmemory"
}

func (l inMemoryAddr) String() string {
	return fmt.Sprintf("inmemory://%s", l.name)
}

var errInMemoryListenerClosed = errors.New("in memory listener closed")
