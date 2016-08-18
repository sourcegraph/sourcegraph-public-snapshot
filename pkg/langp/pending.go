package langp

import (
	"log"
	"sync"
	"time"
)

type pending struct {
	*sync.Mutex
	m map[string]bool
}

func newPending() *pending {
	return &pending{
		Mutex: &sync.Mutex{},
		m:     make(map[string]bool),
	}
}

// acquire acquires ownership of preparing k. If k is already being prepared
// by someone else, this methods blocks until preparation of k is finished
// and handled=true is returned.
//
// When finished with preparation, done should always be called. If acquire
// did not acquire ownership, done is no-op.
func (p *pending) acquire(k string, timeout time.Duration) (wasTimeout, handled bool, done func()) {
	// If nobody is preparing k, mark ownership and return:
	p.Lock()
	if _, pending := p.m[k]; !pending {
		p.m[k] = true
		p.Unlock()
		done = func() {
			p.Lock()
			_, pending := p.m[k]
			if !pending {
				p.Unlock()
				panic("pending: done() called for non-acquired k")
			}
			delete(p.m, k)
			p.Unlock()
		}
		handled = false
		return
	}
	p.Unlock()

	// Someone is preparing k, wait for completion.
	done = func() {}
	log.Printf("preparation of k=%q already underway; waiting\n", k)
	start := time.Now()
	for {
		p.Lock()
		_, pending := p.m[k]
		p.Unlock()
		if !pending {
			handled = true
			return
		}
		if time.Since(start) > timeout {
			wasTimeout = true
			log.Printf("preparation of k=%q finished\n", k)
			return
		}
		// TODO: timeout request
		time.Sleep(1 * time.Millisecond)
	}
}
