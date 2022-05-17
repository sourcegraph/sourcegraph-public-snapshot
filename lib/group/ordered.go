package group

// NewParallelOrdered creates a workerpool that runs tasks in parallel, but
// guarantees that the callback will be called in the same order as the tasks
// are submitted. This is useful when you have a stream of tasks that you'd
// like to process, but the processing is expensive and worth parallelizing.
//
// It is safe to call Submit() from multiple goroutines, but Submit() should
// never be called after calling Done().
func NewParallelOrdered[T any](maxParallel int, callback func(T)) *parallelOrdered[T] {
	return &parallelOrdered[T]{
		callback:    callback,
		tasks:       make(chan task[T]),
		results:     make(chan chan T, maxParallel),
		callbackerC: make(chan struct{}, 1),
		workerpoolC: make(chan struct{}, maxParallel),
	}
}

type parallelOrdered[T any] struct {
	// The callback that will be called with the results of a submitted task
	callback func(T)

	// A channel to send submitted tasks to workers. It is of size 0 so we can
	// use a blocked send as a signal that we should start more workers.
	tasks chan task[T]

	// A queue of channels that will yield the result of a submitted task.
	// This queue will always yield channels in the same order as the tasks
	// were submitted.
	results chan chan T

	// callbackerC is capacity-1 channel that acts as a semaphore and
	// waitgroup for the single callbacker goroutine. An empty channel
	// indicates there is no callbacker running.
	callbackerC chan struct{}

	// workerpoolC is a channel with capacity maxParallel that acts as
	// a semaphore and waitgroup for the worker pool. An empty channel
	// signals no workers are running.
	workerpoolC chan struct{}
}

type task[T any] struct {
	f       func() T
	resultC chan T
}

// Submit queues a task to be executed by the group's worker pool.
// When the submitted function completes, the provided callback function
// will be called with the task's result. The callback will be called
// in the same order that Submit was called even though the tasks are run
// in parallel.
// Submit must not be called after Done().
func (g *parallelOrdered[T]) Submit(f func() T) {
	g.startCallbackerOnce()

	resultC := make(chan T, 1)
	g.results <- resultC

	for {
		select {
		case g.tasks <- task[T]{f, resultC}:
			// If a worker is ready to execute the task, send it and
			// don't start any new workers.
			return
		case g.workerpoolC <- struct{}{}:
			// If we can't immediately submit our task, but our workerpool
			// isn't full, start a worker and try again.
			go func() {
				g.worker()
				<-g.workerpoolC
			}()
		}
	}
}

// Done is called to clean up any goroutines started by the group.
func (g *parallelOrdered[T]) Done() {
	close(g.tasks)
	for i := 0; i < cap(g.workerpoolC); i++ {
		// filling workerpoolC means all the workers have exited
		g.workerpoolC <- struct{}{}
	}
	close(g.results)
	g.callbackerC <- struct{}{}
}

func (g *parallelOrdered[T]) startCallbackerOnce() {
	select {
	case g.callbackerC <- struct{}{}:
		go func() {
			for resultChan := range g.results {
				g.callback(<-resultChan)
			}
			<-g.callbackerC
		}()
	default:
		// if callbackerC is already full, do nothing
	}
}

func (g *parallelOrdered[T]) worker() {
	for task := range g.tasks {
		task.resultC <- task.f()
	}
}
