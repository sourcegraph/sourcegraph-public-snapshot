package search

import (
	ownsearch "github.com/sourcegraph/sourcegraph/enterprise/internal/own/search"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
)

func NewEnterpriseSearchJobs() jobutil.EnterpriseJobs {
	return &enterpriseJobs{}
}

type enterpriseJobs struct{}

func (e *enterpriseJobs) FileHasOwnerJob(child job.Job, features *search.Features, includeOwners, excludeOwners []string) job.Job {
	return ownsearch.NewFileHasOwnersJob(child, features, includeOwners, excludeOwners)
}

func (e *enterpriseJobs) SelectFileOwnerJob(child job.Job, features *search.Features) job.Job {
	return ownsearch.NewSelectOwnersJob(child, features)
}
