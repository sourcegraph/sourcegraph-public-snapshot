package repro

import (
	"context"
	reproGrammar "github.com/sourcegraph/scip/cmd/scip/tests/reprolang/src"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/sourcegraph/scip/bindings/go/scip"
)

func parseSourceFile(ctx context.Context, source *scip.SourceFile) (*reproSourceFile, error) {
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
		case "relationships_statement", "definition_statement":
			docstring := ""
			docstringNode := child.ChildByFieldName("docstring")
			if docstringNode != nil {
				docstring = s.nodeText(docstringNode)[len("# doctring:"):]
			}
			name := newIdentifier(s, child.ChildByFieldName("name"))
			relations := relationships{}
			for i := uint32(0); i < child.NamedChildCount(); i++ {
				relation := child.NamedChild(int(i))
				switch relation.Type() {
				case "implementation_relation":
					relations.implementsRelation = newIdentifier(s, relation.ChildByFieldName("name"))
				case "type_definition_relation":
					relations.typeDefinesRelation = newIdentifier(s, relation.ChildByFieldName("name"))
				case "references_relation":
					relations.referencesRelation = newIdentifier(s, relation.ChildByFieldName("name"))
				case "defined_by_relation":
					relations.definedByRelation = newIdentifier(s, relation.ChildByFieldName("name"))
				}
			}
			if child.Type() == "definition_statement" {
				s.definitions = append(s.definitions, &definitionStatement{
					docstring: docstring,
					name:      name,
					relations: relations,
				})
			} else {
				s.relationships = append(s.relationships, &relationshipsStatement{
					name:      name,
					relations: relations,
				})
			}
		case "reference_statement":
			s.references = append(s.references, &referenceStatement{
				name:         newIdentifier(s, child.ChildByFieldName("name")),
				isForwardDef: child.ChildByFieldName("forward_definition") != nil,
			})
		}
	}
}
