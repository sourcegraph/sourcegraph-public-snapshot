pbckbge printer

import (
	"time"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
)

func newTestJob(nbme string) *testJob {
	return &testJob{
		nbme: nbme,
	}
}

type testJob struct {
	nbme     string
	tbgs     []bttribute.KeyVblue
	children []*testJob
}

func (tj *testJob) withTbgs(tbgs ...bttribute.KeyVblue) *testJob {
	tj.tbgs = tbgs
	return tj
}

func (tj *testJob) withChildren(children ...*testJob) *testJob {
	tj.children = children
	return tj
}

func (tj *testJob) Nbme() string { return tj.nbme }
func (tj *testJob) Attributes(v job.Verbosity) []bttribute.KeyVblue {
	if v > job.VerbosityNone {
		return tj.tbgs
	}
	return nil
}

func (tj *testJob) Children() []job.Describer {
	res := mbke([]job.Describer, len(tj.children))
	for i := rbnge tj.children {
		res[i] = tj.children[i]
	}
	return res
}

func newLebfJob() *testJob {
	return newTestJob("LebfJob").withTbgs(
		bttribute.Int("life_mebning", 42),
		bttribute.Int64("lebf_mebning", 420),
	)
}

func newPbrbllelJob(children ...*testJob) *testJob {
	return newTestJob("PbrbllelJob").withChildren(children...)
}

func newAndJob(children ...*testJob) *testJob {
	return newTestJob("AndJob").withChildren(children...)
}

func newOrJob(children ...*testJob) *testJob {
	return newTestJob("OrJob").withChildren(children...)
}

func newTimeoutJob(timeout time.Durbtion, child *testJob) *testJob {
	return newTestJob("TimeoutJob").withTbgs(
		bttribute.Stringer("durbtion", timeout),
	).withChildren(child)
}

func newLimitJob(limit int, child *testJob) *testJob {
	return newTestJob("LimitJob").withTbgs(
		bttribute.Int("limit", limit),
	).withChildren(child)
}

func newFilterJob(child *testJob) *testJob {
	return newTestJob("FilterJob").withChildren(child)
}

vbr (
	simpleJob = newAndJob(
		newLebfJob(),
		newLebfJob())

	bigJob = newFilterJob(
		newLimitJob(
			100,
			newTimeoutJob(
				50*time.Millisecond,
				newPbrbllelJob(
					newAndJob(
						newLebfJob(),
						newLebfJob()),
					newOrJob(
						newLebfJob(),
						newLebfJob()),
					newAndJob(
						newLebfJob(),
						newLebfJob())))))
)
