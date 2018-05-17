package main

import (
	"container/heap"
	"sync"
)

type queueItem struct {
	// repoName is the name of the repo
	repoName string
	// indexedCommit is the last known indexed commit
	indexedCommit string
	// latestCommit is the latest commit available from gitserver. It is the
	// commit want to index next. It can be the same as indexedCommit.
	latestCommit string
	// heapIdx is the index of the item in the heap. If < 0 then the item is
	// not on the heap.
	heapIdx int
	// seq is a sequence number used as a tie breaker. This is to ensure we
	// act like a FIFO queue.
	seq int64
}

// Queue is a priority queue which returns the next repo to index. It is safe
// to use concurrently. It is a min queue on:
//
//    (indexed commit != latest commit, time added to the queue)
//
// We use the above since:
//
// * We rather index a repo sooner if we know the commit is stale.
// * The order of repos returned by Sourcegraph API are ordered by importance.
type Queue struct {
	mu    sync.Mutex
	items map[string]*queueItem
	pq    pqueue
	seq   int64
}

// Pop returns the repoName and commit of the next repo to index. If the queue
// is empty ok is false.
func (q *Queue) Pop() (repoName string, commit string, ok bool) {
	q.mu.Lock()
	if len(q.pq) == 0 {
		q.mu.Unlock()
		return "", "", false
	}
	item := heap.Pop(&q.pq).(*queueItem)
	repoName = item.repoName
	commit = item.latestCommit
	q.mu.Unlock()
	return repoName, commit, true
}

// Len returns the number of items in the queue.
func (q *Queue) Len() int {
	q.mu.Lock()
	l := len(q.pq)
	q.mu.Unlock()
	return l
}

// AddOrUpdate sets which commit to index next for repoName. If repoName is
// already in the queue, it is updated.
func (q *Queue) AddOrUpdate(repoName, commit string) {
	q.mu.Lock()
	item := q.get(repoName)
	item.latestCommit = commit
	if item.heapIdx < 0 {
		q.seq++
		item.seq = q.seq
		heap.Push(&q.pq, item)
	} else {
		heap.Fix(&q.pq, item.heapIdx)
	}
	q.mu.Unlock()
}

// SetIndexed sets what the currently indexed commit is for repoName.
func (q *Queue) SetIndexed(repoName, indexed string) {
	q.mu.Lock()
	item := q.get(repoName)
	item.indexedCommit = indexed
	if item.heapIdx >= 0 {
		// We only update the position in the queue, never add it.
		heap.Fix(&q.pq, item.heapIdx)
	}
	q.mu.Unlock()
}

// get returns the item for repoName. If the repoName hasn't been seen before,
// it is added to q.items.
//
// Note: get requires that q.mu is held.
func (q *Queue) get(repoName string) *queueItem {
	if q.items == nil {
		q.items = map[string]*queueItem{}
		q.pq = make(pqueue, 0)
	}

	item, ok := q.items[repoName]
	if !ok {
		item = &queueItem{
			repoName: repoName,
			heapIdx:  -1,
		}
		q.items[repoName] = item
	}

	return item
}

// pqueue implements a priority queue via the interface for container/heap
type pqueue []*queueItem

func (pq pqueue) Len() int { return len(pq) }

func (pq pqueue) Less(i, j int) bool {
	// If we know a needs an update and b doesn't, then return true. Otherwise
	// they are either equal priority or b is more urgent.
	x := pq[i]
	y := pq[j]
	if (x.indexedCommit == x.latestCommit) == (y.indexedCommit == y.latestCommit) {
		// tie breaker is to prefer the item added to the queue first
		return x.seq < y.seq
	}
	return x.indexedCommit != x.latestCommit
}

func (pq pqueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].heapIdx = i
	pq[j].heapIdx = j
}

func (pq *pqueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*queueItem)
	item.heapIdx = n
	*pq = append(*pq, item)
}

func (pq *pqueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.heapIdx = -1
	*pq = old[0 : n-1]
	return item
}
