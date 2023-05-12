package printer

import (
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/search/job"
)

func newTestJob(name string) *testJob {
	return &testJob{
		name: name,
	}
}

type testJob struct {
	name     string
	tags     []attribute.KeyValue
	children []*testJob
}

func (tj *testJob) withTags(tags ...attribute.KeyValue) *testJob {
	tj.tags = tags
	return tj
}

func (tj *testJob) withChildren(children ...*testJob) *testJob {
	tj.children = children
	return tj
}

func (tj *testJob) Name() string { return tj.name }
func (tj *testJob) Attributes(v job.Verbosity) []attribute.KeyValue {
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
		attribute.Int("life_meaning", 42),
		attribute.Int64("leaf_meaning", 420),
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
		attribute.Stringer("duration", timeout),
	).withChildren(child)
}

func newLimitJob(limit int, child *testJob) *testJob {
	return newTestJob("LimitJob").withTags(
		attribute.Int("limit", limit),
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
