package lsifstore

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *store) GetCallerLocations(ctx context.Context, bundleID int, path string, line, character, limit, offset int) (_ []shared.Location, err error) {
	references, _, err := s.GetReferenceLocations(ctx, bundleID, path, line, character, limit, offset)
	// fmt.Printf("FOUND REFERENCES FOR LOCATION %+v\n", references)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get references for cursor")
	}

	callerLocations := make([]shared.Location, 0)

	for _, reference := range references {
		documentData, exists, err := s.scanFirstDocumentData(s.db.Query(ctx, sqlf.Sprintf(
			locationsDocumentQuery,
			reference.DumpID,
			reference.Path,
			reference.DumpID,
			reference.Path,
		)))
		if err != nil || !exists {
			return nil, errors.Wrapf(err, "error scanning document data for %+v", reference)
		}

		if documentData.SCIPData == nil {
			continue
		}

		occurrences := types.FindOccurrences(documentData.SCIPData.Occurrences, int32(reference.Range.Start.Line), int32(reference.Range.Start.Character))
		if len(occurrences) > 1 {
			panic("MORE THAN 1 OCCURRENCE FOUND AT REFERENCE LOCATION")
		}
		occurrence := occurrences[0]

		if occurrence.SymbolRoles == int32(scip.SymbolRole_Definition) {
			continue
		}

		// fmt.Printf("FOUND OCCURRENCE %+v\n", occurrence)

		for _, occ := range documentData.SCIPData.Occurrences {
			matches := strings.HasSuffix(occ.Symbol, occurrence.Owner)
			// fmt.Printf("CHECKING OCCURRENCE %s %s %t %t\n", occ.Symbol, occurrence.Owner, matches, occ.SymbolRoles == int32(scip.SymbolRole_Definition))

			if matches && occ.SymbolRoles == int32(scip.SymbolRole_Definition) {

				// fmt.Printf("FOUND MATCHING OCCURRENCE %+v\n", occ)

				var r types.Range
				if len(occ.Range) == 3 {
					r.Start.Line = int(occ.Range[0])
					r.End.Line = int(occ.Range[0])
					r.Start.Character = int(occ.Range[1])
					r.End.Character = int(occ.Range[2])
				} else {
					r.Start.Line = int(occ.Range[0])
					r.End.Line = int(occ.Range[2])
					r.Start.Character = int(occ.Range[1])
					r.End.Character = int(occ.Range[3])
				}

				callerLocations = append(callerLocations, shared.Location{
					DumpID: reference.DumpID,
					Path:   documentData.Path,
					Range:  r,
				})
			}
		}
	}

	return callerLocations, nil
}
