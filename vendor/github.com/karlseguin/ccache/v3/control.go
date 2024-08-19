package ccache

type controlGC struct {
	done chan struct{}
}

type controlClear struct {
	done chan struct{}
}

type controlStop struct {
}

type controlGetSize struct {
	res chan int64
}

type controlGetDropped struct {
	res chan int
}

type controlSetMaxSize struct {
	size int64
	done chan struct{}
}

type controlSyncUpdates struct {
	done chan struct{}
}

type control chan interface{}

func newControl() chan interface{} {
	return make(chan interface{}, 5)
}

// Forces GC. There should be no reason to call this function, except from tests
// which require synchronous GC.
// This is a control command.
func (c control) GC() {
	done := make(chan struct{})
	c <- controlGC{done: done}
	<-done
}

// Sends a stop signal to the worker thread. The worker thread will shut down
// 5 seconds after the last message is received. The cache should not be used
// after Stop is called, but concurrently executing requests should properly finish
// executing.
// This is a control command.
func (c control) Stop() {
	c.SyncUpdates()
	c <- controlStop{}
}

// Clears the cache
// This is a control command.
func (c control) Clear() {
	done := make(chan struct{})
	c <- controlClear{done: done}
	<-done
}

// Gets the size of the cache. This is an O(1) call to make, but it is handled
// by the worker goroutine. It's meant to be called periodically for metrics, or
// from tests.
// This is a control command.
func (c control) GetSize() int64 {
	res := make(chan int64)
	c <- controlGetSize{res: res}
	return <-res
}

// Gets the number of items removed from the cache due to memory pressure since
// the last time GetDropped was called
// This is a control command.
func (c control) GetDropped() int {
	res := make(chan int)
	c <- controlGetDropped{res: res}
	return <-res
}

// Sets a new max size. That can result in a GC being run if the new maxium size
// is smaller than the cached size
// This is a control command.
func (c control) SetMaxSize(size int64) {
	done := make(chan struct{})
	c <- controlSetMaxSize{size: size, done: done}
	<-done
}

// SyncUpdates waits until the cache has finished asynchronous state updates for any operations
// that were done by the current goroutine up to now.
//
// For efficiency, the cache's implementation of LRU behavior is partly managed by a worker
// goroutine that updates its internal data structures asynchronously. This means that the
// cache's state in terms of (for instance) eviction of LRU items is only eventually consistent;
// there is no guarantee that it happens before a Get or Set call has returned. Most of the time
// application code will not care about this, but especially in a test scenario you may want to
// be able to know when the worker has caught up.
//
// This applies only to cache methods that were previously called by the same goroutine that is
// now calling SyncUpdates. If other goroutines are using the cache at the same time, there is
// no way to know whether any of them still have pending state updates when SyncUpdates returns.
// This is a control command.
func (c control) SyncUpdates() {
	done := make(chan struct{})
	c <- controlSyncUpdates{done: done}
	<-done
}
