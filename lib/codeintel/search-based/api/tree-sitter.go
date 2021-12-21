package api

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

type RecurseFunc func(n *sitter.Node)

func AssertType(n *sitter.Node, expectedType string) {
	if n.Type() != expectedType {
		panic(fmt.Sprintf("expected %v, obtained %v", expectedType, n.Type()))
	}
}

func ForeachChild(node *sitter.Node, visitor func(i int, child *sitter.Node)) {
	count := int(node.NamedChildCount())
	for i := 0; i < count; i++ {
		visitor(i, node.NamedChild(i))
	}
}
