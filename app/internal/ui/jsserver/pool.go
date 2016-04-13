package jsserver

import (
	"encoding/json"
	"fmt"
	"sync"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
)

// NewPool creates a pool of servers.
func NewPool(js []byte, size int) Server {
	p := &pool{
		js:      js,
		servers: make(chan Server, size),
	}

	// Fill pool initially.
	go func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.servers == nil {
			return
		}
		for i := 0; i < size; i++ {
			s, err := New(js)
			if err != nil {
				log15.Error("Failed to preinitialize server in jsserver pool.", "err", err)

				// Call will try to reinitialize this server later,
				// and it'll return an error synchronously if
				// initialization fails then.
				p.servers <- nil
			} else {
				p.servers <- s
			}
		}
	}()

	return p
}

type pool struct {
	js []byte

	mu      sync.Mutex
	servers chan Server
}

// Call sends the argument to one of the servers in the pool. If an
// error occurs, it attempts to replenish the pool by creating a new
// server in the failed server's place.
func (p *pool) Call(ctx context.Context, arg json.RawMessage) ([]byte, error) {
	s := <-p.servers

	if s == nil {
		var err error
		s, err = New(p.js)
		if err != nil {
			p.servers <- nil
			return nil, err
		}
	}

	resp, err := s.Call(ctx, arg)
	if err != nil {
		// Kill this server (in case it's not an ephemeral error) and
		// recreate it next time it is called.
		if err2 := s.Close(); err2 != nil {
			err = fmt.Errorf("%s (also failed to close jsserver: %s)", err, err2)
		}
		s = nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.servers != nil {
		p.servers <- s
		return resp, err
	}

	// The pool was closed after we obtained s, so close this server (so nobody else can use it).
	if s != nil {
		if err2 := s.Close(); err != nil {
			err = fmt.Errorf("%s (also failed to close jsserver after pool was closed: %s)", err, err2)
		}
	}
	return resp, err
}

// Close closes all servers in the pool. After calling Close, the
// behavior of calling Call is undefined.
func (p *pool) Close() error {
	p.mu.Lock()
	if p.servers == nil {
		p.mu.Unlock()
		return nil
	}
	servers := p.servers
	p.servers = nil
	close(servers)
	p.mu.Unlock()

	// Close all servers buffered in the channel. Any in-flight calls
	// to Call with close their servers before returning.
	var anyErr error
	for s := range servers {
		if s == nil {
			continue
		}
		if err := s.Close(); err != nil {
			anyErr = err
		}
	}
	return anyErr
}
