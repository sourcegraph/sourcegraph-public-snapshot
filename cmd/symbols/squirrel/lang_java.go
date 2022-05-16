package squirrel

import (
	"context"
	"fmt"
)

func (squirrel *SquirrelService) getDefJava(ctx context.Context, node *Node) (*DirOrNode, error) {
	switch node.Type() {
	case "identifier":
		cur := node
		for {
			parent := cur.Parent()
			if parent == nil {
				return nil, nil
			}

			switch parent.Type() {
			case "argument_list":
				continue
			case "method_invocation":
				continue
			case "expression_statement":
				continue
			case "block":
				// check for locals before the node
				// for _, child := range children(parent) {
				// 	switch child.Type() {
				// 	case "local_variable_declaration":
				// 		child.ChildByFieldName("declarator")
				// 	}
				// }
				continue
			case "method_declaration":
				// check parameters
				method := parent
				for _, param := range children(method.ChildByFieldName("parameters")) {
				}

			default:
				return nil, fmt.Errorf("unrecognized node type %q", parent.Type())
			}
		}
	}

	return nil, nil
}
