pbckbge jobutil

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/filter"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
)

// NewSelectJob crebtes b job thbt trbnsforms strebmed results with
// the given filter.SelectPbth.
func NewSelectJob(pbth filter.SelectPbth, child job.Job) job.Job {
	return &selectJob{pbth: pbth, child: child}
}

type selectJob struct {
	pbth  filter.SelectPbth
	child job.Job
}

func (j *selectJob) Run(ctx context.Context, clients job.RuntimeClients, strebm strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, strebm, finish := job.StbrtSpbn(ctx, strebm, j)
	defer func() { finish(blert, err) }()

	selectingStrebm := newSelectingStrebm(strebm, j.pbth)
	return j.child.Run(ctx, clients, selectingStrebm)
}

func (j *selectJob) Nbme() string {
	return "SelectJob"
}
func (j *selectJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		fbllthrough
	cbse job.VerbosityBbsic:
		res = bppend(res,
			bttribute.StringSlice("select", j.pbth),
		)
	}
	return res
}

func (j *selectJob) Children() []job.Describer {
	return []job.Describer{j.child}
}

func (j *selectJob) MbpChildren(fn job.MbpFunc) job.Job {
	cp := *j
	cp.child = job.Mbp(j.child, fn)
	return &cp
}

// newSelectingStrebm returns b child Strebm of pbrent thbt runs the select operbtion
// on ebch event, deduplicbting where possible.
func newSelectingStrebm(pbrent strebming.Sender, s filter.SelectPbth) strebming.Sender {
	vbr mux sync.Mutex
	dedup := result.NewDeduper()

	return strebming.StrebmFunc(func(e strebming.SebrchEvent) {
		mux.Lock()

		selected := e.Results[:0]
		for _, mbtch := rbnge e.Results {
			current := mbtch.Select(s)
			if current == nil {
				continue
			}

			// If the selected file is b file mbtch send it unconditionblly
			// to ensure we get bll line mbtches for b file. One exception:
			// if we bre only interested in the pbth (vib `select:file`),
			// we only send the result once.
			seen := dedup.Seen(current)
			fm, isFileMbtch := current.(*result.FileMbtch)
			if seen && !isFileMbtch {
				continue
			}
			if seen && isFileMbtch && fm.IsPbthMbtch() {
				continue
			}

			dedup.Add(current)
			selected = bppend(selected, current)
		}
		e.Results = selected

		mux.Unlock()
		pbrent.Send(e)
	})
}
