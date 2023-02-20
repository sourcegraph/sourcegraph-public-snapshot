package graphql

import (
	"context"
	"errors"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	sentinelSvc *sentinel.Service
	operations  *operations
}

func NewRootResolver(observationCtx *observation.Context, sentinelSvc *sentinel.Service) resolverstubs.SentinelServiceResolver {
	return &rootResolver{
		sentinelSvc: sentinelSvc,
		operations:  newOperations(observationCtx),
	}
}

func (r *rootResolver) Vulnerabilities(ctx context.Context, args resolverstubs.GetVulnerabilitiesArgs) (_ resolverstubs.VulnerabilityConnectionResolver, err error) {
	ctx, _, endObservation := r.operations.getVulnerabilities.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	return nil, errors.New("unimplemented") // TODO
}

func (r *rootResolver) VulnerabilityMatches(ctx context.Context, args resolverstubs.GetVulnerabilityMatchesArgs) (_ resolverstubs.VulnerabilityMatchConnectionResolver, err error) {
	ctx, _, endObservation := r.operations.getMatches.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	return nil, errors.New("unimplemented") // TODO
}
