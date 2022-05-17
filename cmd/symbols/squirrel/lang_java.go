package squirrel

import (
	"context"
	"fmt"
	"strings"
)

func (squirrel *SquirrelService) getDefJava(ctx context.Context, node *Node) (ret *Node, err error) {
	defer squirrel.onCall(node, String(node.Type()), lazyNodeStringer(&ret))()

	switch node.Type() {
	case "identifier":
		cur := node.Node
	outer:
		for {
			prev := cur
			cur = cur.Parent()
			if cur == nil {
				squirrel.breadcrumb(WithNodePtr(node, prev), fmt.Sprintf("no more parents"))
				return nil, nil
			}

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
						squirrel.breadcrumb(found, fmt.Sprintf("found field access"))
						return found, nil
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
				squirrel.breadcrumb(WithNodePtr(node, cur), fmt.Sprintf("reached program node, TODO check imports"))
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
							return WithNodePtr(node, capture.Node), nil
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
						return WithNodePtr(node, capture.Node), nil
					}
				}
				continue

			case "class_declaration":
				name := cur.ChildByFieldName("name")
				if name != nil {
					if name.Content(node.Contents) == node.Content(node.Contents) {
						return WithNodePtr(node, name), nil
					}
				}
				found, err := squirrel.lookupFieldJava(ctx, (*Type)(WithNodePtr(node, cur)), node.Content(node.Contents))
				if err != nil {
					return nil, err
				}
				if found != nil {
					return found, nil
				}
				continue

			// Unrecognized nodes:
			default:
				squirrel.breadcrumb(WithNodePtr(node, cur), fmt.Sprintf("unrecognized node type %q", cur.Type()))
				return nil, nil
			}
		}
	}

	return nil, nil
}

func (squirrel *SquirrelService) getFieldJava(ctx context.Context, object *Node, field string) (ret *Node, err error) {
	defer squirrel.onCall(object, &Tuple{String(object.Type()), String(field)}, lazyNodeStringer(&ret))()

	ty, err := squirrel.getTypeDefJava(ctx, object)
	if err != nil {
		return nil, err
	}
	if ty == nil {
		return nil, nil
	}
	return squirrel.lookupFieldJava(ctx, ty, field)
}

func (squirrel *SquirrelService) lookupFieldJava(ctx context.Context, ty *Type, field string) (ret *Node, err error) {
	defer squirrel.onCall((*Node)(ty), &Tuple{String(ty.Type()), String(field)}, lazyNodeStringer(&ret))()

	switch ty.Type() {
	case "class_declaration":
		body := ty.ChildByFieldName("body")
		if body == nil {
			return nil, nil
		}
		for _, child := range children(body) {
			switch child.Type() {
			case "method_declaration":
				name := child.ChildByFieldName("name")
				if name == nil {
					continue
				}
				if name.Content(ty.Contents) == field {
					return WithNodePtr((*Node)(ty), name), nil
				}
			case "class_declaration":
				name := child.ChildByFieldName("name")
				if name == nil {
					continue
				}
				if name.Content(ty.Contents) == field {
					return WithNodePtr((*Node)(ty), name), nil
				}
			case "field_declaration":
				query := "(field_declaration declarator: (variable_declarator name: (identifier) @ident))"
				captures, err := allCaptures(query, WithNodePtr((*Node)(ty), child))
				if err != nil {
					return nil, err
				}
				for _, capture := range captures {
					if capture.Content(capture.Contents) == field {
						return WithNodePtr((*Node)(ty), capture.Node), nil
					}
				}
			}
		}
		return nil, nil
	default:
		squirrel.breadcrumb((*Node)(ty), fmt.Sprintf("lookupFieldJava: unrecognized node type %q", ty.Type()))
		return nil, nil
	}
}

func (squirrel *SquirrelService) getTypeDefJava(ctx context.Context, node *Node) (ret *Type, err error) {
	defer squirrel.onCall(node, String(node.Type()), lazyTypeStringer(&ret))()

	switch node.Type() {
	case "identifier":
		found, err := squirrel.getDefJava(ctx, node)
		if err != nil {
			return nil, err
		}
		if found == nil {
			return nil, nil
		}
		return squirrel.defToType(found), nil
	case "field_access":
		object := node.ChildByFieldName("object")
		if object == nil {
			return nil, nil
		}
		field := node.ChildByFieldName("field")
		if field == nil {
			return nil, nil
		}
		objectType, err := squirrel.getTypeDefJava(ctx, WithNodePtr(node, object))
		if err != nil {
			return nil, err
		}
		if objectType == nil {
			return nil, nil
		}
		found, err := squirrel.lookupFieldJava(ctx, objectType, field.Content(node.Contents))
		if err != nil {
			return nil, err
		}
		return squirrel.defToType(found), nil
	}

	return nil, nil
}

type Type Node

func (squirrel *SquirrelService) defToType(def *Node) *Type {
	if def == nil {
		return nil
	}
	parent := def.Node.Parent()
	if parent == nil {
		return nil
	}
	switch parent.Type() {
	case "class_declaration":
		return (*Type)(WithNodePtr(def, parent))
	default:
		squirrel.breadcrumb(WithNodePtr(def, parent), fmt.Sprintf("unrecognized def parent %q", parent.Type()))
		return nil
	}
}

type String string

func (f String) String() string {
	return string(f)
}

type Tuple []interface{}

func (t *Tuple) String() string {
	s := []string{}
	for _, v := range *t {
		s = append(s, fmt.Sprintf("%v", v))
	}
	return strings.Join(s, ", ")
}

func lazyNodeStringer(node **Node) func() fmt.Stringer {
	return func() fmt.Stringer {
		if node != nil {
			return String(fmt.Sprintf("%s ...%s...", (*node).Type(), snippet(*node)))
		} else {
			return String("<nil>")
		}
	}
}

func lazyTypeStringer(ty **Type) func() fmt.Stringer {
	return func() fmt.Stringer {
		if ty != nil {
			return String(fmt.Sprintf("%s ...%s...", (*ty).Type(), snippet((*Node)(*ty))))
		} else {
			return String("<nil>")
		}
	}
}

func snippet(node *Node) string {
	contextChars := 5
	start := node.StartByte() - uint32(contextChars)
	if start < 0 {
		start = 0
	}
	end := node.StartByte() + uint32(contextChars)
	if end > uint32(len(node.Contents)) {
		end = uint32(len(node.Contents))
	}
	ret := string(node.Contents[start:end])
	ret = strings.ReplaceAll(ret, "\n", "\\n")
	ret = strings.ReplaceAll(ret, "\t", "\\t")
	return ret
}
