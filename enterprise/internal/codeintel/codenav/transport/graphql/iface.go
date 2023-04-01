package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
)

type CodeNavService interface {
	GetHover(ctx context.Context, args codenav.RequestArgs, requestState codenav.RequestState) (_ string, _ types.Range, _ bool, err error)
	GetReferences(ctx context.Context, args codenav.RequestArgs, requestState codenav.RequestState, cursor codenav.ReferencesCursor) (_ []types.UploadLocation, nextCursor codenav.ReferencesCursor, err error)
	GetImplementations(ctx context.Context, args codenav.RequestArgs, requestState codenav.RequestState, cursor codenav.ImplementationsCursor) (_ []types.UploadLocation, nextCursor codenav.ImplementationsCursor, err error)
	GetDefinitions(ctx context.Context, args codenav.RequestArgs, requestState codenav.RequestState) (_ []types.UploadLocation, err error)
	GetDiagnostics(ctx context.Context, args codenav.RequestArgs, requestState codenav.RequestState) (diagnosticsAtUploads []codenav.DiagnosticAtUpload, _ int, err error)
	GetRanges(ctx context.Context, args codenav.RequestArgs, requestState codenav.RequestState, startLine, endLine int) (adjustedRanges []codenav.AdjustedCodeIntelligenceRange, err error)
	GetStencil(ctx context.Context, args codenav.RequestArgs, requestState codenav.RequestState) (adjustedRanges []types.Range, err error)
	GetClosestDumpsForBlob(ctx context.Context, repositoryID int, commit, path string, exactPath bool, indexer string) (_ []types.Dump, err error)
}

type AutoIndexingService interface {
	QueueRepoRev(ctx context.Context, repositoryID int, rev string) error
}

type UploadsService = sharedresolvers.UploadsService
