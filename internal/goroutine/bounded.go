package goroutine

import (
	"sync"
)

// Bounded runs a bounded number of goroutines. It supports waiting for them
// all to run, as well as reporting any error that may occur.
type Bounded struct {
	sema chan struct{}
	mu   sync.Mutex
	err  error
}

// NewBounded initializes Bounded with a capacity.
func NewBounded(capacity int) *Bounded {
	return &Bounded{
		sema: make(chan struct{}, capacity),
	}
}

// Go runs f in a new goroutine. It will only run upto Bounded.N goroutines at
// a time. Go will block until it can start the goroutine.
//
// The first f to return a non-nil error will have that error returned by
// Wait. If an f fails, this does not stop future runs.
func (s *Bounded) Go(f func() error) {
	s.sema <- struct{}{}

	go func() {
		defer func() { <-s.sema }()
		err := f()
		if err != nil {
			s.mu.Lock()
			if s.err == nil {
				s.err = err
			}
			s.mu.Unlock()
		}
	}()
}

// Wait until all goroutines have finished running. If a goroutine returns a
// non-nil error, the first non-nil error recorded will be returned.
func (s *Bounded) Wait() error {
	for i := 0; i < cap(s.sema); i++ {
		s.sema <- struct{}{}
	}

	s.mu.Lock()
	err := s.err
	s.mu.Unlock()

	return err
}
