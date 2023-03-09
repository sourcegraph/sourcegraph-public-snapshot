package search

import (
	codeintelsearch "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/search"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/search"

	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
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

func (e *enterpriseJobs) FileHasOwnerJob(child job.Job, includeOwners, excludeOwners []string) job.Job {
	return search.NewFileHasOwnersJob(child, includeOwners, excludeOwners)
}

func (e *enterpriseJobs) SelectFileOwnerJob(child job.Job) job.Job {
	return search.NewSelectOwnersJob(child)
}

func (e *enterpriseJobs) SymbolRelationshipSearchJob(relationship query.SymbolRelationship, rawSymbolSearch job.Job) job.Job {
	if e.codeIntelSearchService == nil {
		// Caller might not have codeintel services available, e.g. code monitors workers.
		return jobutil.NewUnimplementedJob("symbol:$relationship searches are not available for this feature")
	}
	return codeintelsearch.NewSymbolRelationshipSearchJob(e.codeIntelSearchService, relationship, rawSymbolSearch)
}
