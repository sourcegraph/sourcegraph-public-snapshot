package lsifstore

import (
	"context"
	"fmt"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codegraph"
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
	SCIPDocuments(ctx context.Context, uploadID int, paths []core.UploadRelPath) (map[core.UploadRelPath]*scip.Document, error)

	// Fetch symbol names by position
	GetMonikersByPosition(ctx context.Context, uploadID int, path core.UploadRelPath, line, character int) ([][]precise.MonikerData, error)
	GetPackageInformation(ctx context.Context, uploadID int, packageInformationID string) (precise.PackageInformationData, bool, error)

	// Fetch usages by position
	GetSymbolUsages(ctx context.Context, options SymbolUsagesOptions) (_ []shared.Usage, totalCount int, err error)

	// Metadata by position
	GetHover(ctx context.Context, bundleID int, path core.UploadRelPath, line, character int) (string, shared.Range, bool, error)
	GetDiagnostics(ctx context.Context, bundleID int, prefix core.UploadRelPath, limit, offset int) ([]shared.Diagnostic[core.UploadRelPath], int, error)

	// Extraction methods
	ExtractDefinitionLocationsFromPosition(context.Context, FindUsagesKey) ([]shared.UsageBuilder, []string, error)
	ExtractReferenceLocationsFromPosition(context.Context, FindUsagesKey) ([]shared.UsageBuilder, []string, error)
	ExtractImplementationLocationsFromPosition(context.Context, FindUsagesKey) ([]shared.UsageBuilder, []string, error)
	ExtractPrototypeLocationsFromPosition(context.Context, FindUsagesKey) ([]shared.UsageBuilder, []string, error)
}

type SymbolUsagesOptions struct {
	shared.UsageKind
	UploadIDs           []int
	SkipPathsByUploadID map[int]string
	LookupSymbols       []string
	Limit               int
	Offset              int
}

func (opts SymbolUsagesOptions) Attrs() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("usageKind", opts.UsageKind.String()),
		attribute.Int("numUploadIDs", len(opts.UploadIDs)),
		attribute.IntSlice("uploadIDs", opts.UploadIDs),
		attribute.Int("numSkipPathsByID", len(opts.SkipPathsByUploadID)),
		attribute.String("skipPathsByID", fmt.Sprintf("%v", opts.SkipPathsByUploadID)),
		attribute.Int("numLookupSymbols", len(opts.LookupSymbols)),
		attribute.StringSlice("lookupSymbols", opts.LookupSymbols),
		attribute.Int("limit", opts.Limit),
		attribute.Int("offset", opts.Offset),
	}
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

type MatchStrategy string

const (
	// SinglePositionBasedMatchStrategy represents the legacy mode of finding matching
	// occurrences which intersect with a given source position
	SinglePositionBasedMatchStrategy MatchStrategy = "single-position based"
	// RangeBasedMatchStrategy represents exact matching based on purely source range.
	// In the general case, this may lead to multiple different semantic occurrences
	// being found at the same range.
	RangeBasedMatchStrategy = "range based"
	// RangeAndSymbolBasedMatchStrategy represents the combination of RangeBasedMatchStrategy
	// with exact matching based on symbol name. Generally, this should return
	// at most one occurrence, but exceptions are possible if indexer output
	// is sub-optimal and the associated occurrences cannot be merged.
	RangeAndSymbolBasedMatchStrategy = "range and symbol based"
)

// IdentifyMatchingOccurrences will generally give a small list of occurrences
// since there are usually not a large number of precise occurrences at the
// same location. Known exceptions to this:
//   - C++ macros, where scip-clang currently emits all defs/refs at the macro name range.
//   - Ruby attr_accessor etc., where scip-ruby can emit defs for getters and setters at the same range.
//     (i.e. special-cased AST rewriting in Sorbet, instead of a general macro facility)
//
// The returned 'strategy' represents the algorithm used for matching.
func (key *FindUsagesKey) IdentifyMatchingOccurrences(allOccurrences []*scip.Occurrence) (out []*scip.Occurrence, strategy MatchStrategy) {
	if startPos, ok := key.Matcher.PositionBased(); ok {
		// Preserve back-compat with older APIs providing just a single position.
		strategy = SinglePositionBasedMatchStrategy
		out = scip.FindOccurrences(allOccurrences, startPos.Line, startPos.Character)
		return
	}
	optSymbolToMatch, range_, ok := key.Matcher.SymbolBased()
	if !ok {
		panic(fmt.Sprintf("Unhandled case of locationKey.Matcher: %+v", key.Matcher))
	}
	strategy = RangeBasedMatchStrategy
	sameRangeOccs := codegraph.FindOccurrencesWithEqualRange(allOccurrences, range_)
	if len(sameRangeOccs) == 0 {
		return
	}
	symbolToMatch, isSome := optSymbolToMatch.Get()
	if !isSome {
		out = sameRangeOccs
		return
	}
	strategy = RangeAndSymbolBasedMatchStrategy
	for _, occ := range sameRangeOccs {
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
