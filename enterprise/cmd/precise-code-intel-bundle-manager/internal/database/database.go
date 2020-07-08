package database

import (
	"context"
	"fmt"
	"sort"

	"github.com/opentracing/opentracing-go/ext"
	pkgerrors "github.com/pkg/errors"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// Database wraps access to a single processed bundle.
type Database interface {
	// Close closes the underlying reader.
	Close() error

	// Exists determines if the path exists in the database.
	Exists(ctx context.Context, path string) (bool, error)

	// Ranges returns definition, reference, and hover data for each range within the given span of lines.
	Ranges(ctx context.Context, path string, startLine, endLine int) ([]bundles.CodeIntelligenceRange, error)

	// Definitions returns the set of locations defining the symbol at the given position.
	Definitions(ctx context.Context, path string, line, character int) ([]bundles.Location, error)

	// References returns the set of locations referencing the symbol at the given position.
	References(ctx context.Context, path string, line, character int) ([]bundles.Location, error)

	// Hover returns the hover text of the symbol at the given position.
	Hover(ctx context.Context, path string, line, character int) (string, bundles.Range, bool, error)

	// Diagnostics returns the diagnostics for the documents that have the given path prefix. This method
	// also returns the size of the complete result set to aid in pagination (along with skip and take).
	Diagnostics(ctx context.Context, prefix string, skip, take int) ([]bundles.Diagnostic, int, error)

	// MonikersByPosition returns all monikers attached ranges containing the given position. If multiple
	// ranges contain the position, then this method will return multiple sets of monikers. Each slice
	// of monikers are attached to a single range. The order of the output slice is "outside-in", so that
	// the range attached to earlier monikers enclose the range attached to later monikers.
	MonikersByPosition(ctx context.Context, path string, line, character int) ([][]bundles.MonikerData, error)

	// MonikerResults returns the locations that define or reference the given moniker. This method
	// also returns the size of the complete result set to aid in pagination (along with skip and take).
	MonikerResults(ctx context.Context, tableName, scheme, identifier string, skip, take int) ([]bundles.Location, int, error)

	// PackageInformation looks up package information data by identifier.
	PackageInformation(ctx context.Context, path string, packageInformationID string) (bundles.PackageInformationData, bool, error)
}

type databaseImpl struct {
	filename        string
	reader          persistence.Reader // database file reader
	numResultChunks int                // numResultChunks value from meta row
}

var _ Database = &databaseImpl{}

func newRange(startLine, startCharacter, endLine, endCharacter int) bundles.Range {
	return bundles.Range{
		Start: bundles.Position{
			Line:      startLine,
			Character: startCharacter,
		},
		End: bundles.Position{
			Line:      endLine,
			Character: endCharacter,
		},
	}
}

// DocumentPathRangeID denotes a range qualified by its containing document.
type DocumentPathRangeID struct {
	Path    string
	RangeID types.ID
}

// ErrMalformedBundle is returned when a bundle is missing an expected map key.
type ErrMalformedBundle struct {
	Filename string // the filename of the malformed bundle
	Name     string // the type of value key should contain
	Key      string // the missing key
}

func (e ErrMalformedBundle) Error() string {
	return fmt.Sprintf("malformed bundle: unknown %s %s", e.Name, e.Key)
}

// OpenDatabase opens a handle to the bundle file at the given path.
func OpenDatabase(ctx context.Context, filename string, reader persistence.Reader) (Database, error) {
	meta, err := reader.ReadMeta(ctx)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "reader.ReadMeta")
	}

	return &databaseImpl{
		filename:        filename,
		reader:          reader,
		numResultChunks: meta.NumResultChunks,
	}, nil
}

// Close closes the underlying reader.
func (db *databaseImpl) Close() error {
	return db.reader.Close()
}

// Exists determines if the path exists in the database.
func (db *databaseImpl) Exists(ctx context.Context, path string) (bool, error) {
	_, exists, err := db.getDocumentData(ctx, path)
	return exists, pkgerrors.Wrap(err, "db.getDocumentData")
}

// Ranges returns definition, reference, and hover data for each range within the given span of lines.
func (db *databaseImpl) Ranges(ctx context.Context, path string, startLine, endLine int) ([]bundles.CodeIntelligenceRange, error) {
	documentData, exists, err := db.getDocumentData(ctx, path)
	if err != nil || !exists {
		return nil, pkgerrors.Wrap(err, "db.getDocumentData")
	}

	var codeintelRanges []bundles.CodeIntelligenceRange
	for _, r := range documentData.Ranges {
		if !rangeIntersectsSpan(r, startLine, endLine) {
			continue
		}

		definitions, _, err := db.definitions(ctx, r)
		if err != nil {
			return nil, err
		}

		references, _, err := db.references(ctx, r)
		if err != nil {
			return nil, err
		}

		hoverText, _, err := db.hover(ctx, documentData, r)
		if err != nil {
			return nil, err
		}

		codeintelRanges = append(codeintelRanges, bundles.CodeIntelligenceRange{
			Range:       newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter),
			Definitions: definitions,
			References:  references,
			HoverText:   hoverText,
		})
	}

	sort.Slice(codeintelRanges, func(i, j int) bool {
		cmp := codeintelRanges[i].Range.Start.Line - codeintelRanges[j].Range.Start.Line
		if cmp == 0 {
			cmp = codeintelRanges[i].Range.Start.Character - codeintelRanges[j].Range.Start.Character
		}

		return cmp < 0
	})

	return codeintelRanges, nil
}

// Definitions returns the set of locations defining the symbol at the given position.
func (db *databaseImpl) Definitions(ctx context.Context, path string, line, character int) ([]bundles.Location, error) {
	_, ranges, exists, err := db.getRangeByPosition(ctx, path, line, character)
	if err != nil || !exists {
		return nil, pkgerrors.Wrap(err, "db.getRangeByPosition")
	}

	for _, r := range ranges {
		locations, exists, err := db.definitions(ctx, r)
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}

		return locations, nil
	}

	return []bundles.Location{}, nil
}

// definitions returns the definition locations for the given range.
func (db *databaseImpl) definitions(ctx context.Context, r types.RangeData) ([]bundles.Location, bool, error) {
	if r.DefinitionResultID == "" {
		return nil, false, nil
	}

	definitionResults, err := db.getResultByID(ctx, r.DefinitionResultID)
	if err != nil {
		return nil, false, pkgerrors.Wrap(err, "db.getResultByID")
	}

	locations, err := db.convertRangesToLocations(ctx, definitionResults)
	if err != nil {
		return nil, false, pkgerrors.Wrap(err, "db.convertRangesToLocations")
	}

	return locations, true, nil
}

// References returns the set of locations referencing the symbol at the given position.
func (db *databaseImpl) References(ctx context.Context, path string, line, character int) ([]bundles.Location, error) {
	_, ranges, exists, err := db.getRangeByPosition(ctx, path, line, character)
	if err != nil || !exists {
		return nil, pkgerrors.Wrap(err, "db.getRangeByPosition")
	}

	var allLocations []bundles.Location
	for _, r := range ranges {
		locations, _, err := db.references(ctx, r)
		if err != nil {
			return nil, err
		}

		allLocations = append(allLocations, locations...)
	}

	return allLocations, nil
}

// references returns the reference locations for the given range.
func (db *databaseImpl) references(ctx context.Context, r types.RangeData) ([]bundles.Location, bool, error) {
	if r.ReferenceResultID == "" {
		return nil, false, nil
	}

	referenceResults, err := db.getResultByID(ctx, r.ReferenceResultID)
	if err != nil {
		return nil, false, pkgerrors.Wrap(err, "db.getResultByID")
	}

	locations, err := db.convertRangesToLocations(ctx, referenceResults)
	if err != nil {
		return nil, false, pkgerrors.Wrap(err, "db.convertRangesToLocations")
	}

	return locations, true, nil
}

// Hover returns the hover text of the symbol at the given position.
func (db *databaseImpl) Hover(ctx context.Context, path string, line, character int) (string, bundles.Range, bool, error) {
	documentData, ranges, exists, err := db.getRangeByPosition(ctx, path, line, character)
	if err != nil || !exists {
		return "", bundles.Range{}, false, pkgerrors.Wrap(err, "db.getRangeByPosition")
	}

	for _, r := range ranges {
		text, exists, err := db.hover(ctx, documentData, r)
		if err != nil {
			return "", bundles.Range{}, false, err
		}
		if !exists {
			continue
		}

		return text, newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter), true, nil
	}

	return "", bundles.Range{}, false, nil
}

// hover returns the hover text locations for the given range.
func (db *databaseImpl) hover(ctx context.Context, documentData types.DocumentData, r types.RangeData) (string, bool, error) {
	if r.HoverResultID == "" {
		return "", false, nil
	}

	text, exists := documentData.HoverResults[r.HoverResultID]
	if !exists {
		return "", false, ErrMalformedBundle{
			Filename: db.filename,
			Name:     "hoverResult",
			Key:      string(r.HoverResultID),
			// TODO(efritz) - add document context
		}
	}

	return text, true, nil
}

// Diagnostics returns the diagnostics for the documents that have the given path prefix. This method
// also returns the size of the complete result set to aid in pagination (along with skip and take).
func (db *databaseImpl) Diagnostics(ctx context.Context, prefix string, skip, take int) ([]bundles.Diagnostic, int, error) {
	paths, err := db.getPathsWithPrefix(ctx, prefix)
	if err != nil {
		return nil, 0, pkgerrors.Wrap(err, "db.getPathsWithPrefix")
	}

	// TODO(efritz) - we may need to store the diagnostic count outside of the
	// document so that we can efficiently skip over results that we've already
	// encountered.

	totalCount := 0
	var diagnostics []bundles.Diagnostic
	for _, path := range paths {
		documentData, exists, err := db.getDocumentData(ctx, path)
		if err != nil {
			return nil, 0, pkgerrors.Wrap(err, "db.getDocumentData")
		}
		if !exists {
			return nil, 0, nil
		}

		totalCount += len(documentData.Diagnostics)

		for _, diagnostic := range documentData.Diagnostics {
			skip--
			if skip < 0 && len(diagnostics) < take {
				diagnostics = append(diagnostics, bundles.Diagnostic{
					Path:           path,
					Severity:       diagnostic.Severity,
					Code:           diagnostic.Code,
					Message:        diagnostic.Message,
					Source:         diagnostic.Source,
					StartLine:      diagnostic.StartLine,
					StartCharacter: diagnostic.StartCharacter,
					EndLine:        diagnostic.EndLine,
					EndCharacter:   diagnostic.EndCharacter,
				})
			}
		}
	}

	return diagnostics, totalCount, nil
}

// MonikersByPosition returns all monikers attached ranges containing the given position. If multiple
// ranges contain the position, then this method will return multiple sets of monikers. Each slice
// of monikers are attached to a single range. The order of the output slice is "outside-in", so that
// the range attached to earlier monikers enclose the range attached to later monikers.
func (db *databaseImpl) MonikersByPosition(ctx context.Context, path string, line, character int) ([][]bundles.MonikerData, error) {
	documentData, ranges, exists, err := db.getRangeByPosition(ctx, path, line, character)
	if err != nil || !exists {
		return nil, pkgerrors.Wrap(err, "db.getRangeByPosition")
	}

	var monikerData [][]bundles.MonikerData
	for _, r := range ranges {
		var batch []bundles.MonikerData
		for _, monikerID := range r.MonikerIDs {
			moniker, exists := documentData.Monikers[monikerID]
			if !exists {
				return nil, ErrMalformedBundle{
					Filename: db.filename,
					Name:     "moniker",
					Key:      string(monikerID),
					// TODO(efritz) - add document context
				}
			}

			batch = append(batch, bundles.MonikerData{
				Kind:                 moniker.Kind,
				Scheme:               moniker.Scheme,
				Identifier:           moniker.Identifier,
				PackageInformationID: string(moniker.PackageInformationID),
			})
		}

		monikerData = append(monikerData, batch)
	}

	return monikerData, nil
}

// MonikerResults returns the locations that define or reference the given moniker. This method
// also returns the size of the complete result set to aid in pagination (along with skip and take).
func (db *databaseImpl) MonikerResults(ctx context.Context, tableName, scheme, identifier string, skip, take int) (_ []bundles.Location, _ int, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "getResultChunkByResultID")
	span.SetTag("filename", db.filename)
	span.SetTag("tableName", tableName)
	span.SetTag("scheme", scheme)
	span.SetTag("identifier", identifier)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	var rows []types.Location
	var totalCount int
	if tableName == "definitions" {
		if rows, totalCount, err = db.reader.ReadDefinitions(ctx, scheme, identifier, skip, take); err != nil {
			err = pkgerrors.Wrap(err, "reader.ReadDefinitions")
		}
	} else if tableName == "references" {
		if rows, totalCount, err = db.reader.ReadReferences(ctx, scheme, identifier, skip, take); err != nil {
			err = pkgerrors.Wrap(err, "reader.ReadReferences")
		}
	}

	if err != nil {
		return nil, 0, err
	}

	var locations []bundles.Location
	for _, row := range rows {
		locations = append(locations, bundles.Location{
			Path:  row.URI,
			Range: newRange(row.StartLine, row.StartCharacter, row.EndLine, row.EndCharacter),
		})
	}

	return locations, totalCount, nil
}

// PackageInformation looks up package information data by identifier.
func (db *databaseImpl) PackageInformation(ctx context.Context, path string, packageInformationID string) (bundles.PackageInformationData, bool, error) {
	documentData, exists, err := db.getDocumentData(ctx, path)
	if err != nil {
		return bundles.PackageInformationData{}, false, pkgerrors.Wrap(err, "db.getDocumentData")
	}
	if !exists {
		return bundles.PackageInformationData{}, false, nil
	}

	packageInformationData, exists := documentData.PackageInformation[types.ID(packageInformationID)]
	if exists {
		return bundles.PackageInformationData{
			Name:    packageInformationData.Name,
			Version: packageInformationData.Version,
		}, true, nil
	}

	return bundles.PackageInformationData{}, false, nil
}

func (db *databaseImpl) getPathsWithPrefix(ctx context.Context, prefix string) (_ []string, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "getPathsWithPrefix")
	span.SetTag("filename", db.filename)
	span.SetTag("prefix", prefix)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	return db.reader.PathsWithPrefix(ctx, prefix)
}

// getDocumentData fetches and unmarshals the document data or the given path. This method caches
// document data by a unique key prefixed by the database filename.
func (db *databaseImpl) getDocumentData(ctx context.Context, path string) (_ types.DocumentData, _ bool, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "getDocumentData")
	span.SetTag("filename", db.filename)
	span.SetTag("path", path)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	documentData, ok, err := db.reader.ReadDocument(ctx, path)
	if err != nil {
		return types.DocumentData{}, false, pkgerrors.Wrap(err, "reader.ReadDocument")
	}
	return documentData, ok, nil
}

// getRangeByPosition returns the ranges the given position. The order of the output slice is "outside-in",
// so that earlier ranges properly enclose later ranges.
func (db *databaseImpl) getRangeByPosition(ctx context.Context, path string, line, character int) (types.DocumentData, []types.RangeData, bool, error) {
	documentData, exists, err := db.getDocumentData(ctx, path)
	if err != nil {
		return types.DocumentData{}, nil, false, pkgerrors.Wrap(err, "db.getDocumentData")
	}
	if !exists {
		return types.DocumentData{}, nil, false, nil
	}

	return documentData, findRanges(documentData.Ranges, line, character), true, nil
}

// getResultByID fetches and unmarshals a definition or reference result by identifier.
// This method caches result chunk data by a unique key prefixed by the database filename.
func (db *databaseImpl) getResultByID(ctx context.Context, id types.ID) ([]DocumentPathRangeID, error) {
	resultChunkData, exists, err := db.getResultChunkByResultID(ctx, id)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "db.getResultChunkByResultID")
	}
	if !exists {
		return nil, ErrMalformedBundle{
			Filename: db.filename,
			Name:     "result chunk",
			Key:      string(id),
		}
	}

	documentIDRangeIDs, exists := resultChunkData.DocumentIDRangeIDs[id]
	if !exists {
		return nil, ErrMalformedBundle{
			Filename: db.filename,
			Name:     "result",
			Key:      string(id),
			// TODO(efritz) - add result chunk context
		}
	}

	var resultData []DocumentPathRangeID
	for _, documentIDRangeID := range documentIDRangeIDs {
		path, ok := resultChunkData.DocumentPaths[documentIDRangeID.DocumentID]
		if !ok {
			return nil, ErrMalformedBundle{
				Filename: db.filename,
				Name:     "documentPath",
				Key:      string(documentIDRangeID.DocumentID),
				// TODO(efritz) - add result chunk context
			}
		}

		resultData = append(resultData, DocumentPathRangeID{
			Path:    path,
			RangeID: documentIDRangeID.RangeID,
		})
	}

	return resultData, nil
}

// getResultChunkByResultID fetches and unmarshals the result chunk data with the given identifier.
// This method caches result chunk data by a unique key prefixed by the database filename.
func (db *databaseImpl) getResultChunkByResultID(ctx context.Context, id types.ID) (_ types.ResultChunkData, _ bool, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "getResultChunkByResultID")
	span.SetTag("filename", db.filename)
	span.SetTag("id", id)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	resultChunkData, ok, err := db.reader.ReadResultChunk(ctx, types.HashKey(id, db.numResultChunks))
	if err != nil {
		return types.ResultChunkData{}, false, pkgerrors.Wrap(err, "reader.ReadResultChunk")
	}
	return resultChunkData, ok, nil
}

// convertRangesToLocations converts pairs of document paths and range identifiers
// to a list of locations.
func (db *databaseImpl) convertRangesToLocations(ctx context.Context, resultData []DocumentPathRangeID) ([]bundles.Location, error) {
	// We potentially have to open a lot of documents. Reduce possible pressure on the
	// cache by ordering our queries so we only have to read and unmarshal each document
	// once.

	groupedResults := map[string][]types.ID{}
	for _, documentPathRangeID := range resultData {
		groupedResults[documentPathRangeID.Path] = append(groupedResults[documentPathRangeID.Path], documentPathRangeID.RangeID)
	}

	paths := []string{}
	for path := range groupedResults {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	var locations []bundles.Location
	for _, path := range paths {
		documentData, exists, err := db.getDocumentData(ctx, path)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "db.getDocumentData")
		}

		if !exists {
			return nil, ErrMalformedBundle{
				Filename: db.filename,
				Name:     "document",
				Key:      string(path),
			}
		}

		for _, rangeID := range groupedResults[path] {
			r, exists := documentData.Ranges[rangeID]
			if !exists {
				return nil, ErrMalformedBundle{
					Filename: db.filename,
					Name:     "range",
					Key:      string(rangeID),
					// TODO(efritz) - add document context
				}
			}

			locations = append(locations, bundles.Location{
				Path:  path,
				Range: newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter),
			})
		}
	}

	return locations, nil
}
