package search

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
)

func NewEnterpriseSearchJobs() jobutil.EnterpriseJobs {
	return &enterpriseJobs{}
}

type enterpriseJobs struct{}

func (e *enterpriseJobs) FileHasOwnerJob(child job.Job, includeOwners, excludeOwners []string) job.Job {
	return search.NewFileHasOwnersJob(child, includeOwners, excludeOwners)
}

func (e *enterpriseJobs) SelectFileOwnerJob(child job.Job) job.Job {
	return search.NewSelectOwnersJob(child)
}
