package main

import "sync"

type semaphore struct {
	sema chan bool
	wg   sync.WaitGroup
}

func newSemaphore(size int) *semaphore {
	return &semaphore{
		sema: make(chan bool, size),
	}
}

func (s *semaphore) Acquire() {
	s.sema <- true
	s.wg.Add(1)
}

func (s *semaphore) Release() {
	<-s.sema
	s.wg.Done()
}

func (s *semaphore) Wait() {
	s.wg.Wait()
}
