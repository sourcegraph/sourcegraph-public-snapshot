package graphql

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	sentinelSvc  *sentinel.Service
	autoindexSvc sharedresolvers.AutoIndexingService
	uploadSvc    sharedresolvers.UploadsService
	policySvc    sharedresolvers.PolicyService
	operations   *operations
}

func NewRootResolver(
	observationCtx *observation.Context,
	sentinelSvc *sentinel.Service,
	autoindexSvc sharedresolvers.AutoIndexingService,
	uploadSvc sharedresolvers.UploadsService,
	policySvc sharedresolvers.PolicyService,
) resolverstubs.SentinelServiceResolver {
	return &rootResolver{
		sentinelSvc:  sentinelSvc,
		autoindexSvc: autoindexSvc,
		uploadSvc:    uploadSvc,
		policySvc:    policySvc,
		operations:   newOperations(observationCtx),
	}
}
