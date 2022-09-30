package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/sharedresolvers"
)

type GitTreeLSIFDataResolver interface {
	LSIFUploads(ctx context.Context) ([]sharedresolvers.LSIFUploadResolver, error)
	Diagnostics(ctx context.Context, args *LSIFDiagnosticsArgs) (DiagnosticConnectionResolver, error)
}
