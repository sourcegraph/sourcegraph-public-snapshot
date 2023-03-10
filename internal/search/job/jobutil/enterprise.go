package jobutil

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type EnterpriseJobs interface {
	FileHasOwnerJob(child job.Job, features *search.Features, includeOwners, excludeOwners []string) job.Job
	SelectFileOwnerJob(child job.Job, features *search.Features) job.Job

	SymbolRelationshipSearchJob(relationship query.SymbolRelationship, rawSymbolSearch job.Job) job.Job
}

func NewUnimplementedEnterpriseJobs() EnterpriseJobs {
	return &enterpriseJobs{}
}

type enterpriseJobs struct{}

func (e *enterpriseJobs) FileHasOwnerJob(child job.Job, features *search.Features, includeOwners, excludeOwners []string) job.Job {
	return NewUnimplementedJob("`file:has.owner` searches are not available on this instance")
}

func (e *enterpriseJobs) SelectFileOwnerJob(child job.Job, features *search.Features) job.Job {
	return NewUnimplementedJob("`select:file.owners` searches are not available on this instance")
}

func (e *enterpriseJobs) SymbolRelationshipSearchJob(relationship query.SymbolRelationship, rawSymbolSearch job.Job) job.Job {
	return NewUnimplementedJob("symbol:$relationship searches are not available on this instance")
}

func NewUnimplementedJob(msg string) *UnimplementedJob {
	return &UnimplementedJob{msg: msg}
}

type UnimplementedJob struct {
	msg string
}

func (e *UnimplementedJob) Run(context.Context, job.RuntimeClients, streaming.Sender) (*search.Alert, error) {
	return nil, errors.New(e.msg)
}

func (e *UnimplementedJob) Name() string                     { return "UnimplementedJob" }
func (e *UnimplementedJob) Fields(job.Verbosity) []log.Field { return nil }
func (e *UnimplementedJob) Children() []job.Describer        { return nil }
func (e *UnimplementedJob) MapChildren(job.MapFunc) job.Job  { return e }
