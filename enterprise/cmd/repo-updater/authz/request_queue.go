package authz

import (
	"container/heap"
	"sync"
)

// Priority defines how urgent the permissions syncing request is.
// Generally, if the request is driven from a user action (e.g. sign up, log in)
// then it should be PriorityHigh. All other cases should be PriorityLow.
type Priority int

const (
	PriorityLow Priority = iota
	PriorityHigh
)

// requestType is the type of the permissions syncing request. It defines the
// permissions syncing is either repository-centric or user-centric.
type requestType int

const (
	requestTypeUnknown requestType = iota
	requestTypeRepo
	requestTypeUser
)

// syncRequest is a permissions syncing request with its current status in the queue.
type syncRequest struct {
	typ      requestType
	id       int32
	priority Priority

	updating bool // Whether the request has been acquired
	index    int  // The index in the heap
}

// requestQueue is a priority queue of permissions syncing requests.
// A request can't have more than one location in the queue.
type requestQueue struct {
	mu    sync.Mutex
	heap  []*syncRequest
	index map[requestType]map[int32]*syncRequest

	// The queue performs a non-blocking send on this channel
	// when a new value is enqueued so that the update loop
	// can wake up if it is idle.
	notifyEnqueue chan struct{}
}

func newRequestQueue() *requestQueue {
	q := &requestQueue{
		index: make(map[requestType]map[int32]*syncRequest),
	}

	for _, typ := range []requestType{
		requestTypeRepo,
		requestTypeUser,
	} {
		q.index[typ] = make(map[int32]*syncRequest)
	}
	return q
}

// notify performs a non-blocking send on the channel.
// The channel should be buffered.
var notify = func(ch chan struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}

// enqueue adds a sync request to the queue with the given information.
//
// If the sync request is already in the queue and it isn't yet updating,
// the request is updated.
//
// If the given priority is higher than the one in the queue,
// the sync request's position in the queue is updated accordingly.
func (q *requestQueue) enqueue(typ requestType, id int32, priority Priority) (updated bool) {
	if id == 0 {
		return false
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	request := q.index[typ][id]
	if request == nil {
		heap.Push(q, &syncRequest{
			typ:      typ,
			id:       id,
			priority: priority,
		})
		notify(q.notifyEnqueue)
		return false
	}

	if request.updating {
		return false
	}

	if request.priority >= priority {
		// Request is already in the queue with at least as good priority.
		return false
	}

	request.priority = priority
	heap.Fix(q, request.index)
	notify(q.notifyEnqueue)
	return true
}

// remove removes the sync request from the queue if the request.Updating matches the
// updating argument.
func (q *requestQueue) remove(typ requestType, id int32, updating bool) (removed bool) {
	if id == 0 {
		return false
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	request := q.index[typ][id]
	if request != nil && request.updating == updating {
		heap.Remove(q, request.index)
		return true
	}

	return false
}

// acquireNext acquires the next sync request. The acquired request must be removed from
// the queue when the request finishes (independent of success or failure).
func (q *requestQueue) acquireNext() *syncRequest {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.Len() == 0 {
		return nil
	}

	request := q.heap[0]
	if request.updating {
		// Everything in the queue is already updating.
		return nil
	}

	request.updating = true
	heap.Fix(q, request.index)
	return request
}

// The following methods implement heap.Interface based on the priority queue example:
// https://golang.org/pkg/container/heap/#example__priorityQueue
// These methods are not safe for concurrent use. Therefore, it is the caller's
// responsibility to ensure they're being guarded by a mutex during any heap operation,
// i.e. heap.Fix, heap.Remove, heap.Push, heap.Pop.

func (q *requestQueue) Len() int { return len(q.heap) }

func (q *requestQueue) Less(i, j int) bool {
	qi := q.heap[i]
	qj := q.heap[j]

	if qi.updating != qj.updating {
		// Requests that are already updating are sorted last.
		return qj.updating
	}

	if qi.priority != qj.priority {
		// We want Pop to give us the highest, not lowest, priority so we use greater than here.
		return qi.priority > qj.priority
	}

	if qi.typ != qj.typ {
		// requestTypeUser > requestTypeRepo
		return qi.typ > qj.typ
	}

	return false
}

func (q *requestQueue) Swap(i, j int) {
	q.heap[i], q.heap[j] = q.heap[j], q.heap[i]
	q.heap[i].index = i
	q.heap[j].index = j
}

func (q *requestQueue) Push(x interface{}) {
	n := len(q.heap)
	request := x.(*syncRequest)
	request.index = n
	q.heap = append(q.heap, request)
	q.index[request.typ][request.id] = request
}

func (q *requestQueue) Pop() interface{} {
	n := len(q.heap)
	request := q.heap[n-1]
	request.index = -1 // for safety
	q.heap = q.heap[0 : n-1]
	delete(q.index[request.typ], request.id)
	return request
}
