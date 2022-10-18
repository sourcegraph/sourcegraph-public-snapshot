package graphql

import (
	"context"
	"time"

	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	autoindexingShared "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type Service interface {
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

	GetUnsafeDB() database.DB
}

type GitserverClient interface {
	CommitsExist(ctx context.Context, commits []gitserver.RepositoryCommit) ([]bool, error)
	DiffPath(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, sourceCommit, targetCommit, path string) ([]*diff.Hunk, error)
}

type AutoIndexingService interface {
	GetIndexes(ctx context.Context, opts autoindexingShared.GetIndexesOptions) (_ []types.Index, _ int, err error)
	GetIndexByID(ctx context.Context, id int) (_ types.Index, _ bool, err error)
	GetIndexesByIDs(ctx context.Context, ids ...int) (_ []types.Index, err error)
	GetUnsafeDB() database.DB
	GetListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error)
	QueueRepoRev(ctx context.Context, repositoryID int, rev string) error
}

type UploadsService interface {
	GetUploads(ctx context.Context, opts uploadsshared.GetUploadsOptions) (uploads []types.Upload, totalCount int, err error)
	GetAuditLogsForUpload(ctx context.Context, uploadID int) (_ []types.UploadLog, err error)
	GetListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error)
	GetUploadDocumentsForPath(ctx context.Context, bundleID int, pathPattern string) ([]string, int, error)
	GetUploadsByIDs(ctx context.Context, ids ...int) (_ []types.Upload, err error)
}

type PolicyService interface {
	GetRetentionPolicyOverview(ctx context.Context, upload types.Upload, matchesOnly bool, first int, after int64, query string, now time.Time) (matches []types.RetentionPolicyMatchCandidate, totalCount int, err error)
}
