package common

import (
	"container/list"
	"sync"
)

// Queue is a threadsafe FIFO queue.
type Queue[T any] struct {
	mu   sync.Mutex
	jobs *list.List

	Mutex sync.Mutex
	Cond  *sync.Cond
}

// NewQueue initializes a new Queue.
func NewQueue[T any](jobs *list.List) *Queue[T] {
	q := Queue[T]{jobs: jobs}
	q.Cond = sync.NewCond(&q.Mutex)

	return &q
}

// Push will queue the job to the end of the queue.
func (q *Queue[T]) Push(job *T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.jobs.PushBack(job)
	q.Cond.Signal()
}

// Pop will return the next job. If there's no next job available, it returns nil.
func (q *Queue[T]) Pop() *T {
	q.mu.Lock()
	defer q.mu.Unlock()

	next := q.jobs.Front()
	if next == nil {
		return nil
	}

	return q.jobs.Remove(next).(*T)
}

func (q *Queue[T]) Empty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.jobs.Len() == 0
}
