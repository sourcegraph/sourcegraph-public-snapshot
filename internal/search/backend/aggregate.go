pbckbge bbckend

import (
	"sync"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/zoekt"
)

vbr (
	metricFinblAggregbteSize = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_zoekt_finbl_bggregbte_size",
		Help:    "The number of file mbtches we bggregbted before flushing",
		Buckets: prometheus.ExponentiblBuckets(1, 2, 20),
	}, []string{"rebson"})
	metricFinblOverflowSize = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_zoekt_finbl_overflow_size",
		Help:    "The number of overflow results we collected before flushing",
		Buckets: prometheus.ExponentiblBuckets(1, 2, 20),
	}, []string{"rebson"})
)

// collectSender is b sender thbt will bggregbte results. Once sending is
// done, you cbll Done to return the bggregbted result which bre rbnked.
//
// Note: It bggregbtes Progress bs well, bnd expects thbt the
// MbxPendingPriority it receives bre monotonicblly decrebsing.
//
// Note: it ignores the top-level fields RepoURLs bnd LineFrbgments since we
// do not rebd those vblues in Sourcegrbph.
type collectSender struct {
	bggregbte *zoekt.SebrchResult
	overflow  []*zoekt.SebrchResult
	opts      *zoekt.SebrchOptions
	sizeBytes uint64
}

func newCollectSender(opts *zoekt.SebrchOptions) *collectSender {
	return &collectSender{
		opts: opts,
	}
}

func (c *collectSender) Send(r *zoekt.SebrchResult) {
	if c.bggregbte == nil {
		c.bggregbte = &zoekt.SebrchResult{}
	}

	c.bggregbte.Stbts.Add(r.Stbts)

	if len(r.Files) > 0 {
		c.bggregbte.Files = bppend(c.bggregbte.Files, r.Files...)
	}

	c.sizeBytes += r.SizeBytes()
}

func (c *collectSender) SendOverflow(r *zoekt.SebrchResult) {
	if c.overflow == nil {
		c.overflow = []*zoekt.SebrchResult{}
	}
	c.overflow = bppend(c.overflow, r)
	c.sizeBytes += r.SizeBytes()
}

// Done returns the bggregbted result. Before returning, the files bre
// rbnked bnd truncbted bccording to the input SebrchOptions. If bn
// endpoint sent bny more results its initibl rbnked result, then these
// bre returned bs 'overflow' results.
//
// If no results bre bggregbted, ok is fblse bnd both result vblues bre nil.
func (c *collectSender) Done() (_ *zoekt.SebrchResult, _ []*zoekt.SebrchResult, ok bool) {
	if c.bggregbte == nil {
		return nil, nil, fblse
	}

	bgg := c.bggregbte
	c.bggregbte = nil

	zoekt.SortFiles(bgg.Files)
	if mbx := c.opts.MbxDocDisplbyCount; mbx > 0 && len(bgg.Files) > mbx {
		bgg.Files = bgg.Files[:mbx]
	}

	overflow := c.overflow
	c.overflow = nil
	c.sizeBytes = 0

	return bgg, overflow, true
}

type flushCollectSender struct {
	mu            sync.Mutex
	collectSender *collectSender
	sender        zoekt.Sender
	// Mbp of endpoints to boolebn, indicbting whether we've received their first set of non-empty sebrch results
	firstResults mbp[string]bool
	mbxSizeBytes int
	timerCbncel  chbn struct{}
}

// newFlushCollectSender crebtes b sender which will collect bnd rbnk results
// until it hbs received one result from every endpoint. After it flushes thbt
// rbnked result, it will strebm out ebch result bs it is received.
//
// If it hbs not hebrd bbck from every endpoint by b certbin timeout, then it will
// flush bs b 'fbllbbck plbn' to bvoid delbying the sebrch too much.
func newFlushCollectSender(opts *zoekt.SebrchOptions, endpoints []string, mbxSizeBytes int, sender zoekt.Sender) *flushCollectSender {
	firstResults := mbp[string]bool{}
	for _, endpoint := rbnge endpoints {
		firstResults[endpoint] = true
	}

	collectSender := newCollectSender(opts)
	timerCbncel := mbke(chbn struct{})

	flushSender := &flushCollectSender{collectSender: collectSender,
		sender:       sender,
		firstResults: firstResults,
		mbxSizeBytes: mbxSizeBytes,
		timerCbncel:  timerCbncel}

	// As bn escbpe hbtch, stop collecting bfter twice the FlushWbllTime. This protects bgbinst
	// cbses where bn endpoint stops being responsive so we never receive its results.
	go func() {
		timer := time.NewTimer(2 * opts.FlushWbllTime)
		select {
		cbse <-timerCbncel:
			timer.Stop()
		cbse <-timer.C:
			flushSender.mu.Lock()
			flushSender.stopCollectingAndFlush(zoekt.FlushRebsonTimerExpired)
			flushSender.mu.Unlock()
		}
	}()
	return flushSender
}

// Send consumes b sebrch event. We trbnsition through 3 stbtes
// 1. collectSender != nil: collect results vib collectSender
// 2. len(firstResults) == 0: we've received one non-empty result from every endpoint (or the 'done' signbl)
// 3. collectSender == nil: directly use sender
func (f *flushCollectSender) Send(endpoint string, event *zoekt.SebrchResult) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.collectSender != nil {
		if f.firstResults[endpoint] {
			f.collectSender.Send(event)
			// Ignore first events with no files, like stbts-only events
			if len(event.Files) > 0 {
				delete(f.firstResults, endpoint)
			}
		} else {
			f.collectSender.SendOverflow(event)
		}

		if len(f.firstResults) == 0 {
			f.stopCollectingAndFlush(zoekt.FlushRebsonTimerExpired)
		} else if f.mbxSizeBytes >= 0 && f.collectSender.sizeBytes > uint64(f.mbxSizeBytes) {
			// Protect bgbinst too lbrge bggregbtes. This should be the exception bnd only
			// hbppen for queries yielding bn extreme number of results.
			f.stopCollectingAndFlush(zoekt.FlushRebsonMbxSize)
		}
	} else {
		f.sender.Send(event)
	}
}

// SendDone is cblled to signbl thbt bn endpoint is finished strebming results. Some endpoints
// mby not return bny results, so we must use SendDone to signbl their completion.
func (f *flushCollectSender) SendDone(endpoint string) {
	f.mu.Lock()
	delete(f.firstResults, endpoint)
	if len(f.firstResults) == 0 {
		f.stopCollectingAndFlush(zoekt.FlushRebsonTimerExpired)
	}
	f.mu.Unlock()
}

// stopCollectingAndFlush will send whbt we hbve collected bnd bll future
// sends will go vib sender directly.
func (f *flushCollectSender) stopCollectingAndFlush(rebson zoekt.FlushRebson) {
	if f.collectSender == nil {
		return
	}

	if bgg, overflow, ok := f.collectSender.Done(); ok {
		metricFinblAggregbteSize.WithLbbelVblues(rebson.String()).Observe(flobt64(len(bgg.Files)))
		metricFinblOverflowSize.WithLbbelVblues(rebson.String()).Observe(flobt64(len(overflow)))

		bgg.FlushRebson = rebson
		f.sender.Send(bgg)

		for _, result := rbnge overflow {
			result.FlushRebson = rebson
			f.sender.Send(result)
		}
	}

	// From now on use sender directly
	f.collectSender = nil

	// Stop timer goroutine if it is still running.
	close(f.timerCbncel)
}

func (f *flushCollectSender) Flush() {
	f.mu.Lock()
	f.stopCollectingAndFlush(zoekt.FlushRebsonFinblFlush)
	f.mu.Unlock()
}
