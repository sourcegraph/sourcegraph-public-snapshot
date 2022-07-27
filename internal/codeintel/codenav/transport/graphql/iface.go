package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type Service interface {
	GetHover(ctx context.Context, bundleID int, path string, line, character int) (string, shared.Range, bool, error)
	GetReferences(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState, cursor shared.ReferencesCursor) (_ []shared.UploadLocation, nextCursor shared.ReferencesCursor, err error)
	GetImplementations(ctx context.Context, args shared.RequestArgs, requestState codenav.RequestState, cursor shared.ImplementationsCursor) (_ []shared.UploadLocation, nextCursor shared.ImplementationsCursor, err error)
	GetDefinitions(ctx context.Context, uploadID int, path string, line int, character int, limit int, offset int) (_ []shared.Location, totalCount int, err error)
	GetDiagnostics(ctx context.Context, bundleID int, prefix string, limit, offset int) (_ []shared.Diagnostic, _ int, err error)
	GetRanges(ctx context.Context, bundleID int, path string, startLine, endLine int) (_ []shared.CodeIntelligenceRange, err error)
	GetStencil(ctx context.Context, uploadID int, path string) (_ []shared.Range, err error)

	GetMonikersByPosition(ctx context.Context, bundleID int, path string, line, character int) (_ [][]precise.MonikerData, err error)
	GetBulkMonikerLocations(ctx context.Context, tableName string, uploadIDs []int, monikers []precise.MonikerData, limit, offset int) (_ []shared.Location, _ int, err error)
	GetPackageInformation(ctx context.Context, bundleID int, path, packageInformationID string) (_ precise.PackageInformationData, _ bool, err error)

	// Uploads Service
	GetDumpsByIDs(ctx context.Context, ids []int) (_ []shared.Dump, err error)
	GetUploadsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QualifiedMonikerData) (_ []shared.Dump, err error)
	GetUploadIDsWithReferences(ctx context.Context, orderedMonikers []precise.QualifiedMonikerData, ignoreIDs []int, repositoryID int, commit string, limit int, offset int) (ids []int, recordsScanned int, totalCount int, err error)
}

type GitserverClient interface {
	CommitsExist(ctx context.Context, commits []gitserver.RepositoryCommit) ([]bool, error)
}
