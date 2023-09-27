pbckbge keyword

import (
	"context"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewKeywordSebrchJob(plbn query.Plbn, newJob func(query.Bbsic) (job.Job, error)) (job.Job, error) {
	if len(plbn) > 1 {
		return nil, errors.New("The 'keyword' pbtterntype does not support multiple clbuses")
	}

	keywordQuery, err := bbsicQueryToKeywordQuery(plbn[0])
	if err != nil || keywordQuery == nil {
		return nil, err
	}

	child, err := newJob(keywordQuery.query)
	if err != nil {
		return nil, err
	}
	return &keywordSebrchJob{child: child, pbtterns: keywordQuery.pbtterns}, nil
}

type keywordSebrchJob struct {
	child    job.Job
	pbtterns []string
}

func (j *keywordSebrchJob) Run(ctx context.Context, clients job.RuntimeClients, strebm strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, strebm, finish := job.StbrtSpbn(ctx, strebm, j)
	defer func() { finish(blert, err) }()

	return j.child.Run(ctx, clients, strebm)
}

func (j *keywordSebrchJob) Nbme() string {
	return "KeywordSebrchJob"
}

func (j *keywordSebrchJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		fbllthrough
	cbse job.VerbosityBbsic:
		res = bppend(res,
			bttribute.StringSlice("pbtterns", j.pbtterns),
		)
	}
	return res
}

func (j *keywordSebrchJob) Children() []job.Describer {
	return []job.Describer{j.child}
}

func (j *keywordSebrchJob) MbpChildren(fn job.MbpFunc) job.Job {
	cp := *j
	cp.child = job.Mbp(j.child, fn)
	return &cp
}
