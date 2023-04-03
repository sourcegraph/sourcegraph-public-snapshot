package lsifstore

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type LsifStore interface {
	// Hover
	GetHover(ctx context.Context, bundleID int, path string, line, character int) (string, types.Range, bool, error)

	// References
	GetReferenceLocations(ctx context.Context, uploadID int, path string, line, character, limit, offset int) ([]shared.Location, int, error)

	// Implementation
	GetImplementationLocations(ctx context.Context, uploadID int, path string, line, character, limit, offset int) ([]shared.Location, int, error)

	// Definition
	GetDefinitionLocations(ctx context.Context, uploadID int, path string, line, character, limit, offset int) ([]shared.Location, int, error)

	// Monikers
	GetMonikersByPosition(ctx context.Context, uploadID int, path string, line, character int) ([][]precise.MonikerData, error)
	GetBulkMonikerLocations(ctx context.Context, tableName string, uploadIDs []int, monikers []precise.MonikerData, limit, offset int) ([]shared.Location, int, error)

	// Packages
	GetPackageInformation(ctx context.Context, uploadID int, path, packageInformationID string) (precise.PackageInformationData, bool, error)

	// Diagnostics
	GetDiagnostics(ctx context.Context, bundleID int, prefix string, limit, offset int) ([]shared.Diagnostic, int, error)

	// Stencil
	GetStencil(ctx context.Context, bundleID int, path string) ([]types.Range, error)

	// Ranges
	GetRanges(ctx context.Context, bundleID int, path string, startLine, endLine int) ([]shared.CodeIntelligenceRange, error)

	// Paths
	GetPathExists(ctx context.Context, bundleID int, path string) (bool, error)
}

type store struct {
	db         *basestore.Store
	operations *operations
}

func New(observationCtx *observation.Context, db codeintelshared.CodeIntelDB) LsifStore {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		operations: newOperations(observationCtx),
	}
}
