package lsifstore

import (
	"context"
	"sort"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Ranges returns definition, reference, and hover data for each range within the given span of lines.
func (s *Store) Ranges(ctx context.Context, bundleID int, path string, startLine, endLine int) (_ []CodeIntelligenceRange, err error) {
	ctx, traceLog, endObservation := s.operations.ranges.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("startLine", startLine),
		log.Int("endLine", endLine),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(rangesDocumentQuery, bundleID, path)))
	if err != nil || !exists {
		return nil, err
	}
	traceLog(log.Int("numRanges", len(documentData.Document.Ranges)))

	ranges := map[ID]RangeData{}
	for id, r := range documentData.Document.Ranges {
		if RangeIntersectsSpan(r, startLine, endLine) {
			ranges[id] = r
		}
	}
	traceLog(log.Int("numIntersectingRanges", len(ranges)))

	resultIDMap := make(map[ID]struct{}, 2*len(ranges))
	for _, r := range ranges {
		if r.DefinitionResultID != "" {
			resultIDMap[r.DefinitionResultID] = struct{}{}
		}
		if r.ReferenceResultID != "" {
			resultIDMap[r.ReferenceResultID] = struct{}{}
		}
	}

	resultIDs := make([]ID, 0, len(resultIDMap))
	for id := range resultIDMap {
		resultIDs = append(resultIDs, id)
	}

	locations, err := s.locations(ctx, bundleID, resultIDs)
	if err != nil {
		return nil, err
	}

	codeintelRanges := make([]CodeIntelligenceRange, 0, len(ranges))
	for _, r := range ranges {
		var hoverText string
		if r.HoverResultID != "" {
			if text, exists := documentData.Document.HoverResults[r.HoverResultID]; exists {
				hoverText = text
			}
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

const rangesDocumentQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/ranges.go:Ranges
SELECT dump_id, path, data FROM lsif_data_documents WHERE dump_id = %s AND path = %s LIMIT 1
`
