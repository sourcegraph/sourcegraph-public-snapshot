package group

import (
	"sync"
)

// NewParallelOrdered creates an parallel task-running group of goroutines that
// guarantees that the callback will be called in the same order as the tasks
// are submitted. This is useful when you have a stream of tasks that you'd
// like to process, but the processing is expensive and worth parallelizing.
//
// It is safe to call Submit() from multiple goroutines, but Submit() should
// never be called after calling Done().
func NewParallelOrdered[T any](maxParallel int, callback func(T)) *parallelOrdered[T] {
	return &parallelOrdered[T]{
		cb:          callback,
		tasks:       make(chan task[T], 16),
		results:     make(chan chan T, 16),
		maxParallel: maxParallel,
	}
}

type parallelOrdered[T any] struct {
	// The callback that will be called with the results of a submitted task
	cb func(T)

	// A queue of tasks ready to be executed
	tasks chan task[T]

	// A queue of channels that will yield the result of a submitted task.
	// This queue will always yield channels in the same order as the tasks
	// were submitted.
	results chan chan T

	maxParallel int

	poolOnce             sync.Once
	runnerWg, callbackWg sync.WaitGroup
}

type task[T any] struct {
	f      func() T
	submit chan T
}

// Submit queues a task to be executed by the group's worker pool.
// When the submitted function completes, the provided callback function
// will be called with the task's result. The callback will be called
// in the same order that Submit was called even though the tasks are run
// in parallel.
// Submit must not be called after Done().
func (g *parallelOrdered[T]) Submit(f func() T) {
	g.startPoolOnce()

	resultChan := make(chan T, 1)
	g.results <- resultChan
	g.tasks <- task[T]{f: f, submit: resultChan}
}

// Done is called to clean up any goroutines started by the group.
func (g *parallelOrdered[T]) Done() {
	close(g.tasks)
	g.runnerWg.Wait()
	close(g.results)
	g.callbackWg.Wait()
}

func (g *parallelOrdered[T]) startPoolOnce() {
	g.poolOnce.Do(func() {
		g.callbackWg.Add(1)
		go func() {
			g.callbackWithResults()
			g.callbackWg.Done()
		}()

		for i := 0; i < g.maxParallel; i++ {
			g.runnerWg.Add(1)
			go func() {
				g.runTasks()
				g.runnerWg.Done()
			}()
		}
	})
}

func (g *parallelOrdered[T]) runTasks() {
	for task := range g.tasks {
		task.submit <- task.f()
	}
}

func (g *parallelOrdered[T]) callbackWithResults() {
	for resultChan := range g.results {
		g.cb(<-resultChan)
	}
}
