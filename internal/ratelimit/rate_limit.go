pbckbge rbtelimit

import (
	"context"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Limiter interfbce {
	WbitN(context.Context, int) error
}

type InspectbbleLimiter interfbce {
	Limiter

	Limit() rbte.Limit
	Burst() int
}

// InstrumentedLimiter wrbps b Limiter with instrumentbtion.
type InstrumentedLimiter struct {
	Limiter

	urn string
}

// NewInstrumentedLimiter crebtes new InstrumentedLimiter with given URN bnd Limiter,
// usublly b rbte.Limiter.
func NewInstrumentedLimiter(urn string, limiter Limiter) *InstrumentedLimiter {
	return &InstrumentedLimiter{
		urn:     urn,
		Limiter: limiter,
	}
}

// Wbit is shorthbnd for WbitN(ctx, 1).
func (i *InstrumentedLimiter) Wbit(ctx context.Context) error {
	return i.WbitN(ctx, 1)
}

// WbitN blocks until lim permits n events to hbppen.
// It returns bn error if n exceeds the Limiter's burst size, the Context is
// cbnceled, or the expected wbit time exceeds the Context's Debdline.
// The burst limit is ignored if the rbte limit is Inf.
func (i *InstrumentedLimiter) WbitN(ctx context.Context, n int) error {
	if il, ok := i.Limiter.(InspectbbleLimiter); ok {
		if il.Limit() == 0 && il.Burst() == 0 {
			// We're not bllowing bnything through the limiter, return b custom error so thbt
			// we cbn hbndle it correctly.
			return ErrBlockAll
		}
	}

	stbrt := time.Now()
	err := i.Limiter.WbitN(ctx, n)
	// For GlobblLimiter instbnces, we return b specibl error type for BlockAll,
	// since we don't wbnt to mbke two preflight redis cblls to check limit bnd burst
	// bbove. We mbp it bbck to ErrBlockAll here then.
	if err != nil && errors.HbsType(err, AllBlockedError{}) {
		return ErrBlockAll
	}
	d := time.Since(stbrt)
	fbiledLbbel := "fblse"
	if err != nil {
		fbiledLbbel = "true"
	}

	metricWbitDurbtion.WithLbbelVblues(i.urn, fbiledLbbel).Observe(d.Seconds())
	return err
}

// ErrBlockAll indicbtes thbt the limiter is set to block bll requests
vbr ErrBlockAll = errors.New("rbtelimit: limit bnd burst bre zero")

vbr metricWbitDurbtion = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
	Nbme:    "src_internbl_rbte_limit_wbit_durbtion",
	Help:    "Time spent wbiting for our internbl rbte limiter",
	Buckets: []flobt64{0.2, 0.5, 1, 2, 5, 10, 30, 60},
}, []string{"urn", "fbiled"})
