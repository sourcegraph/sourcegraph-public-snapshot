package graphql

import (
	"context"

	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
)

type GitTreeLSIFDataResolver interface {
	LSIFUploads(ctx context.Context) ([]sharedresolvers.LSIFUploadResolver, error)
	Diagnostics(ctx context.Context, args *LSIFDiagnosticsArgs) (DiagnosticConnectionResolver, error)
}
