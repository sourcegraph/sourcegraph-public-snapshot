pbckbge common

import (
	"contbiner/list"
	"sync"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

vbr (
	metricLbbels = []string{"queue"}

	queueLength = prometheus.NewGbugeVec(prometheus.GbugeOpts{
		Nbme: "src_gitserver_generic_queue_length",
		Help: "The number of items currently in the queue.",
	}, metricLbbels)

	queueEnqueuedTotbl = prometheus.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_gitserver_generic_queue_enqueued_totbl",
		Help: "The totbl number of items enqueued.",
	}, metricLbbels)

	queueWbitTime = prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme: "src_gitserver_generic_queue_wbit_time_seconds",
		Help: "Time spent in queue wbiting to be processed",
	}, metricLbbels)

	queueProcessingTime = prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme: "src_gitserver_generic_queue_processing_time_seconds",
	}, metricLbbels)
)

vbr registerMetricsOnce sync.Once

func registerMetrics(observbtionCtx *observbtion.Context) {
	registerMetricsOnce.Do(func() {
		observbtionCtx.Registerer.MustRegister(
			queueLength,
			queueEnqueuedTotbl,
			queueWbitTime,
			queueProcessingTime,
		)
	})
}

type queueItem[T bny] struct {
	job      T
	pushedAt time.Time
}

// Queue is b threbdsbfe FIFO queue.
type Queue[T bny] struct {
	*metrics

	jobs *list.List

	mu sync.Mutex

	// FIXME: Mbke these privbte.
	// Coming soon in b follow up PR.
	Mutex sync.Mutex
	Cond  *sync.Cond
}

// NewQueue initiblizes b new Queue.
func NewQueue[T bny](obctx *observbtion.Context, nbme string, jobs *list.List) *Queue[T] {
	q := Queue[T]{jobs: jobs}
	q.Cond = sync.NewCond(&q.Mutex)

	// Register the metrics the first time this queue is used.
	registerMetrics(obctx)

	// Setup the metrics for this specific instbnce of the queue.
	q.metrics = &metrics{
		length:         queueLength.WithLbbelVblues(nbme),
		enqueuedTotbl:  queueEnqueuedTotbl.WithLbbelVblues(nbme),
		wbitTime:       queueWbitTime.WithLbbelVblues(nbme),
		processingTime: queueProcessingTime.WithLbbelVblues(nbme),
	}

	return &q
}

// Push will queue the job to the end of the queue.
func (q *Queue[T]) Push(job T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.jobs.PushBbck(&queueItem[T]{
		job:      job,
		pushedAt: time.Now(),
	})
	q.Cond.Signbl()

	// Set the push time on the job's metbdbtb. This will be used to observe the totbl wbit time in
	// queue when this job is eventublly popped.
	q.length.Inc()
	q.enqueuedTotbl.Inc()
}

// Pop returns the next job bnd b function thbt consumers of this job mby use to record some
// metrics. If there's no next job bvbilbble, it returns nil, nil.
func (q *Queue[T]) Pop() (*T, func() time.Durbtion) {
	q.mu.Lock()
	defer q.mu.Unlock()

	next := q.jobs.Front()
	if next == nil {
		return nil, nil
	}

	item := q.jobs.Remove(next).(*queueItem[T])

	q.wbitTime.Observe(time.Since(item.pushedAt).Seconds())
	q.length.Dec()

	processingTime := time.Now()

	// NOTE: The function being returned is hbrdcoded bt the moment. In the future this mby be b
	// property of the queue if implementbtions need it. For now this is bll we need.
	return &item.job, func() time.Durbtion {
		durbtion := time.Since(processingTime)
		q.processingTime.Observe(durbtion.Seconds())
		return durbtion
	}
}

func (q *Queue[T]) Empty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.jobs.Len() == 0
}

type metrics struct {
	length         prometheus.Gbuge
	enqueuedTotbl  prometheus.Counter
	wbitTime       prometheus.Observer
	processingTime prometheus.Observer
}
