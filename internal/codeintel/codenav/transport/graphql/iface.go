package graphql

import (
	"context"

	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
)

type Service interface {
	GetHover(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (_ string, _ shared.Range, _ bool, err error)
	GetReferences(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState, cursor shared.ReferencesCursor) (_ []shared.UploadLocation, nextCursor shared.ReferencesCursor, err error)
	GetImplementations(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState, cursor shared.ImplementationsCursor) (_ []shared.UploadLocation, nextCursor shared.ImplementationsCursor, err error)
	GetDefinitions(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (_ []shared.UploadLocation, err error)
	GetDiagnostics(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (diagnosticsAtUploads []shared.DiagnosticAtUpload, _ int, err error)
	GetRanges(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState, startLine, endLine int) (adjustedRanges []shared.AdjustedCodeIntelligenceRange, err error)
	GetStencil(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState) (adjustedRanges []shared.Range, err error)

	// Symbols client
	GetSupportedByCtags(ctx context.Context, filepath string, repoName api.RepoName) (bool, string, error)

	// Language Support
	GetLanguagesRequestedBy(ctx context.Context, userID int) (_ []string, err error)
	SetRequestLanguageSupport(ctx context.Context, userID int, language string) (err error)

	// Uploads Service
	GetDumpsByIDs(ctx context.Context, ids []int) (_ []shared.Dump, err error)
	GetClosestDumpsForBlob(ctx context.Context, repositoryID int, commit, path string, exactPath bool, indexer string) (_ []shared.Dump, err error)
}

type GitserverClient interface {
	CommitsExist(ctx context.Context, commits []gitserver.RepositoryCommit) ([]bool, error)
	DiffPath(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, sourceCommit, targetCommit, path string) ([]*diff.Hunk, error)
}
