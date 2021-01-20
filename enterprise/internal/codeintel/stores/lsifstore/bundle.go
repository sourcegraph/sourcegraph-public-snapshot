package lsifstore

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"
	pkgerrors "github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// ErrNotFound occurs when data does not exist for a requested bundle.
var ErrNotFound = errors.New("data does not exist")

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

// Exists determines if the path exists in the database.
func (s *Store) Exists(ctx context.Context, bundleID int, path string) (_ bool, err error) {
	ctx, endObservation := s.operations.exists.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
	}})
	defer endObservation(1, observation.Args{})

	_, exists, err := s.getDocumentData(ctx, bundleID, path)
	return exists, pkgerrors.Wrap(err, "s.getDocumentData")
}

// Ranges returns definition, reference, and hover data for each range within the given span of lines.
func (s *Store) Ranges(ctx context.Context, bundleID int, path string, startLine, endLine int) (_ []CodeIntelligenceRange, err error) {
	ctx, endObservation := s.operations.ranges.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("startLine", startLine),
		log.Int("endLine", endLine),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.getDocumentData(ctx, bundleID, path)
	if err != nil || !exists {
		return nil, pkgerrors.Wrap(err, "s.getDocumentData")
	}

	var rangeIDs []ID
	for id, r := range documentData.Ranges {
		if RangeIntersectsSpan(r, startLine, endLine) {
			rangeIDs = append(rangeIDs, id)
		}
	}

	resultIDSet := map[ID]struct{}{}
	for _, rangeID := range rangeIDs {
		r := documentData.Ranges[rangeID]
		resultIDSet[r.DefinitionResultID] = struct{}{}
		resultIDSet[r.ReferenceResultID] = struct{}{}
	}

	// Remove empty results
	delete(resultIDSet, "")

	var resultIDs []ID
	for id := range resultIDSet {
		resultIDs = append(resultIDs, id)
	}

	locations, err := s.locations(ctx, bundleID, resultIDs)
	if err != nil {
		return nil, err
	}

	var codeintelRanges []CodeIntelligenceRange
	for _, rangeID := range rangeIDs {
		r := documentData.Ranges[rangeID]

		hoverText, _, err := s.hover(ctx, bundleID, path, documentData, r)
		if err != nil {
			return nil, err
		}

		// Return only references that are in the same file. Otherwise this set
		// gets very big and such results are of limited use to consumers such as
		// the code intel extensions, which only use references for highlighting
		// uses of an identifier within the same file.
		fileLocalReferences := make([]Location, 0, len(locations[r.ReferenceResultID]))
		for _, r := range locations[r.ReferenceResultID] {
			if r.Path == path {
				fileLocalReferences = append(fileLocalReferences, r)
			}
		}

		codeintelRanges = append(codeintelRanges, CodeIntelligenceRange{
			Range:       newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter),
			Definitions: locations[r.DefinitionResultID],
			References:  fileLocalReferences,
			HoverText:   hoverText,
		})
	}

	sort.Slice(codeintelRanges, func(i, j int) bool {
		return compareBundleRanges(codeintelRanges[i].Range, codeintelRanges[j].Range)
	})

	return codeintelRanges, nil
}

// Definitions returns the set of locations defining the symbol at the given position.
func (s *Store) Definitions(ctx context.Context, bundleID int, path string, line, character int) (_ []Location, err error) {
	ctx, endObservation := s.operations.definitions.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("line", line),
		log.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	_, ranges, exists, err := s.getRangeByPosition(ctx, bundleID, path, line, character)
	if err != nil || !exists {
		return nil, pkgerrors.Wrap(err, "s.getRangeByPosition")
	}

	for _, r := range ranges {
		if r.DefinitionResultID == "" {
			continue
		}

		locations, err := s.locations(ctx, bundleID, []ID{r.DefinitionResultID})
		if err != nil {
			return nil, err
		}

		return locations[r.DefinitionResultID], nil
	}

	return []Location{}, nil
}

// References returns the set of locations referencing the symbol at the given position.
func (s *Store) References(ctx context.Context, bundleID int, path string, line, character int) (_ []Location, err error) {
	ctx, endObservation := s.operations.references.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("line", line),
		log.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	_, ranges, exists, err := s.getRangeByPosition(ctx, bundleID, path, line, character)
	if err != nil || !exists {
		return nil, pkgerrors.Wrap(err, "s.getRangeByPosition")
	}

	var allLocations []Location
	for _, r := range ranges {
		if r.ReferenceResultID == "" {
			continue
		}

		locations, err := s.locations(ctx, bundleID, []ID{r.ReferenceResultID})
		if err != nil {
			return nil, err
		}

		allLocations = append(allLocations, locations[r.ReferenceResultID]...)
	}

	return allLocations, nil
}

// Hover returns the hover text of the symbol at the given position.
func (s *Store) Hover(ctx context.Context, bundleID int, path string, line, character int) (_ string, _ Range, _ bool, err error) {
	ctx, endObservation := s.operations.hover.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("line", line),
		log.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	documentData, ranges, exists, err := s.getRangeByPosition(ctx, bundleID, path, line, character)
	if err != nil || !exists {
		return "", Range{}, false, pkgerrors.Wrap(err, "s.getRangeByPosition")
	}

	for _, r := range ranges {
		text, exists, err := s.hover(ctx, bundleID, path, documentData, r)
		if err != nil {
			return "", Range{}, false, err
		}
		if !exists {
			continue
		}

		return text, newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter), true, nil
	}

	return "", Range{}, false, nil
}

// Diagnostics returns the diagnostics for the documents that have the given path prefix. This method
// also returns the size of the complete result set to aid in pagination (along with skip and take).
func (s *Store) Diagnostics(ctx context.Context, bundleID int, prefix string, skip, take int) (_ []Diagnostic, _ int, err error) {
	ctx, endObservation := s.operations.diagnostics.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("prefix", prefix),
		log.Int("skip", skip),
		log.Int("take", take),
	}})
	defer endObservation(1, observation.Args{})

	paths, err := s.getPathsWithPrefix(ctx, bundleID, prefix)
	if err != nil {
		return nil, 0, pkgerrors.Wrap(err, "s.getPathsWithPrefix")
	}

	// TODO(efritz) - we may need to store the diagnostic count outside of the
	// document so that we can efficiently skip over results that we've already
	// encountered.

	totalCount := 0
	var diagnostics []Diagnostic
	for _, path := range paths {
		documentData, exists, err := s.getDocumentData(ctx, bundleID, path)
		if err != nil {
			return nil, 0, pkgerrors.Wrap(err, "s.getDocumentData")
		}
		if !exists {
			return nil, 0, nil
		}

		totalCount += len(documentData.Diagnostics)

		for _, diagnostic := range documentData.Diagnostics {
			skip--
			if skip < 0 && len(diagnostics) < take {
				diagnostics = append(diagnostics, Diagnostic{
					DumpID: bundleID,
					Path:   path,
					DiagnosticData: DiagnosticData{
						Severity:       diagnostic.Severity,
						Code:           diagnostic.Code,
						Message:        diagnostic.Message,
						Source:         diagnostic.Source,
						StartLine:      diagnostic.StartLine,
						StartCharacter: diagnostic.StartCharacter,
						EndLine:        diagnostic.EndLine,
						EndCharacter:   diagnostic.EndCharacter,
					},
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
func (s *Store) MonikersByPosition(ctx context.Context, bundleID int, path string, line, character int) (_ [][]MonikerData, err error) {
	ctx, endObservation := s.operations.monikersByPosition.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("line", line),
		log.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	documentData, ranges, exists, err := s.getRangeByPosition(ctx, bundleID, path, line, character)
	if err != nil || !exists {
		return nil, pkgerrors.Wrap(err, "s.getRangeByPosition")
	}

	var monikerData [][]MonikerData
	for _, r := range ranges {
		var batch []MonikerData
		for _, monikerID := range r.MonikerIDs {
			moniker, exists := documentData.Monikers[monikerID]
			if !exists {
				log15.Warn("malformed bundle: unknown moniker", "bundleID", bundleID, "path", path, "id", monikerID)
				continue
			}

			batch = append(batch, moniker)
		}

		monikerData = append(monikerData, batch)
	}

	return monikerData, nil
}

// MonikerResults returns the locations that define or reference the given moniker. This method
// also returns the size of the complete result set to aid in pagination (along with skip and take).
func (s *Store) MonikerResults(ctx context.Context, bundleID int, tableName, scheme, identifier string, skip, take int) (_ []Location, _ int, err error) {
	ctx, endObservation := s.operations.monikerResults.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("tableName", tableName),
		log.String("scheme", scheme),
		log.String("identifier", identifier),
		log.Int("skip", skip),
		log.Int("take", take),
	}})
	defer endObservation(1, observation.Args{})

	var rows []LocationData
	var totalCount int
	if tableName == "definitions" {
		if rows, totalCount, err = s.ReadDefinitions(ctx, bundleID, scheme, identifier, skip, take); err != nil {
			err = pkgerrors.Wrap(err, "store.ReadDefinitions")
		}
	} else if tableName == "references" {
		if rows, totalCount, err = s.ReadReferences(ctx, bundleID, scheme, identifier, skip, take); err != nil {
			err = pkgerrors.Wrap(err, "store.ReadReferences")
		}
	}

	if err != nil {
		return nil, 0, err
	}

	var locations []Location
	for _, row := range rows {
		locations = append(locations, Location{
			DumpID: bundleID,
			Path:   row.URI,
			Range:  newRange(row.StartLine, row.StartCharacter, row.EndLine, row.EndCharacter),
		})
	}

	return locations, totalCount, nil
}

// PackageInformation looks up package information data by identifier.
func (s *Store) PackageInformation(ctx context.Context, bundleID int, path, packageInformationID string) (_ PackageInformationData, _ bool, err error) {
	ctx, endObservation := s.operations.packageInformation.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.String("packageInformationID", packageInformationID),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.getDocumentData(ctx, bundleID, path)
	if err != nil {
		return PackageInformationData{}, false, pkgerrors.Wrap(err, "s.getDocumentData")
	}
	if !exists {
		return PackageInformationData{}, false, nil
	}

	packageInformationData, exists := documentData.PackageInformation[ID(packageInformationID)]
	if exists {
		return PackageInformationData{
			Name:    packageInformationData.Name,
			Version: packageInformationData.Version,
		}, true, nil
	}

	return PackageInformationData{}, false, nil
}

// hover returns the hover text locations for the given range.
func (s *Store) hover(ctx context.Context, bundleID int, path string, documentData DocumentData, r RangeData) (string, bool, error) {
	if r.HoverResultID == "" {
		return "", false, nil
	}

	text, exists := documentData.HoverResults[r.HoverResultID]
	if !exists {
		log15.Warn("malformed bundle: unknown hover result", "bundleID", bundleID, "path", path, "id", r.HoverResultID)
		return "", false, nil
	}

	return text, true, nil
}

func (s *Store) getPathsWithPrefix(ctx context.Context, bundleID int, prefix string) (_ []string, err error) {
	return s.PathsWithPrefix(ctx, bundleID, prefix)
}

// getDocumentData fetches and unmarshals the document data or the given path.
func (s *Store) getDocumentData(ctx context.Context, bundleID int, path string) (_ DocumentData, _ bool, err error) {
	documentData, ok, err := s.ReadDocument(ctx, bundleID, path)
	if err != nil {
		return DocumentData{}, false, pkgerrors.Wrap(err, "store.ReadDocument")
	}
	return documentData, ok, nil
}

// getRangeByPosition returns the ranges the given position. The order of the output slice is "outside-in",
// so that earlier ranges properly enclose later ranges.
func (s *Store) getRangeByPosition(ctx context.Context, bundleID int, path string, line, character int) (DocumentData, []RangeData, bool, error) {
	documentData, exists, err := s.getDocumentData(ctx, bundleID, path)
	if err != nil {
		return DocumentData{}, nil, false, pkgerrors.Wrap(err, "s.getDocumentData")
	}
	if !exists {
		return DocumentData{}, nil, false, nil
	}

	return documentData, FindRanges(documentData.Ranges, line, character), true, nil
}

// locations returns the locations for the given definition or reference identifiers.
func (s *Store) locations(ctx context.Context, bundleID int, ids []ID) (map[ID][]Location, error) {
	results, err := s.getResultsByIDs(ctx, bundleID, ids)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "s.getResultByID")
	}

	locationWrapper, err := s.convertRangesToLocations(ctx, bundleID, results)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "s.convertRangesToLocations")
	}

	return locationWrapper, nil
}

// getResultsByIDs fetches and unmarshals a definition or reference results for the given identifiers.
func (s *Store) getResultsByIDs(ctx context.Context, bundleID int, ids []ID) (map[ID][]DocumentPathRangeID, error) {
	xids := map[int]struct{}{}
	for _, id := range ids {
		index, err := s.resultChunkID(ctx, bundleID, id)
		if err != nil {
			return nil, err
		}

		xids[index] = struct{}{}
	}

	resultChunks := map[int]ResultChunkData{}
	for index := range xids {
		resultChunkData, exists, err := s.getResultChunkByID(ctx, bundleID, index)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "s.getResultChunkByID")
		}
		if !exists {
			log15.Warn("malformed bundle: unknown result chunk", "bundleID", bundleID, "index", index)
			continue
		}

		resultChunks[index] = resultChunkData
	}

	data := map[ID][]DocumentPathRangeID{}

	for _, id := range ids {
		index, err := s.resultChunkID(ctx, bundleID, id)
		if err != nil {
			return nil, err
		}

		resultChunkData := resultChunks[index]

		documentIDRangeIDs, exists := resultChunkData.DocumentIDRangeIDs[id]
		if !exists {
			log15.Warn("malformed bundle: unknown result", "bundleID", bundleID, "index", index, "id", id)
			continue
		}

		var resultData []DocumentPathRangeID
		for _, documentIDRangeID := range documentIDRangeIDs {
			path, ok := resultChunkData.DocumentPaths[documentIDRangeID.DocumentID]
			if !ok {
				log15.Warn("malformed bundle: unknown document path", "bundleID", bundleID, "index", index, "id", documentIDRangeID.DocumentID)
				continue
			}

			resultData = append(resultData, DocumentPathRangeID{
				Path:    path,
				RangeID: documentIDRangeID.RangeID,
			})
		}

		data[id] = resultData
	}

	return data, nil
}

// getResultChunkByID fetches and unmarshals the result chunk data with the given identifier.
func (s *Store) getResultChunkByID(ctx context.Context, bundleID int, id int) (_ ResultChunkData, _ bool, err error) {
	resultChunkData, ok, err := s.ReadResultChunk(ctx, bundleID, id)
	if err != nil {
		return ResultChunkData{}, false, pkgerrors.Wrap(err, "store.ReadResultChunk")
	}
	return resultChunkData, ok, nil
}

// resultChunkID returns the identifier of the result chunk that contains the given identifier.
func (s *Store) resultChunkID(ctx context.Context, bundleID int, id ID) (int, error) {
	// TODO(efritz) - keep a cache
	meta, err := s.ReadMeta(ctx, bundleID)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "store.ReadMeta")
	}

	return HashKey(id, meta.NumResultChunks), nil
}

// convertRangesToLocations converts pairs of document paths and range identifiers to a list of locations.
func (s *Store) convertRangesToLocations(ctx context.Context, bundleID int, pairs map[ID][]DocumentPathRangeID) (map[ID][]Location, error) {
	// We potentially have to open a lot of documents. Reduce possible pressure on the
	// cache by ordering our queries so we only have to read and unmarshal each document
	// once.

	groupedResults := map[string][]ID{}
	for _, resultData := range pairs {
		for _, documentPathRangeID := range resultData {
			groupedResults[documentPathRangeID.Path] = append(groupedResults[documentPathRangeID.Path], documentPathRangeID.RangeID)
		}
	}

	documents := map[string]DocumentData{}
	for path := range groupedResults {
		documentData, exists, err := s.getDocumentData(ctx, bundleID, path)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "s.getDocumentData")
		}

		if !exists {
			log15.Warn("malformed bundle: unknown document", "bundleID", bundleID, "path", path)
			continue
		}

		documents[path] = documentData
	}

	locationsByID := map[ID][]Location{}
	for id, resultData := range pairs {
		var locations []Location
		for _, documentPathRangeID := range resultData {
			path := documentPathRangeID.Path
			rangeID := documentPathRangeID.RangeID

			r, exists := documents[path].Ranges[rangeID]
			if !exists {
				log15.Warn("malformed bundle: unknown range", "bundleID", bundleID, "path", path, "id", id)
				continue
			}

			locations = append(locations, Location{
				DumpID: bundleID,
				Path:   path,
				Range:  newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter),
			})
		}

		sort.Slice(locations, func(i, j int) bool {
			if locations[i].Path == locations[j].Path {
				return compareBundleRanges(locations[i].Range, locations[j].Range)
			}

			return strings.Compare(locations[i].Path, locations[j].Path) < 0
		})

		locationsByID[id] = locations
	}

	return locationsByID, nil
}

// compareBundleRanges returns true if r1's start position occurs before r2's start position.
func compareBundleRanges(r1, r2 Range) bool {
	cmp := r1.Start.Line - r2.Start.Line
	if cmp == 0 {
		cmp = r1.Start.Character - r2.Start.Character
	}

	return cmp < 0
}
