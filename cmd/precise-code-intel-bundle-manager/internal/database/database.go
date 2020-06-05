package database

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/opentracing/opentracing-go/ext"
	pkgerrors "github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// Database wraps access to a single processed bundle.
type Database interface {
	// Close closes the underlying reader.
	Close() error

	// Exists determines if the path exists in the database.
	Exists(ctx context.Context, path string) (bool, error)

	// Definitions returns the set of locations defining the symbol at the given position.
	Definitions(ctx context.Context, path string, line, character int) ([]Location, error)

	// References returns the set of locations referencing the symbol at the given position.
	References(ctx context.Context, path string, line, character int) ([]Location, error)

	// Hover returns the hover text of the symbol at the given position.
	Hover(ctx context.Context, path string, line, character int) (string, Range, bool, error)

	// MonikersByPosition returns all monikers attached ranges containing the given position. If multiple
	// ranges contain the position, then this method will return multiple sets of monikers. Each slice
	// of monikers are attached to a single range. The order of the output slice is "outside-in", so that
	// the range attached to earlier monikers enclose the range attached to later monikers.
	MonikersByPosition(ctx context.Context, path string, line, character int) ([][]types.MonikerData, error)

	// MonikerResults returns the locations that define or reference the given moniker. This method
	// also returns the size of the complete result set to aid in pagination (along with skip and take).
	MonikerResults(ctx context.Context, tableName, scheme, identifier string, skip, take int) ([]Location, int, error)

	// PackageInformation looks up package information data by identifier.
	PackageInformation(ctx context.Context, path string, packageInformationID types.ID) (types.PackageInformationData, bool, error)
}

type databaseImpl struct {
	filename         string
	reader           persistence.Reader // database file reader
	documentCache    *DocumentCache     // shared cache
	resultChunkCache *ResultChunkCache  // shared cache
	numResultChunks  int                // numResultChunks value from meta row
}

var _ Database = &databaseImpl{}

type Location struct {
	Path  string `json:"path"`
	Range Range  `json:"range"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

func newRange(startLine, startCharacter, endLine, endCharacter int) Range {
	return Range{
		Start: Position{
			Line:      startLine,
			Character: startCharacter,
		},
		End: Position{
			Line:      endLine,
			Character: endCharacter,
		},
	}
}

// documentPathRangeID denotes a range qualified by its containing document.
type documentPathRangeID struct {
	Path    string
	RangeID types.ID
}

var ErrUnknownDocument = errors.New("unknown document")
var ErrUnknownResultChunk = errors.New("unknown result chunk")

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
func OpenDatabase(ctx context.Context, filename string, reader persistence.Reader, documentCache *DocumentCache, resultChunkCache *ResultChunkCache) (Database, error) {
	meta, err := reader.ReadMeta(ctx)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "reader.ReadMeta")
	}

	return &databaseImpl{
		filename:         filename,
		documentCache:    documentCache,
		resultChunkCache: resultChunkCache,
		reader:           reader,
		numResultChunks:  meta.NumResultChunks,
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

// Definitions returns the set of locations defining the symbol at the given position.
func (db *databaseImpl) Definitions(ctx context.Context, path string, line, character int) ([]Location, error) {
	_, ranges, exists, err := db.getRangeByPosition(ctx, path, line, character)
	if err != nil || !exists {
		return nil, pkgerrors.Wrap(err, "db.getRangeByPosition")
	}

	for _, r := range ranges {
		if r.DefinitionResultID == "" {
			continue
		}

		definitionResults, err := db.getResultByID(ctx, r.DefinitionResultID)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "db.getResultByID")
		}

		locations, err := db.convertRangesToLocations(ctx, definitionResults)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "db.convertRangesToLocations")
		}

		return locations, nil
	}

	return []Location{}, nil
}

// References returns the set of locations referencing the symbol at the given position.
func (db *databaseImpl) References(ctx context.Context, path string, line, character int) ([]Location, error) {
	_, ranges, exists, err := db.getRangeByPosition(ctx, path, line, character)
	if err != nil || !exists {
		return nil, pkgerrors.Wrap(err, "db.getRangeByPosition")
	}

	var allLocations []Location
	for _, r := range ranges {
		if r.ReferenceResultID == "" {
			continue
		}

		referenceResults, err := db.getResultByID(ctx, r.ReferenceResultID)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "db.getResultByID")
		}

		locations, err := db.convertRangesToLocations(ctx, referenceResults)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "db.convertRangesToLocations")
		}

		allLocations = append(allLocations, locations...)
	}

	return allLocations, nil
}

// Hover returns the hover text of the symbol at the given position.
func (db *databaseImpl) Hover(ctx context.Context, path string, line, character int) (string, Range, bool, error) {
	documentData, ranges, exists, err := db.getRangeByPosition(ctx, path, line, character)
	if err != nil || !exists {
		return "", Range{}, false, pkgerrors.Wrap(err, "db.getRangeByPosition")
	}

	for _, r := range ranges {
		if r.HoverResultID == "" {
			continue
		}

		text, exists := documentData.HoverResults[r.HoverResultID]
		if !exists {
			return "", Range{}, false, ErrMalformedBundle{
				Filename: db.filename,
				Name:     "hoverResult",
				Key:      string(r.HoverResultID),
				// TODO(efritz) - add document context
			}
		}

		return text, newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter), true, nil
	}

	return "", Range{}, false, nil
}

// MonikersByPosition returns all monikers attached ranges containing the given position. If multiple
// ranges contain the position, then this method will return multiple sets of monikers. Each slice
// of monikers are attached to a single range. The order of the output slice is "outside-in", so that
// the range attached to earlier monikers enclose the range attached to later monikers.
func (db *databaseImpl) MonikersByPosition(ctx context.Context, path string, line, character int) ([][]types.MonikerData, error) {
	documentData, ranges, exists, err := db.getRangeByPosition(ctx, path, line, character)
	if err != nil || !exists {
		return nil, pkgerrors.Wrap(err, "db.getRangeByPosition")
	}

	var monikerData [][]types.MonikerData
	for _, r := range ranges {
		var batch []types.MonikerData
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

			batch = append(batch, moniker)
		}

		monikerData = append(monikerData, batch)
	}

	return monikerData, nil
}

// MonikerResults returns the locations that define or reference the given moniker. This method
// also returns the size of the complete result set to aid in pagination (along with skip and take).
func (db *databaseImpl) MonikerResults(ctx context.Context, tableName, scheme, identifier string, skip, take int) ([]Location, int, error) {
	// TODO(efritz) - gross
	var rows []types.Location
	var totalCount int
	var err error
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

	var locations []Location
	for _, row := range rows {
		locations = append(locations, Location{
			Path:  row.URI,
			Range: newRange(row.StartLine, row.StartCharacter, row.EndLine, row.EndCharacter),
		})
	}

	return locations, totalCount, nil
}

// PackageInformation looks up package information data by identifier.
func (db *databaseImpl) PackageInformation(ctx context.Context, path string, packageInformationID types.ID) (types.PackageInformationData, bool, error) {
	documentData, exists, err := db.getDocumentData(ctx, path)
	if err != nil {
		return types.PackageInformationData{}, false, pkgerrors.Wrap(err, "db.getDocumentData")
	}

	if !exists {
		return types.PackageInformationData{}, false, nil
	}

	packageInformationData, exists := documentData.PackageInformation[packageInformationID]
	return packageInformationData, exists, nil
}

// getDocumentData fetches and unmarshals the document data or the given path. This method caches
// document data by a unique key prefixed by the database filename.
func (db *databaseImpl) getDocumentData(ctx context.Context, path string) (_ types.DocumentData, _ bool, err error) {
	cached := true
	span, ctx := ot.StartSpanFromContext(ctx, "getDocumentData")
	span.SetTag("filename", db.filename)
	span.SetTag("path", path)
	defer func() {
		span.SetTag("cached", cached)
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	documentData, err := db.documentCache.GetOrCreate(fmt.Sprintf("%s::%s", db.filename, path), func() (types.DocumentData, error) {
		cached = false

		data, ok, err := db.reader.ReadDocument(ctx, path)
		if err != nil {
			return types.DocumentData{}, pkgerrors.Wrap(err, "reader.ReadDocument")
		}
		if !ok {
			return types.DocumentData{}, ErrUnknownDocument
		}
		return data, nil
	})

	if err != nil {
		// TODO(efritz) - should change cache interface instead
		if err == ErrUnknownDocument {
			return types.DocumentData{}, false, nil
		}
		return types.DocumentData{}, false, err
	}

	return documentData, true, nil
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
func (db *databaseImpl) getResultByID(ctx context.Context, id types.ID) ([]documentPathRangeID, error) {
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

	var resultData []documentPathRangeID
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

		resultData = append(resultData, documentPathRangeID{
			Path:    path,
			RangeID: documentIDRangeID.RangeID,
		})
	}

	return resultData, nil
}

// getResultChunkByResultID fetches and unmarshals the result chunk data with the given identifier.
// This method caches result chunk data by a unique key prefixed by the database filename.
func (db *databaseImpl) getResultChunkByResultID(ctx context.Context, id types.ID) (_ types.ResultChunkData, _ bool, err error) {
	cached := true
	span, ctx := ot.StartSpanFromContext(ctx, "getResultChunkByResultID")
	span.SetTag("filename", db.filename)
	span.SetTag("id", id)
	defer func() {
		span.SetTag("cached", cached)
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	resultChunkData, err := db.resultChunkCache.GetOrCreate(fmt.Sprintf("%s::%s", db.filename, id), func() (types.ResultChunkData, error) {
		cached = false

		data, ok, err := db.reader.ReadResultChunk(ctx, types.HashKey(id, db.numResultChunks))
		if err != nil {
			return types.ResultChunkData{}, pkgerrors.Wrap(err, "reader.ReadResultChunk")
		}
		if !ok {
			return types.ResultChunkData{}, ErrUnknownResultChunk
		}
		return data, nil
	})

	if err != nil {
		// TODO(efritz) - should change cache interface instead
		if err == ErrUnknownResultChunk {
			return types.ResultChunkData{}, false, nil
		}
		return types.ResultChunkData{}, false, err
	}

	return resultChunkData, true, nil
}

// convertRangesToLocations converts pairs of document paths and range identifiers
// to a list of locations.
func (db *databaseImpl) convertRangesToLocations(ctx context.Context, resultData []documentPathRangeID) ([]Location, error) {
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

	var locations []Location
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

			locations = append(locations, Location{
				Path:  path,
				Range: newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter),
			})
		}
	}

	return locations, nil
}
