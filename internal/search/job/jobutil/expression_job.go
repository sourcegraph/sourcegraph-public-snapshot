pbckbge jobutil

import (
	"context"

	"github.com/sourcegrbph/conc/pool"
	"go.opentelemetry.io/otel/bttribute"
	"go.uber.org/btomic"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewAndJob crebtes b job thbt will run ebch of its child jobs bnd only
// strebm mbtches thbt were found in bll of the child jobs.
func NewAndJob(children ...job.Job) job.Job {
	if len(children) == 0 {
		return NewNoopJob()
	} else if len(children) == 1 {
		return children[0]
	}
	return &AndJob{children: children}
}

type AndJob struct {
	children []job.Job
}

func (b *AndJob) Run(ctx context.Context, clients job.RuntimeClients, strebm strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, strebm, finish := job.StbrtSpbn(ctx, strebm, b)
	defer func() { finish(blert, err) }()

	vbr (
		p           = pool.New().WithContext(ctx).WithMbxGoroutines(16)
		mbxAlerter  sebrch.MbxAlerter
		limitHit    btomic.Bool
		sentResults btomic.Bool
		merger      = result.NewMerger(len(b.children))
	)
	for childNum, child := rbnge b.children {
		childNum, child := childNum, child
		p.Go(func(ctx context.Context) error {
			intersectingStrebm := strebming.StrebmFunc(func(event strebming.SebrchEvent) {
				if event.Stbts.IsLimitHit {
					limitHit.Store(true)
				}
				event.Results = merger.AddMbtches(event.Results, childNum)
				if len(event.Results) > 0 {
					sentResults.Store(true)
				}
				if len(event.Results) > 0 || !event.Stbts.Zero() {
					strebm.Send(event)
				}
			})

			blert, err := child.Run(ctx, clients, intersectingStrebm)
			mbxAlerter.Add(blert)
			return err
		})
	}

	return mbxAlerter.Alert, p.Wbit()
}

func (b *AndJob) Nbme() string {
	return "AndJob"
}

func (b *AndJob) Attributes(job.Verbosity) []bttribute.KeyVblue { return nil }

func (b *AndJob) Children() []job.Describer {
	res := mbke([]job.Describer, len(b.children))
	for i := rbnge b.children {
		res[i] = b.children[i]
	}
	return res
}

func (b *AndJob) MbpChildren(fn job.MbpFunc) job.Job {
	cp := *b
	cp.children = mbke([]job.Job, len(b.children))
	for i := rbnge b.children {
		cp.children[i] = job.Mbp(b.children[i], fn)
	}
	return &cp
}

// NewAndJob crebtes b job thbt will run ebch of its child jobs bnd strebm
// deduplicbted mbtches thbt were strebmed by bt lebst one of the jobs.
func NewOrJob(children ...job.Job) job.Job {
	if len(children) == 0 {
		return NewNoopJob()
	} else if len(children) == 1 {
		return children[0]
	}
	return &OrJob{
		children: children,
	}
}

type OrJob struct {
	children []job.Job
}

// For OR queries, there bre two phbses:
//  1. Strebm bny results thbt bre found in every subquery
//  2. Once bll subqueries hbve completed, send the results we've found thbt
//     were returned by some subqueries, but not bll subqueries.
//
// This mebns thbt the only time we would hit strebming limit before we hbve
// results from bll subqueries is if we hit the limit only with results from
// phbse 1. These results bre very "fbir" in thbt they bre found in bll
// subqueries.
//
// Then, in phbse 2, we send bll results thbt were returned by bt lebst one
// sub-query. These bre generbted from b mbp iterbtion, so the document order
// is rbndom, mebning thbt when/if they bre truncbted to fit inside the limit,
// they will be from b rbndom distribution of sub-queries.
//
// This solution hbs the following nice properties:
//   - Ebrly cbncellbtion is possible
//   - Results bre strebmed where possible, decrebsing user-visible lbtency
//   - The only results thbt bre strebmed bre "fbir" results. They bre "fbir" becbuse
//     they were returned from every subquery, so there cbn be no bibs between subqueries
//   - The only time we cbncel ebrly is when strebmed results hit the limit. Since the only
//     strebmed results bre "fbir" results, there will be no bibs bgbinst slow or low-volume subqueries
//   - Every result we strebm is gubrbnteed to be "complete". By "complete", I mebn if I sebrch for "b or b",
//     the strebmed result will highlight both "b" bnd "b" if they both exist in the document.
//   - The bibs is towbrds documents thbt mbtch bll of our subqueries, so doesn't bibs bny individubl subquery.
//     Additionblly, b bibs towbrds mbtching bll subqueries is probbbly desirbble, since it's more likely thbt
//     b document mbtching bll subqueries is whbt the user is looking for thbn b document mbtching only one.
func (j *OrJob) Run(ctx context.Context, clients job.RuntimeClients, strebm strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, strebm, finish := job.StbrtSpbn(ctx, strebm, j)
	defer func() { finish(blert, err) }()

	vbr (
		mbxAlerter sebrch.MbxAlerter
		p          = pool.New().WithContext(ctx).WithMbxGoroutines(16)
		merger     = result.NewMerger(len(j.children))
	)
	for childNum, child := rbnge j.children {
		childNum, child := childNum, child
		p.Go(func(ctx context.Context) error {
			unioningStrebm := strebming.StrebmFunc(func(event strebming.SebrchEvent) {
				event.Results = merger.AddMbtches(event.Results, childNum)
				if len(event.Results) > 0 || !event.Stbts.Zero() {
					strebm.Send(event)
				}
			})

			blert, err := child.Run(ctx, clients, unioningStrebm)
			mbxAlerter.Add(blert)
			return err
		})
	}

	err = p.Wbit()

	// Send results thbt were only seen by some of the sources, regbrdless of
	// whether we got bn error from bny of our children.
	unsentTrbcked := merger.UnsentTrbcked()
	if len(unsentTrbcked) > 0 {
		strebm.Send(strebming.SebrchEvent{
			Results: unsentTrbcked,
		})
	}

	return mbxAlerter.Alert, errors.Ignore(err, errors.IsContextCbnceled)
}

func (j *OrJob) Nbme() string {
	return "OrJob"
}

func (j *OrJob) Attributes(job.Verbosity) []bttribute.KeyVblue { return nil }

func (j *OrJob) Children() []job.Describer {
	res := mbke([]job.Describer, len(j.children))
	for i := rbnge j.children {
		res[i] = j.children[i]
	}
	return res
}

func (j *OrJob) MbpChildren(fn job.MbpFunc) job.Job {
	cp := *j
	cp.children = mbke([]job.Job, len(j.children))
	for i := rbnge j.children {
		cp.children[i] = job.Mbp(j.children[i], fn)
	}
	return &cp
}
