package repro

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/sourcegraph/scip/bindings/go/scip"
)

type reproSourceFile struct {
	Source        *scip.SourceFile
	node          *sitter.Node
	definitions   []*definitionStatement
	references    []*referenceStatement
	relationships []*relationshipsStatement
	localScope    *reproScope
}

func newSourceFile(sourceFile *scip.SourceFile, node *sitter.Node) *reproSourceFile {
	return &reproSourceFile{
		Source:      sourceFile,
		node:        node,
		definitions: nil,
		references:  nil,
		localScope:  newScope(),
	}
}

func (s *reproSourceFile) nodeText(n *sitter.Node) string {
	return s.Source.Text[n.StartByte():n.EndByte()]
}
