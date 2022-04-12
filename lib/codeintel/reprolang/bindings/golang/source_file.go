package golang

import (
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsiftyped"
)

type reproSourceFile struct {
	Source      *lsiftyped.SourceFile
	node        *sitter.Node
	definitions []*definitionStatement
	references  []*referenceStatement
	localScope  *reproScope
}

func newSourceFile(sourceFile *lsiftyped.SourceFile, node *sitter.Node) *reproSourceFile {
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
