package search

import (
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/query"

	codeintelsearch "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/search"
	ownsearch "github.com/sourcegraph/sourcegraph/enterprise/internal/own/search"
)

func NewEnterpriseSearchJobs(
	codeIntelSearchService *codeintelsearch.Service,
) jobutil.EnterpriseJobs {
	return &enterpriseJobs{
		codeIntelSearchService: codeIntelSearchService,
	}
}

type enterpriseJobs struct {
	codeIntelSearchService *codeintelsearch.Service
}

func (e *enterpriseJobs) FileHasOwnerJob(child job.Job, features *search.Features, includeOwners, excludeOwners []string) job.Job {
	return ownsearch.NewFileHasOwnersJob(child, features, includeOwners, excludeOwners)
}

func (e *enterpriseJobs) SelectFileOwnerJob(child job.Job, features *search.Features) job.Job {
	return ownsearch.NewSelectOwnersJob(child, features)
}

func (e *enterpriseJobs) SymbolRelationshipSearchJob(relationship query.SymbolRelationship, rawSymbolSearch job.Job) job.Job {
	if e.codeIntelSearchService == nil {
		// Caller might not have codeintel services available, e.g. code monitors workers.
		// TODO as follow-up
		return jobutil.NewUnimplementedJob("`symbol:$relationship` searches are not available for this feature")
	}
	return codeintelsearch.NewSymbolRelationshipSearchJob(e.codeIntelSearchService, relationship, rawSymbolSearch)
}
