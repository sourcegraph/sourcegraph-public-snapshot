package lsifstore

import (
	"context"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	codeintelshared "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type LsifStore interface {
	// Multi-upload data

	// FindDocumentIDs checks one path per upload, and returns the document IDs
	// for those paths. If needed, we could generalize this to allow multiple paths
	// per uploadID.
	FindDocumentIDs(ctx context.Context, uploadIDToLookupPath map[int]core.UploadRelPath) (uploadIDToDocumentID map[int]int, _ error)

	// Whole-document data
	GetStencil(ctx context.Context, bundleID int, path core.UploadRelPath) ([]shared.Range, error)
	GetRanges(ctx context.Context, bundleID int, path core.UploadRelPath, startLine, endLine int) ([]shared.CodeIntelligenceRange, error)
	SCIPDocument(ctx context.Context, uploadID int, path core.UploadRelPath) (_ *scip.Document, err error)

	// Fetch symbol names by position
	GetMonikersByPosition(ctx context.Context, uploadID int, path core.UploadRelPath, line, character int) ([][]precise.MonikerData, error)
	GetPackageInformation(ctx context.Context, uploadID int, packageInformationID string) (precise.PackageInformationData, bool, error)

	// Fetch locations by position
	GetBulkMonikerLocations(ctx context.Context, tableName string, uploadIDs []int, monikers []precise.MonikerData, limit, offset int) ([]shared.Location, int, error)
	GetMinimalBulkMonikerLocations(ctx context.Context, tableName string, uploadIDs []int, skipPaths map[int]string, monikers []precise.MonikerData, limit, offset int) (_ []shared.Location, totalCount int, err error)

	// Metadata by position
	GetHover(ctx context.Context, bundleID int, path core.UploadRelPath, line, character int) (string, shared.Range, bool, error)
	GetDiagnostics(ctx context.Context, bundleID int, prefix core.UploadRelPath, limit, offset int) ([]shared.Diagnostic[core.UploadRelPath], int, error)

	// Extraction methods
	ExtractDefinitionLocationsFromPosition(ctx context.Context, locationKey LocationKey) ([]shared.Location, []string, error)
	ExtractReferenceLocationsFromPosition(ctx context.Context, locationKey LocationKey) ([]shared.Location, []string, error)
	ExtractImplementationLocationsFromPosition(ctx context.Context, locationKey LocationKey) ([]shared.Location, []string, error)
	ExtractPrototypeLocationsFromPosition(ctx context.Context, locationKey LocationKey) ([]shared.Location, []string, error)
}

type LocationKey struct {
	UploadID  int
	Path      core.UploadRelPath
	Line      int
	Character int
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
