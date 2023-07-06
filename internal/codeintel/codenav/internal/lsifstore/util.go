package lsifstore

import (
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
)

func translateRange(r *scip.Range) shared.Range {
	return newRange(int(r.Start.Line), int(r.Start.Character), int(r.End.Line), int(r.End.Character))
}

func newRange(startLine, startCharacter, endLine, endCharacter int) shared.Range {
	return shared.Range{
		Start: shared.Position{
			Line:      startLine,
			Character: startCharacter,
		},
		End: shared.Position{
			Line:      endLine,
			Character: endCharacter,
		},
	}
}
