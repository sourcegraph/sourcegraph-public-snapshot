package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
)

type Service interface {
	GetHover(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (_ string, _ shared.Range, _ bool, err error)
	GetReferences(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState, cursor shared.ReferencesCursor) (_ []shared.UploadLocation, nextCursor shared.ReferencesCursor, err error)
	GetImplementations(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState, cursor shared.ImplementationsCursor) (_ []shared.UploadLocation, nextCursor shared.ImplementationsCursor, err error)
	GetDefinitions(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (_ []shared.UploadLocation, err error)
	GetDiagnostics(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (diagnosticsAtUploads []shared.DiagnosticAtUpload, _ int, err error)
	GetRanges(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState, startLine, endLine int) (adjustedRanges []shared.AdjustedCodeIntelligenceRange, err error)
	GetStencil(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (adjustedRanges []shared.Range, err error)

	// Uploads Service
	GetDumpsByIDs(ctx context.Context, ids []int) (_ []shared.Dump, err error)
}
