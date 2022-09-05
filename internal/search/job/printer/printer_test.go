package printer

import (
	"time"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/search/job"
)

func newTestJob(name string) *testJob {
	return &testJob{
		name: name,
	}
}

type testJob struct {
	name     string
	tags     []otlog.Field
	children []*testJob
}

func (tj *testJob) withTags(tags ...otlog.Field) *testJob {
	tj.tags = tags
	return tj
}

func (tj *testJob) withChildren(children ...*testJob) *testJob {
	tj.children = children
	return tj
}

func (tj *testJob) Name() string { return tj.name }
func (tj *testJob) Fields(v job.Verbosity) []otlog.Field {
	if v > job.VerbosityNone {
		return tj.tags
	}
	return nil
}

func (tj *testJob) Children() []job.Describer {
	res := make([]job.Describer, len(tj.children))
	for i := range tj.children {
		res[i] = tj.children[i]
	}
	return res
}

func newLeafJob() *testJob {
	return newTestJob("LeafJob").withTags(
		otlog.Int32("life_meaning", 42),
		otlog.Int64("leaf_meaning", 420),
	)
}

func newParallelJob(children ...*testJob) *testJob {
	return newTestJob("ParallelJob").withChildren(children...)
}

func newAndJob(children ...*testJob) *testJob {
	return newTestJob("AndJob").withChildren(children...)
}

func newOrJob(children ...*testJob) *testJob {
	return newTestJob("OrJob").withChildren(children...)
}

func newTimeoutJob(timeout time.Duration, child *testJob) *testJob {
	return newTestJob("TimeoutJob").withTags(
		otlog.String("duration", timeout.String()),
	).withChildren(child)
}

func newLimitJob(limit int, child *testJob) *testJob {
	return newTestJob("LimitJob").withTags(
		otlog.Int("limit", limit),
	).withChildren(child)
}

func newFilterJob(child *testJob) *testJob {
	return newTestJob("FilterJob").withChildren(child)
}

var (
	simpleJob = newAndJob(
		newLeafJob(),
		newLeafJob())

	bigJob = newFilterJob(
		newLimitJob(
			100,
			newTimeoutJob(
				50*time.Millisecond,
				newParallelJob(
					newAndJob(
						newLeafJob(),
						newLeafJob()),
					newOrJob(
						newLeafJob(),
						newLeafJob()),
					newAndJob(
						newLeafJob(),
						newLeafJob())))))
)
