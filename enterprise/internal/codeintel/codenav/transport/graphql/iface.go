package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
)

type CodeNavService interface {
	GetHover(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (_ string, _ types.Range, _ bool, err error)
	GetReferences(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState, cursor shared.ReferencesCursor) (_ []types.UploadLocation, nextCursor shared.ReferencesCursor, err error)
	GetImplementations(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState, cursor shared.ImplementationsCursor) (_ []types.UploadLocation, nextCursor shared.ImplementationsCursor, err error)
	GetDefinitions(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (_ []types.UploadLocation, err error)
	GetDiagnostics(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (diagnosticsAtUploads []shared.DiagnosticAtUpload, _ int, err error)
	GetRanges(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState, startLine, endLine int) (adjustedRanges []shared.AdjustedCodeIntelligenceRange, err error)
	GetStencil(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (adjustedRanges []types.Range, err error)

	// Uploads Service
	GetDumpsByIDs(ctx context.Context, ids []int) (_ []types.Dump, err error)
	GetClosestDumpsForBlob(ctx context.Context, repositoryID int, commit, path string, exactPath bool, indexer string) (_ []types.Dump, err error)
}

type GitserverClient = codenav.GitserverClient

type AutoIndexingService interface {
	sharedresolvers.AutoIndexingService

	QueueRepoRev(ctx context.Context, repositoryID int, rev string) error
}

type UploadsService = sharedresolvers.UploadsService

type PolicyService = sharedresolvers.PolicyService
