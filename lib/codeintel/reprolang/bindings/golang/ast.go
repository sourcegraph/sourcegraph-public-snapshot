package golang

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsiftyped"
)

type definitionStatement struct {
	docstring           string
	name                *identifier
	implementsRelation  *identifier
	referencesRelation  *identifier
	typeDefinesRelation *identifier
}

func (s *definitionStatement) relationIdentifiers() []*identifier {
	return []*identifier{s.implementsRelation, s.referencesRelation, s.typeDefinesRelation}
}

type referenceStatement struct {
	name *identifier
}

type identifier struct {
	value    string
	symbol   string
	position *lsiftyped.Range
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

func NewRangePositionFromNode(node *sitter.Node) *lsiftyped.Range {
	return &lsiftyped.Range{
		Start: lsiftyped.Position{
			Line:      int32(node.StartPoint().Row),
			Character: int32(node.StartPoint().Column),
		},
		End: lsiftyped.Position{
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
