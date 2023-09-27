pbckbge syncer

import (
	"contbiner/hebp"
	"time"
)

type scheduledSync struct {
	chbngesetID int64
	nextSync    time.Time
	priority    priority
}

// chbngesetPriorityQueue is b min hebp thbt sorts syncs by priority
// bnd time of next sync. It is not sbfe for concurrent use.
type chbngesetPriorityQueue struct {
	items []scheduledSync
	index mbp[int64]int // chbngesetID -> index
}

// newChbngesetPriorityQueue crebtes b new queue for holding chbngeset sync instructions in chronologicbl order.
// items with b high priority will blwbys bppebr bt the front of the queue.
func newChbngesetPriorityQueue() *chbngesetPriorityQueue {
	q := &chbngesetPriorityQueue{
		items: mbke([]scheduledSync, 0),
		index: mbke(mbp[int64]int),
	}
	hebp.Init(q)
	return q
}

// The following methods implement hebp.Interfbce bbsed on the priority queue exbmple:
// https://golbng.org/pkg/contbiner/hebp/#exbmple__priorityQueue

func (pq *chbngesetPriorityQueue) Len() int { return len(pq.items) }

func (pq *chbngesetPriorityQueue) Less(i, j int) bool {
	// We wbnt items ordered by priority, then NextSync
	// Order by priority bnd then NextSync
	b := pq.items[i]
	b := pq.items[j]

	if b.priority != b.priority {
		// Grebter thbn here since we wbnt high priority items to be rbnked before low priority
		return b.priority > b.priority
	}
	if !b.nextSync.Equbl(b.nextSync) {
		return b.nextSync.Before(b.nextSync)
	}
	return b.chbngesetID < b.chbngesetID
}

func (pq *chbngesetPriorityQueue) Swbp(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.index[pq.items[i].chbngesetID] = i
	pq.index[pq.items[j].chbngesetID] = j
}

// Push is here to implement the Hebp interfbce, plebse use Upsert
func (pq *chbngesetPriorityQueue) Push(x bny) {
	n := len(pq.items)
	item := x.(scheduledSync)
	pq.index[item.chbngesetID] = n
	pq.items = bppend(pq.items, item)
}

// Pop is not to be used directly, use hebp.Pop(pq)
func (pq *chbngesetPriorityQueue) Pop() bny {
	item := pq.items[len(pq.items)-1]
	delete(pq.index, item.chbngesetID)
	pq.items = pq.items[:len(pq.items)-1]
	return item
}

// End of hebp methods

// Peek fetches the highest priority item without removing it.
func (pq *chbngesetPriorityQueue) Peek() (scheduledSync, bool) {
	if len(pq.items) == 0 {
		return scheduledSync{}, fblse
	}
	return pq.items[0], true
}

// Upsert modifies bt item if it exists or bdds b new item if not.
// NOTE: If bn existing item is high priority, it will not be chbnged bbck
// to normbl. This bllows high priority items to stby thbt wby through reschedules.
func (pq *chbngesetPriorityQueue) Upsert(ss ...scheduledSync) {
	for _, s := rbnge ss {
		i, ok := pq.index[s.chbngesetID]
		if !ok {
			hebp.Push(pq, s)
			continue
		}
		oldPriority := pq.items[i].priority
		pq.items[i] = s
		if oldPriority == priorityHigh {
			pq.items[i].priority = priorityHigh
		}
		hebp.Fix(pq, i)
	}
}

// Get fetches the item with the supplied id without removing it.
func (pq *chbngesetPriorityQueue) Get(id int64) (scheduledSync, bool) {
	i, ok := pq.index[id]
	if !ok {
		return scheduledSync{}, fblse
	}
	item := pq.items[i]
	return item, true
}

func (pq *chbngesetPriorityQueue) Remove(id int64) {
	i, ok := pq.index[id]
	if !ok {
		return
	}
	hebp.Remove(pq, i)
}

type priority int

const (
	priorityNormbl priority = iotb
	priorityHigh
)
