package squirrel

import (
	"context"
	"fmt"
	"path/filepath"
)

func (squirrel *SquirrelService) getDefJava(ctx context.Context, node Node) (ret *Node, err error) {
	defer squirrel.onCall(node, String(node.Type()), lazyNodeStringer(&ret))()

	switch node.Type() {
	case "identifier":
		ident := node.Content(node.Contents)

		cur := node.Node

	outer:
		for {
			prev := cur
			cur = cur.Parent()
			if cur == nil {
				return squirrel.symbolSearchOne(
					ctx,
					node.RepoCommitPath.Repo,
					node.RepoCommitPath.Commit,
					[]string{filepath.Dir(node.RepoCommitPath.Path)},
					ident,
				)
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
					found, err := squirrel.getFieldJava(ctx, swapNode(node, object), field.Content(node.Contents))
					if err != nil {
						return nil, err
					}
					if found != nil {
						return found, nil
					}
				}
				continue

			case "method_invocation":
				object := cur.ChildByFieldName("object")
				if object == nil {
					continue
				}
				if nodeId(prev) == nodeId(object) {
					continue
				}
				name := cur.ChildByFieldName("name")
				if name != nil {
					found, err := squirrel.getFieldJava(ctx, swapNode(node, object), name.Content(node.Contents))
					if err != nil {
						return nil, err
					}
					if found != nil {
						return found, nil
					}
				}
				continue

			// Check nodes that might have bindings:
			case "constructor_body":
				fallthrough
			case "block":
				blockChild := prev
				for {
					blockChild = blockChild.PrevNamedSibling()
					if blockChild == nil {
						continue outer
					}
					query := "(local_variable_declaration declarator: (variable_declarator name: (identifier) @ident))"
					captures, err := allCaptures(query, swapNode(node, blockChild))
					if err != nil {
						return nil, err
					}
					for _, capture := range captures {
						if capture.Content(capture.Contents) == ident {
							return swapNodePtr(node, capture.Node), nil
						}
					}
				}

			case "constructor_declaration":
				query := `[
					(constructor_declaration parameters: (formal_parameters (formal_parameter name: (identifier) @ident)))
					(constructor_declaration parameters: (formal_parameters (spread_parameter (variable_declarator name: (identifier) @ident))))
				]`
				captures, err := allCaptures(query, swapNode(node, cur))
				if err != nil {
					return nil, err
				}
				for _, capture := range captures {
					if capture.Content(capture.Contents) == ident {
						return swapNodePtr(node, capture.Node), nil
					}
				}
				continue
			case "method_declaration":
				query := `[
					(method_declaration name: (identifier) @ident)
					(method_declaration parameters: (formal_parameters (formal_parameter name: (identifier) @ident)))
					(method_declaration parameters: (formal_parameters (spread_parameter (variable_declarator name: (identifier) @ident))))
				]`
				captures, err := allCaptures(query, swapNode(node, cur))
				if err != nil {
					return nil, err
				}
				for _, capture := range captures {
					if capture.Content(capture.Contents) == ident {
						return swapNodePtr(node, capture.Node), nil
					}
				}
				continue

			case "class_declaration":
				name := cur.ChildByFieldName("name")
				if name != nil {
					if name.Content(node.Contents) == ident {
						return swapNodePtr(node, name), nil
					}
				}
				found, err := squirrel.lookupFieldJava(ctx, ClassType{def: swapNode(node, cur)}, ident)
				if err != nil {
					return nil, err
				}
				if found != nil {
					return found, nil
				}
				continue

			case "lambda_expression":
				query := `[
					(lambda_expression parameters: (identifier) @ident)
					(lambda_expression parameters: (formal_parameters (formal_parameter name: (identifier) @ident)))
					(lambda_expression parameters: (formal_parameters (spread_parameter (variable_declarator name: (identifier) @ident))))
					(lambda_expression parameters: (inferred_parameters (identifier) @ident))
				]`
				captures, err := allCaptures(query, swapNode(node, cur))
				if err != nil {
					return nil, err
				}
				for _, capture := range captures {
					if capture.Content(capture.Contents) == ident {
						return swapNodePtr(node, capture.Node), nil
					}
				}
				continue

			case "catch_clause":
				query := `(catch_clause (catch_formal_parameter name: (identifier) @ident))`
				captures, err := allCaptures(query, swapNode(node, cur))
				if err != nil {
					return nil, err
				}
				for _, capture := range captures {
					if capture.Content(capture.Contents) == ident {
						return swapNodePtr(node, capture.Node), nil
					}
				}
				continue

			case "for_statement":
				query := `(for_statement init: (local_variable_declaration declarator: (variable_declarator name: (identifier) @ident)))`
				captures, err := allCaptures(query, swapNode(node, cur))
				if err != nil {
					return nil, err
				}
				for _, capture := range captures {
					if capture.Content(capture.Contents) == ident {
						return swapNodePtr(node, capture.Node), nil
					}
				}
				continue

			case "enhanced_for_statement":
				query := `(enhanced_for_statement name: (identifier) @ident)`
				captures, err := allCaptures(query, swapNode(node, cur))
				if err != nil {
					return nil, err
				}
				for _, capture := range captures {
					if capture.Content(capture.Contents) == ident {
						return swapNodePtr(node, capture.Node), nil
					}
				}
				continue

			// Skip all other nodes
			default:
				continue
			}
		}

	case "type_identifier":
		ident := node.Content(node.Contents)

		cur := node.Node

		for {
			prev := cur
			cur = cur.Parent()
			if cur == nil {
				return squirrel.symbolSearchOne(
					ctx,
					node.RepoCommitPath.Repo,
					node.RepoCommitPath.Commit,
					[]string{filepath.Dir(node.RepoCommitPath.Path)},
					ident,
				)
			}

			switch cur.Type() {
			case "program":
				query := `[
					(program (class_declaration name: (identifier) @ident))
					(program (enum_declaration name: (identifier) @ident))
					(program (interface_declaration name: (identifier) @ident))
				]`
				captures, err := allCaptures(query, swapNode(node, cur))
				if err != nil {
					return nil, err
				}
				for _, capture := range captures {
					if capture.Content(capture.Contents) == ident {
						return swapNodePtr(node, capture.Node), nil
					}
				}
				continue
			case "class_declaration":
				query := `[
					(class_declaration name: (identifier) @ident)
					(class_declaration body: (class_body (class_declaration name: (identifier) @ident)))
					(class_declaration body: (class_body (enum_declaration name: (identifier) @ident)))
					(class_declaration body: (class_body (interface_declaration name: (identifier) @ident)))
				]`
				captures, err := allCaptures(query, swapNode(node, cur))
				if err != nil {
					return nil, err
				}
				for _, capture := range captures {
					if capture.Content(capture.Contents) == ident {
						return swapNodePtr(node, capture.Node), nil
					}
				}
				continue
			case "scoped_type_identifier":
				object := cur.Child(0)
				if object != nil && nodeId(prev) == nodeId(object) {
					continue
				}
				field := cur.Child(int(cur.ChildCount()) - 1)
				if field != nil {
					found, err := squirrel.getFieldJava(ctx, swapNode(node, object), field.Content(node.Contents))
					if err != nil {
						return nil, err
					}
					if found != nil {
						return found, nil
					}
				}
				continue
			default:
				continue
			}
		}

	// No other nodes have a definition
	default:
		return nil, nil
	}
}

func (squirrel *SquirrelService) getFieldJava(ctx context.Context, object Node, field string) (ret *Node, err error) {
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

func (squirrel *SquirrelService) lookupFieldJava(ctx context.Context, ty Type, field string) (ret *Node, err error) {
	defer squirrel.onCall(ty.node(), &Tuple{String(ty.variant()), String(field)}, lazyNodeStringer(&ret))()

	switch ty2 := ty.(type) {
	case ClassType:
		body := ty2.def.ChildByFieldName("body")
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
				if name.Content(ty2.def.Contents) == field {
					return swapNodePtr(ty2.def, name), nil
				}
			case "class_declaration":
				name := child.ChildByFieldName("name")
				if name == nil {
					continue
				}
				if name.Content(ty2.def.Contents) == field {
					return swapNodePtr(ty2.def, name), nil
				}
			case "field_declaration":
				query := "(field_declaration declarator: (variable_declarator name: (identifier) @ident))"
				captures, err := allCaptures(query, swapNode(ty2.def, child))
				if err != nil {
					return nil, err
				}
				for _, capture := range captures {
					if capture.Content(capture.Contents) == field {
						return swapNodePtr(ty2.def, capture.Node), nil
					}
				}
			}
		}
		return nil, nil
	case FnType:
		squirrel.breadcrumb(ty.node(), fmt.Sprintf("lookupFieldJava: unexpected object type %s", ty.variant()))
		return nil, nil
	case PrimType:
		squirrel.breadcrumb(ty.node(), fmt.Sprintf("lookupFieldJava: unexpected object type %s", ty.variant()))
		return nil, nil
	default:
		squirrel.breadcrumb(ty.node(), fmt.Sprintf("lookupFieldJava: unrecognized type variant %q", ty.variant()))
		return nil, nil
	}
}

func (squirrel *SquirrelService) getTypeDefJava(ctx context.Context, node Node) (ret Type, err error) {
	defer squirrel.onCall(node, String(node.Type()), lazyTypeStringer(&ret))()

	switch node.Type() {
	case "type_identifier":
		fallthrough
	case "identifier":
		found, err := squirrel.getDefJava(ctx, node)
		if err != nil {
			return nil, err
		}
		if found == nil {
			return nil, nil
		}
		return squirrel.defToType(ctx, *found)
	case "field_access":
		object := node.ChildByFieldName("object")
		if object == nil {
			return nil, nil
		}
		field := node.ChildByFieldName("field")
		if field == nil {
			return nil, nil
		}
		objectType, err := squirrel.getTypeDefJava(ctx, swapNode(node, object))
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
		if found == nil {
			return nil, nil
		}
		return squirrel.defToType(ctx, *found)
	case "method_invocation":
		name := node.ChildByFieldName("name")
		if name == nil {
			return nil, nil
		}
		ty, err := squirrel.getTypeDefJava(ctx, swapNode(node, name))
		if err != nil {
			return nil, err
		}
		if ty == nil {
			return nil, nil
		}
		switch ty2 := ty.(type) {
		case FnType:
			return ty2.ret, nil
		default:
			squirrel.breadcrumb(ty.node(), fmt.Sprintf("getTypeDefJava: expected method, got %q", ty.variant()))
			return nil, nil
		}
	case "void_type":
		return PrimType{noad: node, varient: "void"}, nil
	default:
		squirrel.breadcrumb(node, fmt.Sprintf("getTypeDefJava: unrecognized node type %q", node.Type()))
		return nil, nil
	}
}

type Type interface {
	variant() string
	node() Node
}

type FnType struct {
	ret  Type
	noad Node
}

func (t FnType) variant() string {
	return "fn"
}

func (t FnType) node() Node {
	return t.noad
}

type ClassType struct {
	def Node
}

func (t ClassType) variant() string {
	return "class"
}

func (t ClassType) node() Node {
	return t.def
}

type PrimType struct {
	noad    Node
	varient string
}

func (t PrimType) variant() string {
	return fmt.Sprintf("prim:%s", t.varient)
}

func (t PrimType) node() Node {
	return t.noad
}

func (squirrel *SquirrelService) defToType(ctx context.Context, def Node) (Type, error) {
	parent := def.Node.Parent()
	if parent == nil {
		return nil, nil
	}
	switch parent.Type() {
	case "class_declaration":
		return (Type)(ClassType{def: swapNode(def, parent)}), nil
	case "method_declaration":
		retTyNode := parent.ChildByFieldName("type")
		if retTyNode == nil {
			squirrel.breadcrumb(swapNode(def, parent), "defToType: could not find return type")
			return (Type)(FnType{
				ret:  nil,
				noad: swapNode(def, parent),
			}), nil
		}
		retTy, err := squirrel.getTypeDefJava(ctx, swapNode(def, retTyNode))
		if err != nil {
			return nil, err
		}
		return (Type)(FnType{
			ret:  retTy,
			noad: swapNode(def, parent),
		}), nil
	case "formal_parameter":
		tyNode := parent.ChildByFieldName("type")
		if tyNode == nil {
			squirrel.breadcrumb(swapNode(def, parent), "defToType: could not find parameter type")
			return nil, nil
		}
		return squirrel.getTypeDefJava(ctx, swapNode(def, tyNode))
	default:
		squirrel.breadcrumb(swapNode(def, parent), fmt.Sprintf("unrecognized def parent %q", parent.Type()))
		return nil, nil
	}
}

func lazyTypeStringer(ty *Type) func() fmt.Stringer {
	return func() fmt.Stringer {
		if ty != nil && *ty != nil {
			return String((*ty).variant())
		} else {
			return String("<nil>")
		}
	}
}
