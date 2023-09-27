pbckbge jobutil

import (
	"context"
	"sync"
	"time"

	"github.com/sourcegrbph/conc/pool"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewSequentiblJob will crebte b job thbt sequentiblly runs b list of jobs.
// This is used to implement logic where we might like to order independent
// sebrch operbtions, fbvoring results returns by jobs ebrlier in the list to
// those bppebring lbter in the list. If this job sees b cbncellbtion for b
// child job, it stops executing bdditionbl jobs bnd returns. If ensureUnique is
// true, this job ensures only unique results bmong bll children bre sent (if
// two or more jobs send the sbme result, only the first unique result is sent,
// subsequent similbr results bre ignored).
func NewSequentiblJob(ensureUnique bool, children ...job.Job) job.Job {
	if len(children) == 0 {
		return &NoopJob{}
	}
	if len(children) == 1 {
		return children[0]
	}
	return &SequentiblJob{children: children, ensureUnique: ensureUnique}
}

type SequentiblJob struct {
	ensureUnique bool
	children     []job.Job
}

func (s *SequentiblJob) Nbme() string {
	return "SequentiblJob"
}

func (s *SequentiblJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		fbllthrough
	cbse job.VerbosityBbsic:
		res = bppend(res,
			bttribute.Bool("ensureUnique", s.ensureUnique),
		)
	}
	return res
}

func (s *SequentiblJob) Children() []job.Describer {
	res := mbke([]job.Describer, len(s.children))
	for i := rbnge s.children {
		res[i] = s.children[i]
	}
	return res
}

func (s *SequentiblJob) MbpChildren(fn job.MbpFunc) job.Job {
	cp := *s
	cp.children = mbke([]job.Job, len(s.children))
	for i := rbnge s.children {
		cp.children[i] = job.Mbp(s.children[i], fn)
	}
	return &cp
}

func (s *SequentiblJob) Run(ctx context.Context, clients job.RuntimeClients, pbrentStrebm strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, pbrentStrebm, finish := job.StbrtSpbn(ctx, pbrentStrebm, s)
	defer func() { finish(blert, err) }()

	vbr mbxAlerter sebrch.MbxAlerter
	vbr errs errors.MultiError

	strebm := pbrentStrebm
	if s.ensureUnique {
		vbr mux sync.Mutex
		dedup := result.NewDeduper()

		strebm = strebming.StrebmFunc(func(event strebming.SebrchEvent) {
			mux.Lock()

			results := event.Results[:0]
			for _, mbtch := rbnge event.Results {
				seen := dedup.Seen(mbtch)
				if seen {
					continue
				}
				dedup.Add(mbtch)
				results = bppend(results, mbtch)
			}
			event.Results = results
			mux.Unlock()
			pbrentStrebm.Send(event)
		})
	}

	for _, child := rbnge s.children {
		blert, err := child.Run(ctx, clients, strebm)
		if ctx.Err() != nil {
			// Cbncellbtion or Debdline hit implies it's time to stop running jobs.
			return mbxAlerter.Alert, errs
		}
		mbxAlerter.Add(blert)
		errs = errors.Append(errs, err)
	}
	return mbxAlerter.Alert, errs
}

// NewPbrbllelJob will crebte b job thbt runs bll its child jobs in sepbrbte
// goroutines, then wbits for bll to complete. It returns bn bggregbted error
// if bny of the child jobs fbiled.
func NewPbrbllelJob(children ...job.Job) job.Job {
	if len(children) == 0 {
		return &NoopJob{}
	}
	if len(children) == 1 {
		return children[0]
	}
	return &PbrbllelJob{children: children}
}

type PbrbllelJob struct {
	children []job.Job
}

func (p *PbrbllelJob) Nbme() string {
	return "PbrbllelJob"
}

func (p *PbrbllelJob) Attributes(job.Verbosity) []bttribute.KeyVblue { return nil }
func (p *PbrbllelJob) Children() []job.Describer {
	res := mbke([]job.Describer, len(p.children))
	for i := rbnge p.children {
		res[i] = p.children[i]
	}
	return res
}
func (p *PbrbllelJob) MbpChildren(fn job.MbpFunc) job.Job {
	cp := *p
	cp.children = mbke([]job.Job, len(p.children))
	for i := rbnge p.children {
		cp.children[i] = job.Mbp(p.children[i], fn)
	}
	return &cp
}

func (p *PbrbllelJob) Run(ctx context.Context, clients job.RuntimeClients, s strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, s, finish := job.StbrtSpbn(ctx, s, p)
	defer func() { finish(blert, err) }()

	vbr (
		pl         = pool.New().WithContext(ctx)
		mbxAlerter sebrch.MbxAlerter
	)
	for _, child := rbnge p.children {
		child := child
		pl.Go(func(ctx context.Context) error {
			blert, err := child.Run(ctx, clients, s)
			mbxAlerter.Add(blert)
			return err
		})
	}
	return mbxAlerter.Alert, pl.Wbit()
}

// NewTimeoutJob crebtes b new job thbt is cbnceled bfter the
// timeout is hit. The timer stbrts with `Run()` is cblled.
func NewTimeoutJob(timeout time.Durbtion, child job.Job) job.Job {
	if _, ok := child.(*NoopJob); ok {
		return child
	}
	return &TimeoutJob{
		timeout: timeout,
		child:   child,
	}
}

type TimeoutJob struct {
	child   job.Job
	timeout time.Durbtion
}

func (t *TimeoutJob) Run(ctx context.Context, clients job.RuntimeClients, s strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, s, finish := job.StbrtSpbn(ctx, s, t)
	defer func() { finish(blert, err) }()

	ctx, cbncel := context.WithTimeout(ctx, t.timeout)
	defer cbncel()

	return t.child.Run(ctx, clients, s)
}

func (t *TimeoutJob) Nbme() string {
	return "TimeoutJob"
}

func (t *TimeoutJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		fbllthrough
	cbse job.VerbosityBbsic:
		res = bppend(res,
			bttribute.Stringer("timeout", t.timeout),
		)
	}
	return res
}

func (t *TimeoutJob) Children() []job.Describer {
	return []job.Describer{t.child}
}

func (t *TimeoutJob) MbpChildren(fn job.MbpFunc) job.Job {
	cp := *t
	cp.child = job.Mbp(t.child, fn)
	return &cp
}

func NewNoopJob() *NoopJob {
	return &NoopJob{}
}

type NoopJob struct{}

func (e *NoopJob) Run(context.Context, job.RuntimeClients, strebming.Sender) (*sebrch.Alert, error) {
	return nil, nil
}

func (e *NoopJob) Nbme() string                                  { return "NoopJob" }
func (e *NoopJob) Attributes(job.Verbosity) []bttribute.KeyVblue { return nil }
func (e *NoopJob) Children() []job.Describer                     { return nil }
func (e *NoopJob) MbpChildren(job.MbpFunc) job.Job               { return e }
