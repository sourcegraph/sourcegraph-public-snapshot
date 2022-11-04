package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
)

type DiagnosticConnectionResolver interface {
	Nodes(ctx context.Context) ([]DiagnosticResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*PageInfo, error)
}

type diagnosticConnectionResolver struct {
	diagnostics      []shared.DiagnosticAtUpload
	totalCount       int
	locationResolver *sharedresolvers.CachedLocationResolver
}

func NewDiagnosticConnectionResolver(diagnostics []shared.DiagnosticAtUpload, totalCount int, locationResolver *sharedresolvers.CachedLocationResolver) DiagnosticConnectionResolver {
	return &diagnosticConnectionResolver{
		diagnostics:      diagnostics,
		totalCount:       totalCount,
		locationResolver: locationResolver,
	}
}

func (r *diagnosticConnectionResolver) Nodes(ctx context.Context) ([]DiagnosticResolver, error) {
	resolvers := make([]DiagnosticResolver, 0, len(r.diagnostics))
	for i := range r.diagnostics {
		resolvers = append(resolvers, NewDiagnosticResolver(r.diagnostics[i], r.locationResolver))
	}
	return resolvers, nil
}

func (r *diagnosticConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(r.totalCount), nil
}

func (r *diagnosticConnectionResolver) PageInfo(ctx context.Context) (*PageInfo, error) {
	return HasNextPage(len(r.diagnostics) < r.totalCount), nil
}
