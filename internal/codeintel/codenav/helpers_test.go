package codenav

import (
	"context"
	"testing"

	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore"
	lsifstoremocks "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore/mocks"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var uploadIDSupply = 0

func newUploadID() int {
	uploadIDSupply += 1
	return uploadIDSupply
}

// Generates a fake scip.Range that is easy to tell from other ranges
func testRange(r int) scip.Range {
	return scip.NewRangeUnchecked([]int32{int32(r), int32(r), int32(r)})
}

type fakeOccurrence struct {
	symbol       string
	isDefinition bool
	range_       scip.Range
}

type fakeDocument struct {
	path        core.UploadRelPath
	occurrences []fakeOccurrence
}

func (d fakeDocument) Occurrences() []*scip.Occurrence {
	occs := make([]*scip.Occurrence, 0, len(d.occurrences))
	for _, occ := range d.occurrences {
		var symbolRoles scip.SymbolRole = 0
		if occ.isDefinition {
			symbolRoles = scip.SymbolRole_Definition
		}
		occs = append(occs, &scip.Occurrence{
			Range:       occ.range_.SCIPRange(),
			Symbol:      occ.symbol,
			SymbolRoles: int32(symbolRoles),
		})
	}
	return occs
}

func sym(name string) string {
	return "test . . . " + name + "."
}

func ref(symbol string, range_ scip.Range) fakeOccurrence {
	return fakeOccurrence{
		symbol:       sym(symbol),
		isDefinition: false,
		range_:       range_,
	}
}

func def(symbol string, range_ scip.Range) fakeOccurrence {
	return fakeOccurrence{
		symbol:       sym(symbol),
		isDefinition: true,
		range_:       range_,
	}
}

func local(symbol string, range_ scip.Range) fakeOccurrence {
	return fakeOccurrence{
		symbol:       "local " + symbol,
		isDefinition: false,
		range_:       range_,
	}
}

func doc(path string, occurrences ...fakeOccurrence) fakeDocument {
	return fakeDocument{
		path:        core.NewUploadRelPathUnchecked(path),
		occurrences: occurrences,
	}
}

// Set up uploads + lsifstore
func setupUpload(commit api.CommitID, root string, documents ...fakeDocument) (uploadsshared.CompletedUpload, lsifstore.LsifStore) {
	id := newUploadID()
	lsifStore := lsifstoremocks.NewMockLsifStore()
	lsifStore.SCIPDocumentFunc.SetDefaultHook(func(ctx context.Context, uploadId int, path core.UploadRelPath) (core.Option[*scip.Document], error) {
		if id != uploadId {
			return core.None[*scip.Document](), errors.New("unknown upload id")
		}
		for _, document := range documents {
			if document.path.Equal(path) {
				return core.Some(&scip.Document{
					RelativePath: document.path.RawValue(),
					Occurrences:  document.Occurrences(),
				}), nil
			}
		}
		return core.None[*scip.Document](), nil
	})

	return uploadsshared.CompletedUpload{
		ID:     id,
		Commit: string(commit),
		Root:   root,
	}, lsifStore
}

func shiftSCIPRange(r scip.Range, numLines int) scip.Range {
	return scip.NewRangeUnchecked([]int32{
		r.Start.Line + int32(numLines),
		r.Start.Character,
		r.End.Line + int32(numLines),
		r.End.Character,
	})
}

func shiftPos(pos scip.Position, numLines int32) scip.Position {
	return scip.Position{
		Line:      pos.Line + numLines,
		Character: pos.Character,
	}
}

// A GitTreeTranslator that returns positions and ranges shifted by numLines
// and returns failed translations for path/range pairs if shouldFail returns true
func fakeTranslator(
	from, to api.CommitID,
	numLines int,
	shouldFail func(core.RepoRelPath, scip.Range) bool,
) GitTreeTranslator {
	translator := NewMockGitTreeTranslator()
	translator.TranslatePositionFunc.SetDefaultHook(func(ctx context.Context, f, t api.CommitID, path core.RepoRelPath, pos scip.Position) (core.Option[scip.Position], error) {
		numLines := numLines
		if f == to && t == from {
			numLines = -numLines
		}
		if shouldFail(path, scip.Range{Start: pos, End: pos}) {
			return core.None[scip.Position](), nil
		}
		return core.Some(shiftPos(pos, int32(numLines))), nil
	})
	translator.TranslateRangeFunc.SetDefaultHook(func(ctx context.Context, f, t api.CommitID, path core.RepoRelPath, range_ scip.Range) (core.Option[scip.Range], error) {
		numLines := numLines
		if f == to && t == from {
			numLines = -numLines
		}
		if shouldFail(path, range_) {
			return core.None[scip.Range](), nil
		}
		return core.Some(shiftSCIPRange(range_, numLines)), nil
	})
	return translator
}

// A GitTreeTranslator that returns all positions and ranges shifted by numLines.
func shiftAllTranslator(from, to api.CommitID, numLines int) GitTreeTranslator {
	return fakeTranslator(from, to, numLines, func(path core.RepoRelPath, range_ scip.Range) bool { return false })
}

// A GitTreeTranslator that returns all positions and ranges unchanged
func noopTranslator() GitTreeTranslator {
	return shiftAllTranslator("a", "b", 0)
}

type MatchLike interface {
	GetRange() scip.Range
	GetIsDefinition() bool
}

func expectRanges[T MatchLike](t *testing.T, matches []T, ranges ...scip.Range) {
	t.Helper()
	for _, match := range matches {
		_, err := genslices.Find(ranges, func(r scip.Range) bool {
			return match.GetRange().CompareStrict(r) == 0
		})
		require.NoErrorf(t, err, "Did not expect match at %q", match.GetRange().String())
	}

	for _, r := range ranges {
		_, err := genslices.Find(matches, func(match T) bool {
			return match.GetRange().CompareStrict(r) == 0
		})
		require.NoErrorf(t, err, "Expected match at %q", r.String())
	}
}

func expectDefinitionRanges[T MatchLike](t *testing.T, matches []T, ranges ...scip.Range) {
	t.Helper()
	for _, match := range matches {
		_, err := genslices.Find(ranges, func(r scip.Range) bool {
			return match.GetRange().CompareStrict(r) == 0
		})
		if match.GetIsDefinition() && err != nil {
			t.Errorf("Did not expect match at %q to be a definition", match.GetRange().String())
			return
		} else if !match.GetIsDefinition() && err == nil {
			t.Errorf("Expected match at %q to be a definition", match.GetRange().String())
			return
		}
	}
}
