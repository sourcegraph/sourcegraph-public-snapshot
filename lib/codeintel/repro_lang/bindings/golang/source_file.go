package repro_lang

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
)

type reproSourceFile struct {
	Source       *lsif_typed.SourceFile
	node         *sitter.Node
	definitions  []*definitionStatement
	references   []*referenceStatement
	localScope   *reproScope
	localCounter int
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

func (d *reproSourceFile) enterNewLocalSymbol(name identifier) string {
	symbol := fmt.Sprintf("local %v", name.value[len("local"):])
	d.localScope.names[name.value] = symbol
	return symbol
}

func (d *reproSourceFile) slicePosition(n *sitter.Node) string {
	return d.Source.Text[n.StartByte():n.EndByte()]
}
func (d *reproSourceFile) newIdentifier(n *sitter.Node) *identifier {
	if n == nil {
		return nil
	}
	if n.Type() != "identifier" {
		panic("expected identifier, obtained " + n.Type())
	}
	value := d.slicePosition(n)
	globalIdentifier := n.ChildByFieldName("global")
	if globalIdentifier != nil {
		projectName := globalIdentifier.ChildByFieldName("project_name")
		descriptors := globalIdentifier.ChildByFieldName("descriptors")
		if projectName != nil && descriptors != nil {
			value = fmt.Sprintf("global %v %v", d.slicePosition(projectName), d.slicePosition(descriptors))
		}
	}
	return &identifier{
		value:    value,
		position: lsif_typed.NewRangePositionFromNode(n),
	}
}
