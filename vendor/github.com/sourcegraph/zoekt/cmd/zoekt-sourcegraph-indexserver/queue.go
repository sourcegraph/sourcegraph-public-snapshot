package main

import (
	"bytes"
	"container/heap"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	sglog "github.com/sourcegraph/log"
)

type queueItem struct {
	// repoID is the ID of the repo
	repoID uint32
	// opts are the options to use when indexing repoID.
	opts IndexOptions
	// indexed is true if opts has been indexed.
	indexed bool
	// indexState is the indexState of the last attempt at indexing repoID.
	indexState indexState
	// heapIdx is the index of the item in the heap. If < 0 then the item is
	// not on the heap.
	heapIdx int
	// seq is a sequence number used as a tiebreaker. This is to ensure we
	// act like a FIFO queue.
	seq int64
	// dateAddedToQueue is the time when this indexing job was added to the queue. If this item is no longer
	// in the heap (i.e. it has been processed already), this value is nonsensical.
	dateAddedToQueue time.Time
	// backoff will handle backing off of future indexing requests for a duration of time based on previous failures
	backoff backoff
}

// Queue is a priority queue which returns the next repo to index. It is safe
// to use concurrently. It is a min queue on:
//
//	(!indexed, time added to the queue)
//
// We use the above since we'd rather index a repo sooner if we know the commit is stale.
type Queue struct {
	mu           sync.Mutex
	items        map[uint32]*queueItem
	pq           pqueue
	seq          int64
	logger       sglog.Logger
	newQueueItem func(uint32) *queueItem
}

func NewQueue(backoffDuration, maxBackoffDuration time.Duration, l sglog.Logger) *Queue {
	if backoffDuration < 0 || maxBackoffDuration < 0 {
		backoffDuration = 0
		maxBackoffDuration = 0
	}
	newQueueItem := func(repoID uint32) *queueItem {
		return &queueItem{
			repoID:  repoID,
			heapIdx: -1,
			backoff: backoff{
				backoffDuration: backoffDuration,
				maxBackoff:      maxBackoffDuration,
			},
		}
	}

	q := &Queue{
		newQueueItem: newQueueItem,
		logger:       l.Scoped("queue"),
	}

	q.init()
	return q
}

type QueueItem struct {
	// Opts are the options to use when indexing the repo.
	Opts IndexOptions
	// DateAddedToQueue is the time when this indexing job was added to the queue, used for telemetry.
	DateAddedToQueue time.Time
}

// Pop returns options and metadata for the next repo to index. If the queue is empty ok is false.
func (q *Queue) Pop() (result QueueItem, ok bool) {
	q.mu.Lock()
	if len(q.pq) == 0 {
		q.mu.Unlock()
		return QueueItem{}, false
	}

	item := heap.Pop(&q.pq).(*queueItem)

	metricQueueLen.Set(float64(len(q.pq)))
	metricQueueCap.Set(float64(len(q.items)))

	dateAdded := item.dateAddedToQueue
	item.dateAddedToQueue = time.Unix(0, 0)

	q.mu.Unlock()

	name := repoNameForMetric(item.opts.Name)
	age := time.Since(dateAdded)
	metricQueueAge.WithLabelValues(name).Observe(age.Seconds())

	return QueueItem{item.opts, dateAdded}, true
}

// Len returns the number of items in the queue.
func (q *Queue) Len() int {
	q.mu.Lock()
	l := len(q.pq)
	q.mu.Unlock()
	return l
}

// AddOrUpdate sets which opts to index next. If opts.RepoID is already in the
// queue, it is updated.
func (q *Queue) AddOrUpdate(opts IndexOptions) {
	q.mu.Lock()
	item := q.getOrAdd(opts.RepoID)
	if !reflect.DeepEqual(item.opts, opts) {
		item.indexed = false
		item.opts = opts
	}
	if item.heapIdx < 0 {
		if item.backoff.Allow(time.Now()) {
			q.seq++
			item.seq = q.seq
			item.dateAddedToQueue = time.Now()

			heap.Push(&q.pq, item)
			metricQueueLen.Set(float64(len(q.pq)))
			metricQueueCap.Set(float64(len(q.items)))
		}
	} else {
		heap.Fix(&q.pq, item.heapIdx)
	}
	q.mu.Unlock()
}

// Bump will take any repository in ids which is not on the queue and
// re-insert it with the last known IndexOptions. Bump returns ids that are
// unknown to the queue.
func (q *Queue) Bump(ids []uint32) []uint32 {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.items == nil {
		q.init()
	}

	var missing []uint32
	for _, id := range ids {
		item, ok := q.items[id]
		if !ok {
			missing = append(missing, id)
		} else if item.heapIdx < 0 {
			if item.backoff.Allow(time.Now()) {
				q.seq++
				item.seq = q.seq
				item.dateAddedToQueue = time.Now()

				heap.Push(&q.pq, item)
				metricQueueLen.Set(float64(len(q.pq)))
				metricQueueCap.Set(float64(len(q.items)))
			}
		}
	}

	return missing
}

// Iterate will call f on each item known to the queue, including items that
// have been popped from the queue. Note: this is done in a random order and
// the queue mutex is held during all calls to f. Do not mutate the data.
func (q *Queue) Iterate(f func(*IndexOptions)) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, item := range q.items {
		f(&item.opts)
	}
}

// debugIteratedOrdered will call f on each queueItem (sorted by indexing priority)
// known to the queue, including queueItems that have been popped from the queue).
//
// Note: The mutex is held during all calls to f. Callers must not modify the queueItems.
func (q *Queue) debugIteratedOrdered(f func(*queueItem)) {
	q.mu.Lock()
	defer q.mu.Unlock()

	queueItems := make([]*queueItem, 0, len(q.items))
	for _, item := range q.items {
		queueItems = append(queueItems, item)
	}

	sort.Slice(queueItems, func(i, j int) bool {
		x, y := queueItems[i], queueItems[j]

		xOnQueue, yOnQueue := x.heapIdx >= 0, y.heapIdx >= 0
		if xOnQueue != yOnQueue {
			return xOnQueue
		}

		return lessQueueItemPriority(x, y)
	})

	for _, item := range queueItems {
		f(item)
	}
}

func (q *Queue) handleDebugQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method must be GET", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	var bufferedWriter bytes.Buffer

	writer := tabwriter.NewWriter(&bufferedWriter, 16, 8, 4, ' ', 0)

	_, err := fmt.Fprintf(writer, "Position\tName\tID\tIsOnQueue\tAge\tBranches\t\n")
	if err != nil {
		http.Error(w, fmt.Sprintf("writing column headers: %s", err), http.StatusInternalServerError)
		return
	}

	now := time.Now()

	position := -1
	q.debugIteratedOrdered(func(item *queueItem) {
		position++

		if err != nil {
			return
		}

		var branches []string
		for _, b := range item.opts.Branches {
			branches = append(branches, b.String())
		}

		isOnQueue := item.heapIdx >= 0
		age := "-"
		if isOnQueue {
			age = now.Sub(item.dateAddedToQueue).Round(time.Second).String()
		}

		_, err = fmt.Fprintf(writer, "%d\t%s\t%d\t%t\t%s\t%s\n", position, item.opts.Name, item.repoID, isOnQueue, age, strings.Join(branches, ", "))
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("writing queueItem: %s", err), http.StatusInternalServerError)
		return
	}

	err = writer.Flush()
	if err != nil {
		http.Error(w, fmt.Sprintf("flushing tabwriter: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Length", strconv.Itoa(bufferedWriter.Len()))

	_, err = io.Copy(w, &bufferedWriter)
	if err != nil {
		http.Error(w, fmt.Sprintf("copying output to response writer: %s", err), http.StatusInternalServerError)
		return
	}
}

// SetIndexed sets what the currently indexed options are for opts.RepoID.
func (q *Queue) SetIndexed(opts IndexOptions, state indexState) {
	q.mu.Lock()
	item := q.getOrAdd(opts.RepoID)

	item.indexState = state
	if state != indexStateFail {
		item.indexed = reflect.DeepEqual(opts, item.opts)
		item.backoff.Reset()

		if item.heapIdx >= 0 {
			// We only update the position in the queue, never add it.
			heap.Fix(&q.pq, item.heapIdx)
		}
	} else {
		item.backoff.Fail(time.Now(), q.logger, item.opts)

		if item.heapIdx >= 0 {
			// Remove from queue
			heap.Remove(&q.pq, item.heapIdx)
			item.heapIdx = -1
		}
	}

	q.mu.Unlock()
}

// MaybeRemoveMissing will remove all queue items not in ids and return the
// ids of items removed from the queue. It will heuristically not run to
// conserve resources.
//
// In the server's steady state we expect that the list of names is equal to
// the items in queue. As such in the steady state this function should do no
// removals. Removal requires memory allocation and coarse locking. To avoid
// that we use a heuristic which can falsely decide it doesn't need to
// remove. However, we will converge onto removing items.
func (q *Queue) MaybeRemoveMissing(ids []uint32) []uint32 {
	q.mu.Lock()
	sameSize := len(q.items) == len(ids)
	q.mu.Unlock()

	// heuristically skip expensive work
	if sameSize {
		debug.Printf("skipping MaybeRemoveMissing due to same size: %d", len(ids))
		return nil
	}

	set := make(map[uint32]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	var removed []uint32
	for _, item := range q.items {
		if _, ok := set[item.opts.RepoID]; ok {
			continue
		}

		if item.heapIdx >= 0 {
			heap.Remove(&q.pq, item.heapIdx)
		}

		item.indexState = ""

		delete(q.items, item.opts.RepoID)

		removed = append(removed, item.opts.RepoID)
	}

	metricQueueLen.Set(float64(len(q.pq)))
	metricQueueCap.Set(float64(len(q.items)))

	return removed
}

// getOrAdd returns the item for repoID. If the repoID hasn't been seen before, it
// is added to q.items.
//
// Note: getOrAdd requires that q.mu is held.
func (q *Queue) getOrAdd(repoID uint32) *queueItem {
	if q.items == nil {
		q.init()
	}

	item, ok := q.items[repoID]
	if !ok {
		item = q.newQueueItem(repoID)
		q.items[repoID] = item
	}

	return item
}

// get returns the item for repoID.
//
// Note: get requires that q.mu is held.
func (q *Queue) get(repoID uint32) *queueItem {
	return q.items[repoID]
}

func (q *Queue) init() {
	q.items = map[uint32]*queueItem{}
	q.pq = make(pqueue, 0)
}

// pqueue implements a priority queue via the interface for container/heap
type pqueue []*queueItem

func (pq pqueue) Len() int { return len(pq) }

func (pq pqueue) Less(i, j int) bool {
	x := pq[i]
	y := pq[j]
	return lessQueueItemPriority(x, y)
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

// lessQueueItemPriority returns true if indexing x should be prioritized over indexing y
func lessQueueItemPriority(x, y *queueItem) bool {
	// If we know x needs an update and y doesn't, then return true. Otherwise
	// they are either equal priority or y is more urgent.
	if x.indexed != y.indexed {
		return !x.indexed
	}

	if xFail, yFail := x.indexState == indexStateFail, y.indexState == indexStateFail; xFail != yFail {
		// if you failed to index, you are likely to fail again. So prefer
		// non-failed.
		return !xFail
	}

	// tiebreaker is to prefer the item added to the queue first
	return x.seq < y.seq
}

var (
	metricQueueLen = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "index_queue_len",
		Help: "The number of repositories in the index queue.",
	})
	metricQueueCap = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "index_queue_cap",
		Help: "The number of repositories tracked by the index queue, including popped items. Should be the same as index_num_assigned.",
	})
	metricQueueAge = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "index_queue_age_seconds",
		Help: "A histogram of the amount of time a popped repository spent sitting in the queue beforehand.",
		Buckets: []float64{
			60,     // 1m
			300,    // 5m
			1200,   // 20m
			2400,   // 40m
			3600,   // 1h
			10800,  // 3h
			18000,  // 5h
			36000,  // 10h
			43200,  // 12h
			54000,  // 15h
			72000,  // 20h
			86400,  // 24h
			108000, // 30h
			126000, // 35h
			172800, // 48h
		},
	}, []string{"name"}) // name=name of the repository that was just popped from the queue
)
