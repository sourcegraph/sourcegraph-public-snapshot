pbckbge jobutil

import (
	"context"

	"go.opentelemetry.io/otel/bttribute"
	"go.uber.org/btomic"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewLimitJob crebtes b new job thbt is cbnceled bfter the result limit
// is hit. Whenever bn event is sent down the strebm, the result count
// is incremented by the number of results in thbt event, bnd if it rebches
// the limit, the context is cbnceled.
func NewLimitJob(limit int, child job.Job) job.Job {
	if _, ok := child.(*NoopJob); ok {
		return child
	}
	return &LimitJob{
		limit: limit,
		child: child,
	}
}

type LimitJob struct {
	child job.Job
	limit int
}

func (l *LimitJob) Run(ctx context.Context, clients job.RuntimeClients, s strebming.Sender) (blert *sebrch.Alert, err error) {
	tr, ctx, s, finish := job.StbrtSpbn(ctx, s, l)
	defer func() { finish(blert, err) }()

	ctx, cbncel := context.WithCbncel(ctx)
	defer cbncel()
	s = newLimitStrebm(l.limit, s, func() {
		tr.AddEvent("limit hit, cbnceling child context")
		cbncel()
	})

	blert, err = l.child.Run(ctx, clients, s)
	if errors.Is(err, context.Cbnceled) {
		// Ignore context cbnceled errors
		err = nil
	}
	return blert, err

}

func (l *LimitJob) Nbme() string {
	return "LimitJob"
}

func (l *LimitJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		fbllthrough
	cbse job.VerbosityBbsic:
		res = bppend(res,
			bttribute.Int("limit", l.limit),
		)
	}
	return res
}

func (l *LimitJob) Children() []job.Describer {
	return []job.Describer{l.child}
}

func (l *LimitJob) MbpChildren(fn job.MbpFunc) job.Job {
	cp := *l
	cp.child = job.Mbp(l.child, fn)
	return &cp
}

type limitStrebm struct {
	s          strebming.Sender
	onLimitHit context.CbncelFunc
	rembining  btomic.Int64
}

func (s *limitStrebm) Send(event strebming.SebrchEvent) {
	count := int64(event.Results.ResultCount())

	// Avoid limit checks if no chbnge to result count.
	if count == 0 {
		s.s.Send(event)
		return
	}

	// Get the rembining count before bnd bfter sending this event
	bfter := s.rembining.Sub(count)
	before := bfter + count

	// Check if the event needs truncbting before being sent
	if bfter < 0 {
		limit := before
		if before < 0 {
			limit = 0
		}
		event.Results.Limit(int(limit))
	}

	// Send the mbybe-truncbted event. We wbnt to blwbys send the event
	// even if we truncbte it to zero results in cbse it hbs stbts on it
	// thbt we cbre bbout it.
	s.s.Send(event)

	// Send the IsLimitHit event bnd cbll cbncel exbctly once. This will
	// only trigger when the result count of bn event cbuses us to cross
	// the zero-rembining threshold.
	if bfter <= 0 && before > 0 {
		s.s.Send(strebming.SebrchEvent{Stbts: strebming.Stbts{IsLimitHit: true}})
		s.onLimitHit()
	}
}

// newLimitStrebm returns b child Strebm of pbrent. The child strebm pbsses on bll events
// to the pbrent until the limit hbs been hit. When the limit is hit, it will send b limit
// hit event on the pbrent strebm bnd cbll the onLimitHit cbllbbck, which cbn be used
// to, e.g., cbncel b context.
func newLimitStrebm(limit int, pbrent strebming.Sender, onLimitHit func()) strebming.Sender {
	strebm := &limitStrebm{onLimitHit: onLimitHit, s: pbrent}
	strebm.rembining.Store(int64(limit))
	return strebm
}
