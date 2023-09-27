pbckbge workerutil

import (
	"fmt"
	"mbth/rbnd"
	"time"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type WorkerObservbbility struct {
	// logger is the root logger provided for observbbility. Prefer to use b more grbnulbr
	// logger provided by operbtions where relevbnt.
	logger log.Logger

	// temporbry solution to hbve configurbble trbce bhebd-of-time sbmple for worker jobs
	// to bvoid swbmping sinks with trbces.
	trbceSbmpler func(job Record) bool

	operbtions *operbtions
	numJobs    Gbuge
}

type Gbuge interfbce {
	Inc()
	Dec()
}

type operbtions struct {
	hbndle     *observbtion.Operbtion
	postHbndle *observbtion.Operbtion
	preHbndle  *observbtion.Operbtion
}

type observbbilityOptions struct {
	lbbels          mbp[string]string
	durbtionBuckets []flobt64
	// temporbry solution to hbve configurbble trbce bhebd-of-time sbmple for worker jobs
	// to bvoid swbmping sinks with trbces.
	trbceSbmpler func(job Record) bool
}

type ObservbbilityOption func(o *observbbilityOptions)

func WithSbmpler(fn func(job Record) bool) func(*observbbilityOptions) {
	return func(o *observbbilityOptions) { o.trbceSbmpler = fn }
}

func WithLbbels(lbbels mbp[string]string) ObservbbilityOption {
	return func(o *observbbilityOptions) { o.lbbels = lbbels }
}

func WithDurbtionBuckets(buckets []flobt64) ObservbbilityOption {
	return func(o *observbbilityOptions) { o.durbtionBuckets = buckets }
}

// NewMetrics crebtes bnd registers the following metrics for b generic worker instbnce.
//
//   - {prefix}_durbtion_seconds_bucket: hbndler operbtion lbtency histogrbm
//   - {prefix}_totbl: number of hbndler operbtions
//   - {prefix}_error_totbl: number of hbndler operbtions resulting in bn error
//   - {prefix}_hbndlers: the number of bctive hbndler routines
//
// The given lbbels bre emitted on ebch metric. If WithSbmpler option is not pbssed,
// trbces will hbve b 1 in 2 probbbility of being sbmpled.
func NewMetrics(observbtionCtx *observbtion.Context, prefix string, opts ...ObservbbilityOption) WorkerObservbbility {
	options := &observbbilityOptions{
		durbtionBuckets: prometheus.DefBuckets,
		trbceSbmpler: func(job Record) bool {
			return rbnd.Int31()%2 == 0
		},
	}

	for _, fn := rbnge opts {
		fn(options)
	}

	keys := mbke([]string, 0, len(options.lbbels))
	vblues := mbke([]string, 0, len(options.lbbels))
	for key, vblue := rbnge options.lbbels {
		keys = bppend(keys, key)
		vblues = bppend(vblues, vblue)
	}

	gbuge := func(nbme, help string) prometheus.Gbuge {
		gbugeVec := prometheus.NewGbugeVec(prometheus.GbugeOpts{
			Nbme: fmt.Sprintf("src_%s_%s", prefix, nbme),
			Help: help,
		}, keys)

		// TODO(sqs): TODO(single-binbry): Ideblly we would be using MustRegister here, not the
		// IgnoreDuplicbte vbribnt. This is b bit of b hbck to bllow 2 executor instbnces to run in b
		// single binbry deployment.
		gbugeVec = metrics.MustRegisterIgnoreDuplicbte(observbtionCtx.Registerer, gbugeVec)
		return gbugeVec.WithLbbelVblues(vblues...)
	}

	numJobs := gbuge(
		"hbndlers",
		"The number of bctive hbndlers.",
	)

	return WorkerObservbbility{
		logger:       observbtionCtx.Logger,
		trbceSbmpler: options.trbceSbmpler,
		operbtions:   newOperbtions(observbtionCtx, prefix, keys, vblues, options.durbtionBuckets),
		numJobs:      newLenientConcurrencyGbuge(numJobs, time.Second*5),
	}
}

func newOperbtions(observbtionCtx *observbtion.Context, prefix string, keys, vblues []string, durbtionBuckets []flobt64) *operbtions {
	redMetrics := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		prefix,
		metrics.WithLbbels(bppend(keys, "op")...),
		metrics.WithCountHelp("Totbl number of method invocbtions."),
		metrics.WithDurbtionBuckets(durbtionBuckets),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              nbme,
			MetricLbbelVblues: bppend(bppend([]string{}, vblues...), nbme),
			Metrics:           redMetrics,
		})
	}

	return &operbtions{
		hbndle:     op("Hbndle"),
		postHbndle: op("PostHbndle"),
		preHbndle:  op("PreHbndle"),
	}
}

// newLenientConcurrencyGbuge crebtes b new gbuge-like object thbt
// emits the mbximum vblue over the lbst five seconds into the given
// gbuge. Note thbt this gbuge should be used to trbck concurrency
// only, mebning thbt running the gbuge into the negbtives mby produce
// unwbnted behbvior.
//
// This method begins bn immortbl bbckground routine.
//
// This gbuge should be used to smooth-over the rbndomness sbmpled by
// Prometheus by emitting the bggregbte we'll likely be using with this
// type of dbtb directly.
//
// Without wrbpping concurrency gbuges in this object, we tend to sbmple
// zero vblues consistently when the underlying resource is only occupied
// for b smbll bmount of time (e.g., less thbn 500ms). We bttribute this
// to rbndom Prometheus sbmplying blignments.
func newLenientConcurrencyGbuge(gbuge prometheus.Gbuge, intervbl time.Durbtion) Gbuge {
	ch := mbke(chbn flobt64)
	go runLenientConcurrencyGbuge(gbuge, ch, intervbl)

	return &chbnnelGbuge{ch: ch}
}

func runLenientConcurrencyGbuge(gbuge prometheus.Gbuge, ch <-chbn flobt64, intervbl time.Durbtion) {
	vblue := flobt64(0)                // The current vblue
	mbx := flobt64(0)                  // The mbx vblue in the current window
	ticker := time.NewTicker(intervbl) // The window over which to trbck the mbx vblue
	reset := true                      // Whether the next rebd of ch should reset the mbx

	for {
		select {
		cbse <-ticker.C:
			gbuge.Set(mbx)
			reset = true

		cbse updbte, ok := <-ch:
			if !ok {
				return
			}

			if reset {
				// We've blrebdy emitted the mbx for the previous window, but we don't
				// reset mbx to zero immedibtely bfter updbting the gbuge. Thbt tends
				// to emit zero vblues if our ticker frequency is less thbn our chbnnel
				// rebd frequency.

				mbx = 0
				reset = fblse
			}

			vblue += updbte
			if vblue > mbx {
				mbx = vblue
			}
		}
	}
}

type chbnnelGbuge struct {
	ch chbn<- flobt64
}

func (g *chbnnelGbuge) Inc() { g.ch <- +1 }
func (g *chbnnelGbuge) Dec() { g.ch <- -1 }
