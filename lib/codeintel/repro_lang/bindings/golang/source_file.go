package golang

import (
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
)

type reproSourceFile struct {
	Source      *lsif_typed.SourceFile
	node        *sitter.Node
	definitions []*definitionStatement
	references  []*referenceStatement
	localScope  *reproScope
}

func newSourceFile(sourceFile *lsif_typed.SourceFile, node *sitter.Node) *reproSourceFile {
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
