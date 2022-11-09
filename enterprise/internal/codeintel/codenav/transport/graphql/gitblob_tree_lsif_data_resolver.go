package graphql

import (
	"context"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type GitTreeLSIFDataResolver interface {
	LSIFUploads(ctx context.Context) ([]sharedresolvers.LSIFUploadResolver, error)
	Diagnostics(ctx context.Context, args *resolverstubs.LSIFDiagnosticsArgs) (DiagnosticConnectionResolver, error)
}
