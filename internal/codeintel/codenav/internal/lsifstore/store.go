package lsifstore

import (
	"context"
	"fmt"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/collections"

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
	SCIPDocument(ctx context.Context, uploadID int, path core.UploadRelPath) (core.Option[*scip.Document], error)

	// Fetch symbol names by position
	GetMonikersByPosition(ctx context.Context, uploadID int, path core.UploadRelPath, line, character int) ([][]precise.MonikerData, error)
	GetPackageInformation(ctx context.Context, uploadID int, packageInformationID string) (precise.PackageInformationData, bool, error)

	// Fetch locations by position
	GetBulkMonikerLocations(ctx context.Context, tableName string, uploadIDs []int, monikers []precise.MonikerData, limit, offset int) ([]shared.Location, int, error)
	GetMinimalBulkMonikerLocations(ctx context.Context, usageKind shared.UsageKind, uploadIDs []int, skipPaths map[int]string, monikers []precise.MonikerData, limit, offset int) (_ []shared.Usage, totalCount int, err error)

	// Metadata by position
	GetHover(ctx context.Context, bundleID int, path core.UploadRelPath, line, character int) (string, shared.Range, bool, error)
	GetDiagnostics(ctx context.Context, bundleID int, prefix core.UploadRelPath, limit, offset int) ([]shared.Diagnostic[core.UploadRelPath], int, error)

	// Extraction methods
	ExtractDefinitionLocationsFromPosition(context.Context, FindUsagesKey) ([]shared.UsageBuilder, []string, error)
	ExtractReferenceLocationsFromPosition(context.Context, FindUsagesKey) ([]shared.UsageBuilder, []string, error)
	ExtractImplementationLocationsFromPosition(context.Context, FindUsagesKey) ([]shared.UsageBuilder, []string, error)
	ExtractPrototypeLocationsFromPosition(context.Context, FindUsagesKey) ([]shared.UsageBuilder, []string, error)
}

type LocationKey struct {
	UploadID  int
	Path      core.UploadRelPath
	Line      int
	Character int
}

// FindUsagesKey represents a specific scip.Document in the database,
// along with a matching object to identify usages for a symbol
// in that Document.
type FindUsagesKey struct {
	// UploadID is the upload to locate the Document.
	UploadID int
	// Path is the exact path within the upload, which may be different
	// from the path from the root of the repository, if the index
	// itself was generated for a subdirectory.
	Path core.UploadRelPath
	// Matcher describes how to find usages within a specific Document.
	Matcher shared.Matcher
}

type MatchingKind string

const (
	SinglePositionBasedMatching MatchingKind = "single-position based"
	RangeBasedMatching                       = "range based"
	RangeAndSymbolBasedMatching              = "range and symbol based"
)

// IdentifyMatchingOccurrences will generally give a small list of occurrences
// since there are usually not a large number of precise occurrences at the
// same location. Known exceptions to this:
// - C++ macros, where scip-clang currently emits all defs/refs at the macro name range.
func (key *FindUsagesKey) IdentifyMatchingOccurrences(allOccurrences []*scip.Occurrence) (out []*scip.Occurrence, kind MatchingKind) {
	if startPos, ok := key.Matcher.PositionBased(); ok {
		// Preserve back-compat with older APIs providing just a single position.
		kind = SinglePositionBasedMatching
		out = scip.FindOccurrences(allOccurrences, startPos.Line, startPos.Character)
		return
	}
	symbolToMatch, range_, ok := key.Matcher.SymbolBased()
	if !ok {
		panic(fmt.Sprintf("Unhandled case of locationKey.Matcher: %+v", key.Matcher))
	}
	kind = RangeBasedMatching
	idxRange := collections.BinarySearchRangeFunc(allOccurrences, range_, func(occ *scip.Occurrence, r scip.Range) int {
		return scip.NewRangeUnchecked(occ.Range).Compare(range_)
	})
	if idxRange.IsEmpty() {
		return
	}
	if symbolToMatch == "" {
		out = allOccurrences[idxRange.Start:idxRange.End]
		return
	}
	kind = RangeAndSymbolBasedMatching
	for _, occ := range allOccurrences[idxRange.Start:idxRange.End] {
		if occ.Symbol == symbolToMatch {
			out = append(out, occ)
		}
	}
	return
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
