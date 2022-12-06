package lsifstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
)

func TestGetRanges(t *testing.T) {
	store := populateTestStore(t)
	path := "template/src/util/helpers.ts"

	// (comments above)
	// `export function nonEmpty<T>(value: T | T[] | null | undefined): value is T | T[] {`
	//                  ^^^^^^^^ ^  ^^^^^  ^   ^                        ^^^^^    ^   ^

	ranges, err := store.GetRanges(context.Background(), testSCIPUploadID, path, 13, 16)
	if err != nil {
		t.Fatalf("unexpected error querying ranges: %s", err)
	}
	for i := range ranges {
		// NOTE: currently in-flight as how we're doing this for now,
		// so we're just un-setting it for the assertions below.
		ranges[i].Implementations = nil
	}

	const (
		nonEmptyHoverText = "```ts\nfunction nonEmpty<T>(value: T | T[] | null | undefined): value is T | T[]\n```\nReturns true if the value is defined and, if an array, contains at least\none element."
		valueHoverText    = "```ts\n(parameter) value: T | T[] | null | undefined\n```\nThe value to test."
		tHoverText        = "```ts\nT: T\n```"
	)

	var (
		nonEmptyDefinitionLocations = []shared.Location{{DumpID: testSCIPUploadID, Path: path, Range: newRange(15, 16, 15, 24)}}
		tDefinitionLocations        = []shared.Location{{DumpID: testSCIPUploadID, Path: path, Range: newRange(15, 25, 15, 26)}}
		valueDefinitionLocations    = []shared.Location{{DumpID: testSCIPUploadID, Path: path, Range: newRange(15, 28, 15, 33)}}

		nonEmptyReferenceLocations = []shared.Location{}
		tReferenceLocations        = []shared.Location{
			{DumpID: testSCIPUploadID, Path: path, Range: newRange(15, 35, 15, 36)},
			{DumpID: testSCIPUploadID, Path: path, Range: newRange(15, 39, 15, 40)},
			{DumpID: testSCIPUploadID, Path: path, Range: newRange(15, 73, 15, 74)},
			{DumpID: testSCIPUploadID, Path: path, Range: newRange(15, 77, 15, 78)},
		}
		valueReferenceLocations = []shared.Location{
			{DumpID: testSCIPUploadID, Path: path, Range: newRange(15, 64, 15, 69)},
			{DumpID: testSCIPUploadID, Path: path, Range: newRange(16, 13, 16, 18)},
			{DumpID: testSCIPUploadID, Path: path, Range: newRange(16, 38, 16, 43)},
			{DumpID: testSCIPUploadID, Path: path, Range: newRange(16, 48, 16, 53)},
		}

		nonEmptyImplementationLocations = []shared.Location(nil)
		tImplementationLocations        = []shared.Location(nil)
		valueImplementationLocations    = []shared.Location(nil)
	)

	expectedRanges := []shared.CodeIntelligenceRange{
		{
			// `nonEmpty`
			Range:           newRange(15, 16, 15, 24),
			Definitions:     nonEmptyDefinitionLocations,
			References:      nonEmptyReferenceLocations,
			Implementations: nonEmptyImplementationLocations,
			HoverText:       nonEmptyHoverText,
		},
		{
			// `T`
			Range:           newRange(15, 25, 15, 26),
			Definitions:     tDefinitionLocations,
			References:      tReferenceLocations,
			Implementations: tImplementationLocations,
			HoverText:       tHoverText,
		},
		{
			// `value`
			Range:           newRange(15, 28, 15, 33),
			Definitions:     valueDefinitionLocations,
			References:      valueReferenceLocations,
			Implementations: valueImplementationLocations,
			HoverText:       valueHoverText,
		},
		{
			// `T`
			Range:           newRange(15, 35, 15, 36),
			Definitions:     tDefinitionLocations,
			References:      tReferenceLocations,
			Implementations: tImplementationLocations,
			HoverText:       tHoverText,
		},
		{
			// `T`
			Range:           newRange(15, 39, 15, 40),
			Definitions:     tDefinitionLocations,
			References:      tReferenceLocations,
			Implementations: tImplementationLocations,
			HoverText:       tHoverText,
		},
		{
			// `value`
			Range:           newRange(15, 64, 15, 69),
			Definitions:     valueDefinitionLocations,
			References:      valueReferenceLocations,
			Implementations: valueImplementationLocations,
			HoverText:       valueHoverText,
		},
		{
			// `T`
			Range:           newRange(15, 73, 15, 74),
			Definitions:     tDefinitionLocations,
			References:      tReferenceLocations,
			Implementations: tImplementationLocations,
			HoverText:       tHoverText,
		},
		{
			// `T`
			Range:           newRange(15, 77, 15, 78),
			Definitions:     tDefinitionLocations,
			References:      tReferenceLocations,
			Implementations: tImplementationLocations,
			HoverText:       tHoverText,
		},
	}
	if diff := cmp.Diff(expectedRanges, ranges); diff != "" {
		t.Errorf("unexpected ranges (-want +got):\n%s", diff)
	}
}
