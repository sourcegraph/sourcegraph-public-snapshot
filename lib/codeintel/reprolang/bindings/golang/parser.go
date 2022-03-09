package golang

import (
	"context"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsiftyped"
	reproGrammar "github.com/sourcegraph/sourcegraph/lib/codeintel/reprolang/src"
)

func parseSourceFile(ctx context.Context, source *lsiftyped.SourceFile) (*reproSourceFile, error) {
	tree, err := sitter.ParseCtx(ctx, []byte(source.Text), reproGrammar.GetLanguage())
	if err != nil {
		return nil, err
	}
	reproSource := newSourceFile(source, tree)
	reproSource.loadStatements()
	return reproSource, nil
}

func (s *reproSourceFile) loadStatements() {
	for i := uint32(0); i < s.node.ChildCount(); i++ {
		child := s.node.Child(int(i))
		name := child.ChildByFieldName("name")
		if name == nil {
			continue
		}
		switch child.Type() {
		case "definition_statement":
			docstring := ""
			docstringNode := child.ChildByFieldName("docstring")
			if docstringNode != nil {
				docstring = s.nodeText(docstringNode)[len("# doctring:"):]
			}
			statement := &definitionStatement{
				docstring: docstring,
				name:      newIdentifier(s, child.ChildByFieldName("name")),
			}
			for i := uint32(0); i < child.NamedChildCount(); i++ {
				relation := child.NamedChild(int(i))
				switch relation.Type() {
				case "implementation_relation":
					statement.implementsRelation = newIdentifier(s, relation.ChildByFieldName("name"))
				case "type_definition_relation":
					statement.typeDefinesRelation = newIdentifier(s, relation.ChildByFieldName("name"))
				case "references_relation":
					statement.referencesRelation = newIdentifier(s, relation.ChildByFieldName("name"))
				}
			}
			s.definitions = append(s.definitions, statement)
		case "reference_statement":
			s.references = append(s.references, &referenceStatement{
				name: newIdentifier(s, child.ChildByFieldName("name")),
			})
		}
	}
}
