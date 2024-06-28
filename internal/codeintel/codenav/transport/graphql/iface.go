package graphql

import (
	"context"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type CodeNavService interface {
	GetHover(ctx context.Context, args codenav.PositionalRequestArgs, requestState codenav.RequestState) (_ string, _ shared.Range, _ bool, err error)
	GetReferences(ctx context.Context, args codenav.PositionalRequestArgs, requestState codenav.RequestState, cursor codenav.Cursor) (_ []shared.UploadLocation, nextCursor codenav.Cursor, err error)
	GetImplementations(ctx context.Context, args codenav.PositionalRequestArgs, requestState codenav.RequestState, cursor codenav.Cursor) (_ []shared.UploadLocation, nextCursor codenav.Cursor, err error)
	GetPrototypes(ctx context.Context, args codenav.PositionalRequestArgs, requestState codenav.RequestState, cursor codenav.Cursor) (_ []shared.UploadLocation, nextCursor codenav.Cursor, err error)
	GetDefinitions(ctx context.Context, args codenav.PositionalRequestArgs, requestState codenav.RequestState, cursor codenav.Cursor) (_ []shared.UploadLocation, nextCursor codenav.Cursor, err error)
	GetDiagnostics(ctx context.Context, args codenav.PositionalRequestArgs, requestState codenav.RequestState) (diagnosticsAtUploads []codenav.DiagnosticAtUpload, _ int, err error)
	GetRanges(ctx context.Context, args codenav.PositionalRequestArgs, requestState codenav.RequestState, startLine, endLine int) (adjustedRanges []codenav.AdjustedCodeIntelligenceRange, err error)
	GetStencil(ctx context.Context, args codenav.PositionalRequestArgs, requestState codenav.RequestState) (adjustedRanges []shared.Range, err error)
	// The resulting uploads are guaranteed to be unique per (indexer, root) pair,
	// see NOTE(id: closest-uploads-postcondition).
	GetClosestCompletedUploadsForBlob(context.Context, uploadsshared.UploadMatchingOptions) (_ []uploadsshared.CompletedUpload, err error)
	VisibleUploadsForPath(ctx context.Context, requestState codenav.RequestState) ([]uploadsshared.CompletedUpload, error)
	SnapshotForDocument(ctx context.Context, repositoryID int, commit string, path core.RepoRelPath, uploadID int) (data []shared.SnapshotData, err error)
	SCIPDocument(ctx context.Context, uploadID int, path core.RepoRelPath) (*scip.Document, error)
	SyntacticUsages(ctx context.Context, repo types.Repo, commit api.CommitID, path core.RepoRelPath, symbolRange scip.Range) (codenav.SyntacticUsagesResult, codenav.PreviousSyntacticSearch, *codenav.SyntacticUsagesError)
	SearchBasedUsages(ctx context.Context, repo types.Repo, commit api.CommitID, path core.RepoRelPath, symbolRange scip.Range, previous codenav.PreviousSyntacticSearch) ([]codenav.SearchBasedMatch, error)
}

type AutoIndexingService interface {
	QueueRepoRev(ctx context.Context, repositoryID int, rev string) error
}
