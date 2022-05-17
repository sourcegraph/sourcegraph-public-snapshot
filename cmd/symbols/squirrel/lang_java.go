package squirrel

import (
	"context"
	"fmt"
)

func (squirrel *SquirrelService) getDefJava(ctx context.Context, node *Node) (*DirOrNode, error) {
	squirrel.breadcrumbs.add(node, fmt.Sprintf("getDefJava"))

	switch node.Type() {
	case "identifier":
		cur := node.Node
	outer:
		for {
			prev := cur
			cur = cur.Parent()
			if cur == nil {
				squirrel.breadcrumbs.add(WithNodePtr(node, prev), fmt.Sprintf("no more parents"))
				return nil, nil
			}

			squirrel.breadcrumbs.add(WithNodePtr(node, cur), fmt.Sprintf("visiting parent node %s", cur.Type()))

			switch cur.Type() {

			// Check for field access
			case "field_access":
				object := cur.ChildByFieldName("object")
				if object != nil && nodeId(prev) == nodeId(object) {
					continue
				}
				field := cur.ChildByFieldName("field")
				if field != nil {
					found, err := squirrel.getFieldJava(ctx, WithNodePtr(node, object), field.Content(node.Contents))
					if err != nil {
						return nil, err
					}
					if found != nil {
						squirrel.breadcrumbs.add(found, fmt.Sprintf("found field access"))
						return &DirOrNode{Node: found}, nil
					}
				}
				continue

			// Skip nodes that don't have bindings or are covered by other cases:
			case "argument_list":
				continue
			case "method_invocation":
				continue
			case "expression_statement":
				continue
			case "binary_expression":
				continue
			case "variable_declarator":
				continue
			case "local_variable_declaration":
				continue
			case "class_body":
				continue
			case "assignment_expression":
				continue
			case "program":
				squirrel.breadcrumbs.add(WithNodePtr(node, cur), fmt.Sprintf("reached program node, TODO check imports"))
				return nil, nil

			// Check nodes that might have bindings:
			case "block":
				blockChild := prev
				for {
					blockChild = blockChild.PrevNamedSibling()
					if blockChild == nil {
						continue outer
					}
					query := "(local_variable_declaration declarator: (variable_declarator name: (identifier) @ident))"
					captures, err := allCaptures(query, WithNodePtr(node, blockChild))
					if err != nil {
						return nil, err
					}
					for _, capture := range captures {
						if capture.Content(capture.Contents) == node.Content(node.Contents) {
							return &DirOrNode{Node: WithNodePtr(node, capture.Node)}, nil
						}
					}
				}

			case "method_declaration":
				query := `[
					(method_declaration name: (identifier) @ident)
					(formal_parameter name: (identifier) @ident)
					(spread_parameter (variable_declarator name: (identifier) @ident))
				]`
				captures, err := allCaptures(query, WithNodePtr(node, cur))
				if err != nil {
					return nil, err
				}
				for _, capture := range captures {
					if capture.Content(capture.Contents) == node.Content(node.Contents) {
						return &DirOrNode{Node: WithNodePtr(node, capture.Node)}, nil
					}
				}
				continue

			case "class_declaration":
				name := cur.ChildByFieldName("name")
				if name != nil {
					if name.Content(node.Contents) == node.Content(node.Contents) {
						return &DirOrNode{Node: WithNodePtr(node, name)}, nil
					}
				}

				body := cur.ChildByFieldName("body")
				if body == nil {
					continue
				}

				squirrel.lookupFieldJava(ctx, WithNodePtr(node, cur), node.Content(node.Contents))
				continue

			// Unrecognized nodes:
			default:
				squirrel.breadcrumbs.add(WithNodePtr(node, cur), fmt.Sprintf("unrecognized node type %q", cur.Type()))
				return nil, nil
			}
		}
	}

	return nil, nil
}

func (squirrel *SquirrelService) getFieldJava(ctx context.Context, object *Node, field string) (*Node, error) {
	ty, err := squirrel.getTypeDefJava(ctx, object)
	if err != nil {
		return nil, err
	}
	if ty == nil {
		return nil, nil
	}
	return squirrel.lookupFieldJava(ctx, ty, field)
}

func (squirrel *SquirrelService) lookupFieldJava(ctx context.Context, node *Node, field string) (*Node, error) {
	for _, child := range children(body) {
		switch child.Type() {
		case "method_declaration":
			name := child.ChildByFieldName("name")
			if name == nil {
				continue
			}
			if name.Content(node.Contents) == node.Content(node.Contents) {
				return &DirOrNode{Node: WithNodePtr(node, name)}, nil
			}
		case "class_declaration":
			name := child.ChildByFieldName("name")
			if name == nil {
				continue
			}
			if name.Content(node.Contents) == node.Content(node.Contents) {
				return &DirOrNode{Node: WithNodePtr(node, name)}, nil
			}
		case "field_declaration":
			query := "(field_declaration declarator: (variable_declarator name: (identifier) @ident))"
			captures, err := allCaptures(query, WithNodePtr(node, child))
			if err != nil {
				return nil, err
			}
			for _, capture := range captures {
				if capture.Content(capture.Contents) == node.Content(node.Contents) {
					return &DirOrNode{Node: WithNodePtr(node, capture.Node)}, nil
				}
			}
		}
	}
}

func (squirrel *SquirrelService) getTypeDefJava(ctx context.Context, node *Node) (*Node, error) {
}

func (squirrel *SquirrelService) getTypeDefJava(ctx context.Context, node *Node) (*Node, error) {
}
