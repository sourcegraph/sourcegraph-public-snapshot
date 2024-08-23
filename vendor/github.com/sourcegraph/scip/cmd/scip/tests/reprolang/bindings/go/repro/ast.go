package repro

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/sourcegraph/scip/bindings/go/scip"
)

type relationships struct {
	implementsRelation  *identifier
	referencesRelation  *identifier
	typeDefinesRelation *identifier
	definedByRelation   *identifier
}

type definitionStatement struct {
	docstring string
	name      *identifier
	relations relationships
}

func (r *relationships) identifiers() []*identifier {
	return []*identifier{r.implementsRelation, r.referencesRelation, r.typeDefinesRelation, r.definedByRelation}
}

type referenceStatement struct {
	name         *identifier
	isForwardDef bool
}

type identifier struct {
	value    string
	symbol   string
	position *scip.Range
}

func newIdentifier(s *reproSourceFile, n *sitter.Node) *identifier {
	if n == nil {
		return nil
	}
	if n.Type() != "identifier" {
		panic("expected identifier, obtained " + n.Type())
	}
	value := s.nodeText(n)
	globalIdentifier := n.ChildByFieldName("global")
	if globalIdentifier != nil {
		projectName := globalIdentifier.ChildByFieldName("project_name")
		descriptors := globalIdentifier.ChildByFieldName("descriptors")
		if projectName != nil && descriptors != nil {
			value = fmt.Sprintf("global %v %v", s.nodeText(projectName), s.nodeText(descriptors))
		}
	}
	return &identifier{
		value:    value,
		position: NewRangePositionFromNode(n),
	}
}

type relationshipsStatement struct {
	name      *identifier
	relations relationships
}

func NewRangePositionFromNode(node *sitter.Node) *scip.Range {
	return &scip.Range{
		Start: scip.Position{
			Line:      int32(node.StartPoint().Row),
			Character: int32(node.StartPoint().Column),
		},
		End: scip.Position{
			Line:      int32(node.EndPoint().Row),
			Character: int32(node.EndPoint().Column),
		},
	}
}

func (i *identifier) resolveSymbol(localScope *reproScope, context *reproContext) {
	scope := context.globalScope
	if i.isLocalSymbol() {
		scope = localScope
	}
	symbol, ok := scope.names[i.value]
	if !ok {
		symbol = "local ERROR_UNRESOLVED_SYMBOL"
	}
	i.symbol = symbol
}

func (i *identifier) isLocalSymbol() bool {
	return strings.HasPrefix(i.value, "local")
}
