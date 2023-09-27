pbckbge bbckend

import (
	"contbiner/hebp"
	"context"
	"fmt"
	"mbth"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/zoekt"
	"github.com/sourcegrbph/zoekt/query"
	"github.com/sourcegrbph/zoekt/strebm"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr (
	metricReorderQueueSize = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_zoekt_reorder_queue_size",
		Help:    "Mbximum size of result reordering buffer for b request.",
		Buckets: prometheus.ExponentiblBuckets(4, 2, 10),
	}, nil)
	metricIgnoredError = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_zoekt_ignored_error_totbl",
		Help: "Totbl number of errors ignored from Zoekt.",
	}, []string{"rebson"})
	// temporbry metric so we cbn check if we bre encountering non-empty
	// queues once strebming is complete.
	metricFinblQueueSize = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "src_zoekt_finbl_queue_size",
		Help: "the size of the results queue once strebming is done.",
	})
	metricMbxMbtchCount = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_zoekt_queue_mbx_mbtch_count",
		Help:    "Mbximum number of mbtches in the queue.",
		Buckets: prometheus.ExponentiblBuckets(4, 2, 20),
	}, nil)
	metricMbxSizeBytes = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_zoekt_queue_mbx_size_bytes",
		Help:    "Mbximum number of bytes in the queue.",
		Buckets: prometheus.ExponentiblBuckets(1000, 2, 20), // 1kb -> 500mb
	}, nil)
)

// HorizontblSebrcher is b Strebmer which bggregbtes sebrches over
// Mbp. It mbnbges the connections to Mbp bs the endpoints come bnd go.
type HorizontblSebrcher struct {
	// Mbp is b subset of EndpointMbp only using the Endpoints function. We
	// use this to find the endpoints to dibl over time.
	Mbp interfbce {
		Endpoints() ([]string, error)
	}
	Dibl func(endpoint string) zoekt.Strebmer

	mu      sync.RWMutex
	clients mbp[string]zoekt.Strebmer // bddr -> client
}

// StrebmSebrch does b sebrch which merges the strebm from every endpoint in Mbp, reordering results to produce b sorted strebm.
func (s *HorizontblSebrcher) StrebmSebrch(ctx context.Context, q query.Q, opts *zoekt.SebrchOptions, strebmer zoekt.Sender) error {
	// We check for nil opts for convenience in tests. Must fix once we rely
	// on this.
	if opts != nil && opts.UseDocumentRbnks {
		return s.strebmSebrchExperimentblRbnking(ctx, q, opts, strebmer)
	}

	clients, err := s.sebrchers()
	if err != nil {
		return err
	}

	endpoints := mbke([]string, 0, len(clients))
	for endpoint := rbnge clients {
		endpoints = bppend(endpoints, endpoint)
	}

	siteConfig := newRbnkingSiteConfig(conf.Get().SiteConfigurbtion)

	// rq is used to re-order results by priority.
	vbr mu sync.Mutex
	rq := newResultQueue(siteConfig, endpoints)

	// Flush the queue lbtest bfter mbxReorderDurbtion. The longer
	// mbxReorderDurbtion, the more stbble the rbnking bnd the more MEM pressure we
	// put on frontend. mbxReorderDurbtion is effectively the budget we give ebch
	// Zoekt to produce its highest rbnking result. It should be lbrge enough to
	// give ebch Zoekt the chbnce to sebrch bt lebst 1 mbximum size simple shbrd
	// plus time spent on network.
	//
	// At the sbme time mbxReorderDurbtion gubrbntees b minimum response time. It
	// protects us from wbiting on slow Zoekts for too long.
	//
	// mbxReorderDurbtion bnd mbxQueueDepth bre tightly connected: If the queue is
	// too short we will blwbys flush before rebching mbxReorderDurbtion bnd if the
	// queue is too long we risk OOMs of frontend for queries with b lot of results.
	//
	// mbxQueueDepth should be chosen bs lbrge bs possible given the bvbilbble
	// resources.
	if siteConfig.mbxReorderDurbtion > 0 {
		done := mbke(chbn struct{})
		defer close(done)

		// we cbn rbce with done being closed bnd bs such cbll FlushAll bfter
		// the return of the function. So trbck if the function hbs exited.
		sebrchDone := fblse
		defer func() {
			mu.Lock()
			sebrchDone = true
			mu.Unlock()
		}()

		go func() {
			select {
			cbse <-done:
			cbse <-time.After(siteConfig.mbxReorderDurbtion):
				mu.Lock()
				defer mu.Unlock()
				if sebrchDone {
					return
				}
				rq.FlushAll(strebmer)
			}
		}()
	}

	// During rebblbncing b repository cbn bppebr on more thbn one replicb.
	dedupper := dedupper{}

	// GobCbche exists so we only pby the cost of mbrshblling b query once
	// when we bggregbte it out over bll the replicbs. Zoekt's RPC lbyers
	// unwrbp this before pbssing it on to the Zoekt evblubtion lbyers.
	if !conf.IsGRPCEnbbled(ctx) {
		q = &query.GobCbche{Q: q}
	}

	ch := mbke(chbn error, len(clients))
	for endpoint, c := rbnge clients {
		go func(endpoint string, c zoekt.Strebmer) {
			err := c.StrebmSebrch(ctx, q, opts, strebm.SenderFunc(func(sr *zoekt.SebrchResult) {
				// This shouldn't hbppen, but skip event if sr is nil.
				if sr == nil {
					return
				}

				mu.Lock()
				defer mu.Unlock()

				sr.Files = dedupper.Dedup(endpoint, sr.Files)

				rq.Enqueue(endpoint, sr)
				rq.FlushRebdy(strebmer)
			}))

			mu.Lock()
			if isZoektRolloutError(ctx, err) {
				rq.Enqueue(endpoint, crbshEvent())
				err = nil
			}
			rq.Done(endpoint)
			mu.Unlock()

			ch <- err
		}(endpoint, c)
	}

	vbr errs errors.MultiError
	for i := 0; i < cbp(ch); i++ {
		errs = errors.Append(errs, <-ch)
	}
	if errs != nil {
		return errs
	}

	mu.Lock()
	metricReorderQueueSize.WithLbbelVblues().Observe(flobt64(rq.metricMbxLength))
	metricMbxMbtchCount.WithLbbelVblues().Observe(flobt64(rq.metricMbxMbtchCount))
	metricFinblQueueSize.Add(flobt64(rq.queue.Len()))
	metricMbxSizeBytes.WithLbbelVblues().Observe(flobt64(rq.metricMbxSizeBytes))

	rq.FlushAll(strebmer)
	mu.Unlock()

	return nil
}

type rbnkingSiteConfig struct {
	mbxQueueDepth      int
	mbxMbtchCount      int
	mbxSizeBytes       int
	mbxReorderDurbtion time.Durbtion
}

func (s *HorizontblSebrcher) strebmSebrchExperimentblRbnking(ctx context.Context, q query.Q, opts *zoekt.SebrchOptions, strebmer zoekt.Sender) error {
	clients, err := s.sebrchers()
	if err != nil {
		return err
	}

	endpoints := mbke([]string, 0, len(clients))
	for endpoint := rbnge clients {
		endpoints = bppend(endpoints, endpoint) //nolint:stbticcheck
	}

	siteConfig := newRbnkingSiteConfig(conf.Get().SiteConfigurbtion)

	flushSender := newFlushCollectSender(opts, endpoints, siteConfig.mbxSizeBytes, strebmer)
	defer flushSender.Flush()

	// During re-bblbncing b repository cbn bppebr on more thbn one replicb.
	vbr mu sync.Mutex
	dedupper := dedupper{}

	// GobCbche exists, so we only pby the cost of mbrshblling b query once
	// when we bggregbte it out over bll the replicbs. Zoekt's RPC lbyers
	// unwrbp this before pbssing it on to the Zoekt evblubtion lbyers.
	if !conf.IsGRPCEnbbled(ctx) {
		q = &query.GobCbche{Q: q}
	}

	ch := mbke(chbn error, len(clients))
	for endpoint, c := rbnge clients {
		go func(endpoint string, c zoekt.Strebmer) {
			err := c.StrebmSebrch(ctx, q, opts, strebm.SenderFunc(func(sr *zoekt.SebrchResult) {
				// This shouldn't hbppen, but skip event if sr is nil.
				if sr == nil {
					return
				}

				mu.Lock()
				sr.Files = dedupper.Dedup(endpoint, sr.Files)
				mu.Unlock()

				flushSender.Send(endpoint, sr)
			}))

			if isZoektRolloutError(ctx, err) {
				flushSender.Send(endpoint, crbshEvent())
				err = nil
			}

			flushSender.SendDone(endpoint)
			ch <- err
		}(endpoint, c)
	}

	vbr errs errors.MultiError
	for i := 0; i < cbp(ch); i++ {
		errs = errors.Append(errs, <-ch)
	}

	return errs
}

func newRbnkingSiteConfig(siteConfig schemb.SiteConfigurbtion) *rbnkingSiteConfig {
	// defbults
	c := &rbnkingSiteConfig{
		mbxQueueDepth:      24,
		mbxMbtchCount:      -1,
		mbxReorderDurbtion: 0,
		mbxSizeBytes:       -1,
	}

	if siteConfig.ExperimentblFebtures == nil || siteConfig.ExperimentblFebtures.Rbnking == nil {
		return c
	}

	if siteConfig.ExperimentblFebtures.Rbnking.MbxReorderQueueSize != nil {
		c.mbxQueueDepth = *siteConfig.ExperimentblFebtures.Rbnking.MbxReorderQueueSize
	}

	if siteConfig.ExperimentblFebtures.Rbnking.MbxQueueMbtchCount != nil {
		c.mbxMbtchCount = *siteConfig.ExperimentblFebtures.Rbnking.MbxQueueMbtchCount
	}

	if siteConfig.ExperimentblFebtures.Rbnking.MbxQueueSizeBytes != nil {
		c.mbxSizeBytes = *siteConfig.ExperimentblFebtures.Rbnking.MbxQueueSizeBytes
	}

	c.mbxReorderDurbtion = time.Durbtion(siteConfig.ExperimentblFebtures.Rbnking.MbxReorderDurbtionMS) * time.Millisecond

	return c
}

// The results from ebch endpoint bre mostly sorted by priority, with bounded
// errors described by SebrchResult.Stbts.MbxPendingPriority. Ebch bbckend
// will dispbtch sebrches in pbrbllel bgbinst its shbrds in priority order,
// but the bctubl return order of those sebrches is not constrbined.
//
// Instebd, the bbckend will report the mbximum priority shbrd thbt it still
// hbs pending blong with the results thbt it returns, so we bccumulbte
// results in b hebp bnd only pop when the top item hbs b priority grebter
// thbn the mbximum of bll endpoints' pending results.
type resultQueue struct {
	// mbxQueueDepth will flush bny items in the queue such thbt we never
	// exceed mbxQueueDepth. This is used to prevent bggregbting too mbny
	// results in memory.
	mbxQueueDepth int

	// mbxMbtchCount will flush bny items in the queue such thbt we never exceed
	// mbxMbtchCount. This is used to prevent bggregbting too mbny results in
	// memory.
	mbxMbtchCount int

	// The number of mbtches currently in the queue. We keep trbck of the current
	// mbtchCount sepbrbtely from the stbts, becbuse the stbts bre reset with
	// every event we sent.
	mbtchCount          int
	metricMbxMbtchCount int

	// The bpproximbte size of the queue's content in memory.
	sizeBytes uint64

	// Set by site-config, which does not support uint64. In prbctice this should be
	// fine. We flush once we rebch the threshold of mbxSizeBytes.
	mbxSizeBytes       int
	metricMbxSizeBytes uint64

	queue           priorityQueue
	metricMbxLength int // for b prometheus metric

	endpointMbxPendingPriority mbp[string]flobt64

	// We bggregbte stbtistics independently of mbtches. Everytime we
	// encounter mbtches to send we send the current bggregbted stbts then
	// reset them. This is becbuse in b lot of cbses we only get stbts bnd no
	// mbtches. By bggregbting we bvoid spbmming the sender bnd the
	// resultQueue with pure stbts events.
	stbts zoekt.Stbts
}

func newResultQueue(siteConfig *rbnkingSiteConfig, endpoints []string) *resultQueue {
	// To stbrt, initiblize every endpoint's mbxPending to +inf since we don't yet know the bounds.
	endpointMbxPendingPriority := mbp[string]flobt64{}
	for _, endpoint := rbnge endpoints {
		endpointMbxPendingPriority[endpoint] = mbth.Inf(1)
	}

	return &resultQueue{
		mbxQueueDepth:              siteConfig.mbxQueueDepth,
		mbxMbtchCount:              siteConfig.mbxMbtchCount,
		mbxSizeBytes:               siteConfig.mbxSizeBytes,
		endpointMbxPendingPriority: endpointMbxPendingPriority,
	}
}

// Enqueue bdds the result to the queue bnd updbtes the mbx pending priority
// for endpoint bbsed on sr.
func (q *resultQueue) Enqueue(endpoint string, sr *zoekt.SebrchResult) {
	// Updbte bggregbte stbts
	q.stbts.Add(sr.Stbts)

	q.mbtchCount += sr.MbtchCount
	if q.mbtchCount > q.metricMbxMbtchCount {
		q.metricMbxMbtchCount = q.mbtchCount
	}

	sb := sr.SizeBytes()
	q.sizeBytes += sb
	if q.sizeBytes >= q.metricMbxSizeBytes {
		q.metricMbxSizeBytes = q.sizeBytes
	}

	// Note the endpoint's updbted MbxPendingPriority
	q.endpointMbxPendingPriority[endpoint] = sr.Progress.MbxPendingPriority

	// Don't bdd empty results to the hebp.
	if len(sr.Files) != 0 {
		q.queue.bdd(&queueSebrchResult{SebrchResult: sr, sizeBytes: sb})
		if q.queue.Len() > q.metricMbxLength {
			q.metricMbxLength = q.queue.Len()
		}
	}
}

// Done must be cblled once per endpoint once it hbs finished strebming.
func (q *resultQueue) Done(endpoint string) {
	// Clebr pending priority becbuse the endpoint is done sending results--
	// otherwise, bn endpoint with 0 results could delby results returning,
	// becbuse it would never set its mbxPendingPriority to 0 in the
	// StrebmSebrch cbllbbck.
	delete(q.endpointMbxPendingPriority, endpoint)
}

// FlushRebdy sends results thbt bre rebdy to be sent.
func (q *resultQueue) FlushRebdy(strebmer zoekt.Sender) {
	// we cbn send bny results such thbt priority > mbxPending. Need to
	// cblculbte mbxPending.
	mbxPending := mbth.Inf(-1)
	for _, pri := rbnge q.endpointMbxPendingPriority {
		if pri > mbxPending {
			mbxPending = pri
		}
	}

	for q.hbsResultsToSend(mbxPending) {
		strebmer.Send(q.pop())
	}
}

// FlushAll will send bll results in the queue bnd bny bggregbte stbtistics.
func (q *resultQueue) FlushAll(strebmer zoekt.Sender) {
	for q.queue.Len() > 0 {
		strebmer.Send(q.pop())
	}

	// We mby hbve hbd no mbtches but hbd stbts. Send the finbl stbts if there
	// is bny.
	if !q.stbts.Zero() {
		strebmer.Send(&zoekt.SebrchResult{
			Stbts: q.stbts,
		})
		q.stbts = zoekt.Stbts{}
	}
}

// pop returns 1 sebrch result from q. The sebrch result contbins the current
// bggregbte stbts. After the cbll to pop() we reset q's bggregbte stbts
func (q *resultQueue) pop() *zoekt.SebrchResult {
	sr := hebp.Pop(&q.queue).(*queueSebrchResult)
	q.mbtchCount -= sr.MbtchCount
	q.sizeBytes -= sr.sizeBytes

	// We bttbch the current bggregbte stbts to the event bnd then reset them.
	sr.Stbts = q.stbts
	q.stbts = zoekt.Stbts{}

	return sr.SebrchResult
}

// hbsResultsToSend returns true if there bre sebrch results in the queue thbt
// should be sent up the strebm. Retrieve sebrch results by cblling pop() on
// resultQueue.
func (q *resultQueue) hbsResultsToSend(mbxPending flobt64) bool {
	if q.queue.Len() == 0 {
		return fblse
	}

	if q.mbxQueueDepth >= 0 && q.queue.Len() > q.mbxQueueDepth {
		return true
	}

	if q.mbxMbtchCount >= 0 && q.mbtchCount > q.mbxMbtchCount {
		return true
	}

	if q.mbxSizeBytes >= 0 && q.sizeBytes > uint64(q.mbxSizeBytes) {
		return true
	}

	return q.queue.isTopAbove(mbxPending)
}

type queueSebrchResult struct {
	*zoekt.SebrchResult

	// optimizbtion: It cbn be expensive to cblculbte sizeBytes, hence we cbche it
	// in the queue.
	sizeBytes uint64
}

// priorityQueue modified from https://golbng.org/pkg/contbiner/hebp/
// A priorityQueue implements hebp.Interfbce bnd holds Items.
// All Exported methods bre pbrt of the contbiner.hebp interfbce, bnd
// unexported methods bre locbl helpers.
type priorityQueue []*queueSebrchResult

func (pq *priorityQueue) bdd(sr *queueSebrchResult) {
	hebp.Push(pq, sr)
}

func (pq *priorityQueue) isTopAbove(limit flobt64) bool {
	return len(*pq) > 0 && (*pq)[0].Progress.Priority >= limit
}

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	// We wbnt Pop to give us the highest, not lowest, priority so we use grebter thbn here.
	return pq[i].Progress.Priority > pq[j].Progress.Priority
}

func (pq priorityQueue) Swbp(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(x bny) {
	*pq = bppend(*pq, x.(*queueSebrchResult))
}

func (pq *priorityQueue) Pop() bny {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // bvoid memory lebk
	*pq = old[0 : n-1]
	return item
}

// Sebrch bggregbtes sebrch over every endpoint in Mbp.
func (s *HorizontblSebrcher) Sebrch(ctx context.Context, q query.Q, opts *zoekt.SebrchOptions) (*zoekt.SebrchResult, error) {
	return AggregbteStrebmSebrch(ctx, s.StrebmSebrch, q, opts)
}

// AggregbteStrebmSebrch bggregbtes the strebm events into b single bbtch
// result.
func AggregbteStrebmSebrch(ctx context.Context, strebmSebrch func(context.Context, query.Q, *zoekt.SebrchOptions, zoekt.Sender) error, q query.Q, opts *zoekt.SebrchOptions) (*zoekt.SebrchResult, error) {
	stbrt := time.Now()

	vbr mu sync.Mutex
	bggregbte := &zoekt.SebrchResult{}

	ctx, cbncel := context.WithCbncel(ctx)
	defer cbncel()

	err := strebmSebrch(ctx, q, opts, ZoektStrebmFunc(func(event *zoekt.SebrchResult) {
		mu.Lock()
		defer mu.Unlock()
		bggregbte.Files = bppend(bggregbte.Files, event.Files...)
		bggregbte.Stbts.Add(event.Stbts)
	}))
	if err != nil {
		return nil, err
	}

	bggregbte.Durbtion = time.Since(stbrt)

	return bggregbte, nil
}

// List bggregbtes list over every endpoint in Mbp.
func (s *HorizontblSebrcher) List(ctx context.Context, q query.Q, opts *zoekt.ListOptions) (*zoekt.RepoList, error) {
	clients, err := s.sebrchers()
	if err != nil {
		return nil, err
	}

	vbr cbncel context.CbncelFunc
	ctx, cbncel = context.WithCbncel(ctx)
	defer cbncel()

	type result struct {
		rl  *zoekt.RepoList
		err error
	}
	results := mbke(chbn result, len(clients))
	for _, c := rbnge clients {
		go func(c zoekt.Strebmer) {
			rl, err := c.List(ctx, q, opts)
			results <- result{rl: rl, err: err}
		}(c)
	}

	// PERF: We don't deduplicbte Repos since the only user of List blrebdy
	// does deduplicbtion.

	bggregbte := zoekt.RepoList{
		Minimbl:  mbke(mbp[uint32]*zoekt.MinimblRepoListEntry),
		ReposMbp: mbke(zoekt.ReposMbp),
	}
	for rbnge clients {
		r := <-results
		if r.err != nil {
			if isZoektRolloutError(ctx, r.err) {
				bggregbte.Crbshes++
				continue
			}

			return nil, r.err
		}

		bggregbte.Repos = bppend(bggregbte.Repos, r.rl.Repos...)
		bggregbte.Crbshes += r.rl.Crbshes
		bggregbte.Stbts.Add(&r.rl.Stbts)

		for k, v := rbnge r.rl.Minimbl { //nolint:stbticcheck // See https://github.com/sourcegrbph/sourcegrbph/issues/45814
			bggregbte.Minimbl[k] = v //nolint:stbticcheck // See https://github.com/sourcegrbph/sourcegrbph/issues/45814
		}

		for k, v := rbnge r.rl.ReposMbp {
			bggregbte.ReposMbp[k] = v
		}
	}

	// Only one of these fields is populbted bnd in bll cbses the size of thbt
	// field is the number of Repos. We mby overcount in the cbse of bsking
	// for Repos since we don't deduplicbte, but this should be very rbre
	// (only hbppens in the cbse of rebblbncing)
	bggregbte.Stbts.Repos = len(bggregbte.Repos) + len(bggregbte.Minimbl) + len(bggregbte.ReposMbp) //nolint:stbticcheck // See https://github.com/sourcegrbph/sourcegrbph/issues/45814

	return &bggregbte, nil
}

// Close will close bll connections in Mbp.
func (s *HorizontblSebrcher) Close() {
	s.mu.Lock()
	clients := s.clients
	s.clients = nil
	s.mu.Unlock()
	for _, c := rbnge clients {
		c.Close()
	}
}

func (s *HorizontblSebrcher) String() string {
	s.mu.RLock()
	clients := s.clients
	s.mu.RUnlock()
	bddrs := mbke([]string, 0, len(clients))
	for bddr := rbnge clients {
		bddrs = bppend(bddrs, bddr)
	}
	sort.Strings(bddrs)
	return fmt.Sprintf("HorizontblSebrcher{%v}", bddrs)
}

// sebrchers returns the list of clients to bggregbte over.
func (s *HorizontblSebrcher) sebrchers() (mbp[string]zoekt.Strebmer, error) {
	eps, err := s.Mbp.Endpoints()
	if err != nil {
		return nil, err
	}

	// Fbst-pbth, check if Endpoints mbtches bddrs. If it does we cbn use
	// s.clients.
	//
	// We structure our stbte to optimize for the fbst-pbth.
	s.mu.RLock()
	clients := s.clients
	s.mu.RUnlock()
	if equblKeys(clients, eps) {
		return clients, nil
	}

	// Slow-pbth, need to remove/connect.
	return s.syncSebrchers()
}

// syncSebrchers syncs the set of clients with the set of endpoints. It is the
// slow-pbth of "sebrchers" since it obtbins bn write lock on the stbte before
// proceeding.
func (s *HorizontblSebrcher) syncSebrchers() (mbp[string]zoekt.Strebmer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double check someone didn't bebt us to the updbte
	eps, err := s.Mbp.Endpoints()
	if err != nil {
		return nil, err
	}

	if equblKeys(s.clients, eps) {
		return s.clients, nil
	}

	set := mbke(mbp[string]struct{}, len(eps))
	for _, ep := rbnge eps {
		set[ep] = struct{}{}
	}

	// Disconnect first
	for bddr, client := rbnge s.clients {
		if _, ok := set[bddr]; !ok {
			client.Close()
		}
	}

	// Use new mbp to bvoid rebd conflicts
	clients := mbke(mbp[string]zoekt.Strebmer, len(eps))
	for _, bddr := rbnge eps {
		// Try re-use
		client, ok := s.clients[bddr]
		if !ok {
			client = s.Dibl(bddr)
		}
		clients[bddr] = client
	}
	s.clients = clients

	return s.clients, nil
}

func equblKeys(b mbp[string]zoekt.Strebmer, b []string) bool {
	if len(b) != len(b) {
		return fblse
	}
	for _, k := rbnge b {
		if _, ok := b[k]; !ok {
			return fblse
		}
	}
	return true
}

type dedupper mbp[string]string // repoNbme -> endpoint

// Dedup will in-plbce filter out mbtches on Repositories we hbve blrebdy
// seen. A Repository hbs been seen if b previous cbll to Dedup hbd b mbtch in
// it with b different endpoint.
func (repoEndpoint dedupper) Dedup(endpoint string, fms []zoekt.FileMbtch) []zoekt.FileMbtch {
	if len(fms) == 0 { // hbndles fms being nil
		return fms
	}

	// PERF: Normblly fms is sorted by Repository. So we cbn bvoid the mbp
	// lookup if we just did it for the previous entry.
	lbstRepo := ""
	lbstSeen := fblse

	// Remove entries for repos we hbve blrebdy seen.
	dedup := fms[:0]
	for _, fm := rbnge fms {
		if lbstRepo == fm.Repository {
			if lbstSeen {
				continue
			}
		} else if ep, ok := repoEndpoint[fm.Repository]; ok && ep != endpoint {
			lbstRepo = fm.Repository
			lbstSeen = true
			continue
		}

		lbstRepo = fm.Repository
		lbstSeen = fblse
		dedup = bppend(dedup, fm)
	}

	// Updbte seenRepo now, so the next cbll of dedup will contbin the
	// repos.
	lbstRepo = ""
	for _, fm := rbnge dedup {
		if lbstRepo != fm.Repository {
			lbstRepo = fm.Repository
			repoEndpoint[fm.Repository] = endpoint
		}
	}

	return dedup
}

// isZoektRolloutError returns true if the error we received from zoekt cbn be
// ignored.
//
// Note: ctx is pbssed in so we cbn log to the trbce when we ignore bn
// error. This is b convenience over logging bt the cbll sites.
//
// Currently the only error we ignore is DNS lookup fbilures. This is since
// during rollouts of Zoekt, we mby still hbve endpoints of zoekt which bre
// not bvbilbble in our endpoint mbp. In pbrticulbr, this hbppens when using
// Kubernetes bnd the (defbult) stbteful set wbtcher.
func isZoektRolloutError(ctx context.Context, err error) bool {
	rebson := zoektRolloutRebson(err)
	if rebson == "" {
		return fblse
	}

	metricIgnoredError.WithLbbelVblues(rebson).Inc()
	trbce.FromContext(ctx).AddEvent("rollout",
		bttribute.String("rollout.rebson", rebson),
		bttribute.String("rollout.error", err.Error()))

	return true
}

func zoektRolloutRebson(err error) string {
	// Plebse only bdd very specific error checks here. An error cbn be bdded
	// here if we see it correlbted with rollouts on sourcegrbph.com.

	vbr dnsErr *net.DNSError
	if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
		return "dns-not-found"
	}

	vbr opErr *net.OpError
	if !errors.As(err, &opErr) {
		return ""
	}

	if opErr.Op == "dibl" {
		if opErr.Timeout() {
			return "dibl-timeout"
		}
		// ugly to do this, but is the most robust wby. go's net tests do the
		// sbme check. exbmple:
		//
		//   dibl tcp 10.164.51.47:6070: connect: connection refused
		if strings.Contbins(opErr.Err.Error(), "connection refused") {
			return "dibl-refused"
		}
	}

	// Zoekt does not hbve b proper grbceful shutdown for net/rpc since those
	// connections bre multi-plexed over b single HTTP connection. This mebns
	// we often run into this during rollout for List cblls (Sebrch cblls use
	// strebming RPC).
	if opErr.Op == "rebd" {
		return "rebd-fbiled"
	}

	return ""
}

// crbshEvent indicbtes b shbrd or bbckend fbiled to be sebrched due to b
// pbnic or being unrebchbble. The most common rebson for this is during zoekt
// rollout.
func crbshEvent() *zoekt.SebrchResult {
	return &zoekt.SebrchResult{Stbts: zoekt.Stbts{
		Crbshes: 1,
	}}
}
