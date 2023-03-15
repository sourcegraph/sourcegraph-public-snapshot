package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type diagnosticConnectionResolver struct {
	diagnostics      []shared.DiagnosticAtUpload
	totalCount       int
	locationResolver *sharedresolvers.CachedLocationResolver
}

func NewDiagnosticConnectionResolver(diagnostics []shared.DiagnosticAtUpload, totalCount int, locationResolver *sharedresolvers.CachedLocationResolver) resolverstubs.DiagnosticConnectionResolver {
	return &diagnosticConnectionResolver{
		diagnostics:      diagnostics,
		totalCount:       totalCount,
		locationResolver: locationResolver,
	}
}

func (r *diagnosticConnectionResolver) Nodes(ctx context.Context) ([]resolverstubs.DiagnosticResolver, error) {
	resolvers := make([]resolverstubs.DiagnosticResolver, 0, len(r.diagnostics))
	for i := range r.diagnostics {
		resolvers = append(resolvers, NewDiagnosticResolver(r.diagnostics[i], r.locationResolver))
	}
	return resolvers, nil
}

func (r *diagnosticConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(r.totalCount), nil
}

func (r *diagnosticConnectionResolver) PageInfo(ctx context.Context) (resolverstubs.PageInfo, error) {
	return HasNextPage(len(r.diagnostics) < r.totalCount), nil
}
