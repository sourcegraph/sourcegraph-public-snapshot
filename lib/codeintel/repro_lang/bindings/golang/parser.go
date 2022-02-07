package repro_lang

import (
	"context"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
	reproGrammar "github.com/sourcegraph/sourcegraph/lib/codeintel/repro_lang/src"
)

func parseSourceFile(ctx context.Context, source *lsif_typed.SourceFile) (*reproSourceFile, error) {
	tree, err := sitter.ParseCtx(ctx, []byte(source.Text), reproGrammar.GetLanguage())
	if err != nil {
		return nil, err
	}
	reproSource := newSourceFile(source, tree)
	reproSource.loadStatements()
	return reproSource, nil
}

func (d *reproSourceFile) loadStatements() {
	for i := uint32(0); i < d.node.ChildCount(); i++ {
		child := d.node.Child(int(i))
		name := child.ChildByFieldName("name")
		if name == nil {
			continue
		}
		switch child.Type() {
		case "definition_statement":
			docstring := ""
			docstringNode := child.ChildByFieldName("docstring")
			if docstringNode != nil {
				docstring = d.slicePosition(docstringNode)[len("# doctring:"):]
			}
			statement := &definitionStatement{
				docstring: docstring,
				name:      d.newIdentifier(child.ChildByFieldName("name")),
			}
			for i := uint32(0); i < child.NamedChildCount(); i++ {
				relation := child.NamedChild(int(i))
				switch relation.Type() {
				case "implementation_relation":
					statement.implementsRelation = d.newIdentifier(relation.ChildByFieldName("name"))
				case "type_definition_relation":
					statement.typeDefinesRelation = d.newIdentifier(relation.ChildByFieldName("name"))
				case "references_relation":
					statement.referencesRelation = d.newIdentifier(relation.ChildByFieldName("name"))
				}
			}
			d.definitions = append(d.definitions, statement)
		case "reference_statement":
			d.references = append(d.references, &referenceStatement{
				name: d.newIdentifier(child.ChildByFieldName("name")),
			})
		}
	}
}
