package squirrel

import (
	"context"
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

func (squirrel *SquirrelService) getDefPython(ctx context.Context, node Node) (ret *Node, err error) {
	defer squirrel.onCall(node, String(node.Type()), lazyNodeStringer(&ret))()

	switch node.Type() {
	case "identifier":
		ident := node.Content(node.Contents)

		cur := node.Node

		for {
			prev := cur
			cur = cur.Parent()
			if cur == nil {
				squirrel.breadcrumb(node, "getDefPython: ran out of parents")
				return nil, nil
			}

			switch cur.Type() {

			case "module":
				found := findNodeInScope(cur, node, ident)
				if found != nil {
					return found, nil
				}
				return squirrel.getDefInImportsOrCurrentModulePython(ctx, swapNode(node, cur), ident)

			case "attribute":
				object := cur.ChildByFieldName("object")
				if object == nil {
					squirrel.breadcrumb(node, "getDefPython: attribute has no object field")
					return nil, nil
				}
				attribute := cur.ChildByFieldName("attribute")
				if attribute == nil {
					squirrel.breadcrumb(node, "getDefPython: attribute has no attribute field")
					return nil, nil
				}
				if nodeId(object) == nodeId(prev) {
					continue
				}
				return squirrel.getFieldPython(ctx, swapNode(node, object), attribute.Content(node.Contents))

			case "for_statement":
				left := cur.ChildByFieldName("left")
				if left == nil {
					continue
				}
				if left.Type() == "identifier" {
					if left.Content(node.Contents) == ident {
						return swapNodePtr(node, left), nil
					}
				}
				continue

			case "except_clause":
				if cur.NamedChildCount() < 3 {
					continue
				}
				//        vvvvvvvvv identifier
				//                     v identifier
				//                      v block
				// except Exception as e:
				exceptIdent := cur.NamedChild(1)
				if exceptIdent == nil || exceptIdent.Type() != "identifier" {
					continue
				}
				if exceptIdent.Content(node.Contents) == ident {
					return swapNodePtr(node, exceptIdent), nil
				}
				continue

			case "lambda":
				// Check the parameters
				query := `
					(lambda parameters:
						(lambda_parameters [
							(identifier) @ident
							(default_parameter name: (identifier) @ident)
							(list_splat_pattern (identifier) @ident)
							(dictionary_splat_pattern (identifier) @ident)
						])
					)
				`
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

			case "function_definition":
				// Check the function name and parameters
				name := cur.ChildByFieldName("name")
				if name != nil && name.Type() == "identifier" && name.Content(node.Contents) == ident {
					return swapNodePtr(node, name), nil
				}
				parameters := cur.ChildByFieldName("parameters")
				if parameters == nil {
					continue
				}
				query := `
					(parameters [
						(identifier) @ident
						(default_parameter name: (identifier) @ident)
						(list_splat_pattern (identifier) @ident)
						(dictionary_splat_pattern (identifier) @ident)

						(typed_parameter [
							(identifier) @ident
							(list_splat_pattern (identifier) @ident)
							(dictionary_splat_pattern (identifier) @ident)
						])
						(typed_default_parameter name: (identifier) @ident)
					])
				`
				captures, err := allCaptures(query, swapNode(node, parameters))
				if err != nil {
					return nil, err
				}
				for _, capture := range captures {
					if capture.Content(capture.Contents) == ident {
						return swapNodePtr(node, capture.Node), nil
					}
				}

				// Check the function body by doing an in-order traversal of all expression-statements
				// scoped to this function.
				body := cur.ChildByFieldName("body")
				if body == nil || body.Type() != "block" {
					squirrel.breadcrumb(swapNode(node, cur), "getDefPython: expected function_definition to have a block body")
					continue
				}
				found := findNodeInScope(body, node, ident)
				if found != nil {
					return found, nil
				}

				continue

			case "block":
				continue // Blocks are not scopes in Python, so keep looking up the tree

			// Skip all other nodes
			default:
				continue
			}
		}

	// No other nodes have a definition
	default:
		return nil, nil
	}
}

func findNodeInScope(block *sitter.Node, node Node, ident string) *Node {
	if block == nil {
		return nil
	}

	for i := 0; i < int(block.NamedChildCount()); i++ {
		child := block.NamedChild(i)

		switch child.Type() {
		case "function_definition":
			name := child.ChildByFieldName("name")
			if name != nil && name.Type() == "identifier" && name.Content(node.Contents) == ident {
				return swapNodePtr(node, name)
			}
			continue
		case "class_definition":
			name := child.ChildByFieldName("name")
			if name != nil && name.Type() == "identifier" && name.Content(node.Contents) == ident {
				return swapNodePtr(node, name)
			}
			continue
		case "expression_statement":
			query := `(expression_statement (assignment left: (identifier) @ident))`
			captures, err := allCaptures(query, swapNode(node, child))
			if err != nil {
				return nil
			}
			for _, capture := range captures {
				if capture.Content(capture.Contents) == ident {
					return swapNodePtr(node, capture.Node)
				}
			}
			continue
		case "if_statement":
			var found *Node
			found = findNodeInScope(child.ChildByFieldName("consequence"), node, ident)
			if found != nil {
				return found
			}
			elseClause := child.ChildByFieldName("alternative")
			if elseClause == nil {
				continue
			}
			found = findNodeInScope(elseClause.ChildByFieldName("body"), node, ident)
			if found != nil {
				return found
			}
			continue
		case "while_statement":
			fallthrough
		case "for_statement":
			found := findNodeInScope(child.ChildByFieldName("body"), node, ident)
			if found != nil {
				return found
			}
			continue
		case "try_statement":
			found := findNodeInScope(child.ChildByFieldName("body"), node, ident)
			if found != nil {
				return found
			}
			for j := 0; j < int(child.NamedChildCount()); j++ {
				tryChild := child.NamedChild(j)
				if tryChild.Type() == "except_clause" {
					for k := 0; k < int(tryChild.NamedChildCount()); k++ {
						exceptChild := tryChild.NamedChild(k)
						if exceptChild.Type() == "block" {
							found := findNodeInScope(exceptChild, node, ident)
							if found != nil {
								return found
							}
						}
					}
				}
			}
			continue
		default:
			continue
		}
	}

	return nil
}

func (squirrel *SquirrelService) getFieldPython(ctx context.Context, object Node, field string) (ret *Node, err error) {
	defer squirrel.onCall(object, &Tuple{String(object.Type()), String(field)}, lazyNodeStringer(&ret))()

	ty, err := squirrel.getTypeDefPython(ctx, object)
	if err != nil {
		return nil, err
	}
	if ty == nil {
		return nil, nil
	}
	return squirrel.lookupFieldPython(ctx, ty, field)
}

func (squirrel *SquirrelService) lookupFieldPython(ctx context.Context, ty TypePython, field string) (ret *Node, err error) {
	defer squirrel.onCall(ty.node(), &Tuple{String(ty.variant()), String(field)}, lazyNodeStringer(&ret))()

	switch ty2 := ty.(type) {
	case ClassTypePython:
		body := ty2.def.ChildByFieldName("body")
		if body == nil {
			return nil, nil
		}
		for _, child := range children(body) {
			switch child.Type() {
			case "expression_statement":
				query := `(expression_statement (assignment left: (identifier) @ident))`
				captures, err := allCaptures(query, swapNode(ty2.def, child))
				if err != nil {
					return nil, err
				}
				for _, capture := range captures {
					if capture.Content(capture.Contents) == field {
						return swapNodePtr(ty2.def, capture.Node), nil
					}
				}
				continue
			case "function_definition":
				name := child.ChildByFieldName("name")
				if name == nil {
					continue
				}
				if name.Content(ty2.def.Contents) == field {
					return swapNodePtr(ty2.def, name), nil
				}
			case "class_definition":
				name := child.ChildByFieldName("name")
				if name == nil {
					continue
				}
				if name.Content(ty2.def.Contents) == field {
					return swapNodePtr(ty2.def, name), nil
				}
			}
		}
		for _, super := range getSuperclassesPython(ty2.def) {
			found, err := squirrel.getFieldPython(ctx, super, field)
			if err != nil {
				return nil, err
			}
			if found != nil {
				return found, nil
			}
		}
		return nil, nil
	case FnTypePython:
		squirrel.breadcrumb(ty.node(), fmt.Sprintf("lookupFieldPython: unexpected object type %s", ty.variant()))
		return nil, nil
	case PrimTypePython:
		squirrel.breadcrumb(ty.node(), fmt.Sprintf("lookupFieldPython: unexpected object type %s", ty.variant()))
		return nil, nil
	default:
		squirrel.breadcrumb(ty.node(), fmt.Sprintf("lookupFieldPython: unrecognized type variant %q", ty.variant()))
		return nil, nil
	}
}

func (squirrel *SquirrelService) getTypeDefPython(ctx context.Context, node Node) (ret TypePython, err error) {
	defer squirrel.onCall(node, String(node.Type()), lazyTypePythonStringer(&ret))()

	onIdent := func() (TypePython, error) {
		found, err := squirrel.getDefPython(ctx, node)
		if err != nil {
			return nil, err
		}
		if found == nil {
			return nil, nil
		}
		return squirrel.defToTypePython(ctx, *found)
	}

	switch node.Type() {
	case "type":
		for _, child := range children(node.Node) {
			return squirrel.getTypeDefPython(ctx, swapNode(node, child))
		}
		return nil, nil
	case "identifier":
		return onIdent()
	case "attribute":
		object := node.ChildByFieldName("object")
		if object == nil {
			return nil, nil
		}
		attribute := node.ChildByFieldName("attribute")
		if attribute == nil {
			return nil, nil
		}
		objectType, err := squirrel.getTypeDefPython(ctx, swapNode(node, object))
		if err != nil {
			return nil, err
		}
		if objectType == nil {
			return nil, nil
		}
		found, err := squirrel.lookupFieldPython(ctx, objectType, attribute.Content(node.Contents))
		if err != nil {
			return nil, err
		}
		if found == nil {
			return nil, nil
		}
		return squirrel.defToTypePython(ctx, *found)
	case "call":
		fn := node.ChildByFieldName("function")
		if fn == nil {
			return nil, nil
		}
		ty, err := squirrel.getTypeDefPython(ctx, swapNode(node, fn))
		if err != nil {
			return nil, err
		}
		if ty == nil {
			return nil, nil
		}
		switch ty2 := ty.(type) {
		case FnTypePython:
			return ty2.ret, nil
		case ClassTypePython:
			return ty2, nil
		default:
			squirrel.breadcrumb(ty.node(), fmt.Sprintf("getTypeDefPython: expected function, got %q", ty.variant()))
			return nil, nil
		}
	// case "generic_type":
	// 	for _, child := range children(node.Node) {
	// 		if child.Type() == "type_identifier" || child.Type() == "scoped_type_identifier" {
	// 			return squirrel.getTypeDefPython(ctx, swapNode(node, child))
	// 		}
	// 	}
	// 	squirrel.breadcrumb(node, "getTypeDefPython: expected an identifier")
	// 	return nil, nil
	// case "scoped_type_identifier":
	// 	for i := int(node.NamedChildCount()) - 1; i >= 0; i-- {
	// 		child := node.NamedChild(i)
	// 		if child.Type() == "type_identifier" {
	// 			return squirrel.getTypeDefPython(ctx, swapNode(node, child))
	// 		}
	// 	}
	// 	return nil, nil
	// case "object_creation_expression":
	// 	ty := node.ChildByFieldName("type")
	// 	if ty == nil {
	// 		return nil, nil
	// 	}
	// 	return squirrel.getTypeDefPython(ctx, swapNode(node, ty))
	// case "void_type":
	// 	return PrimType{noad: node, varient: "void"}, nil
	// case "integral_type":
	// 	return PrimType{noad: node, varient: "integral"}, nil
	// case "floating_point_type":
	// 	return PrimType{noad: node, varient: "floating"}, nil
	// case "boolean_type":
	// 	return PrimType{noad: node, varient: "boolean"}, nil
	default:
		squirrel.breadcrumb(node, fmt.Sprintf("getTypeDefPython: unrecognized node type %q", node.Type()))
		return nil, nil
	}
}

func (squirrel *SquirrelService) getDefInImportsOrCurrentModulePython(ctx context.Context, program Node, ident string) (ret *Node, err error) {
	defer squirrel.onCall(program, &Tuple{String(program.Type()), String(ident)}, lazyNodeStringer(&ret))()

	query := `
		(module (function_definition name: (identifier) @a))
		(module (expression_statement (assignment left: (identifier) @definition)))
	`
	captures, err := allCaptures(query, program)
	if err != nil {
		return nil, err
	}
	for _, capture := range captures {
		if capture.Content(capture.Contents) == ident {
			return swapNodePtr(program, capture.Node), nil
		}
	}

	return nil, nil
}

func (squirrel *SquirrelService) defToTypePython(ctx context.Context, def Node) (TypePython, error) {
	parent := def.Node.Parent()
	if parent == nil {
		return nil, nil
	}
	switch parent.Type() {
	case "parameters":
		if def.Node.Type() == "identifier" && def.Node.Content(def.Contents) == "self" {
			fn := parent.Parent()
			if fn == nil || fn.Type() != "function_definition" {
				return nil, nil
			}
			block := fn.Parent()
			if block == nil || block.Type() != "block" {
				return nil, nil
			}
			class := block.Parent()
			if class == nil || class.Type() != "class_definition" {
				return nil, nil
			}
			name := class.ChildByFieldName("name")
			if name == nil {
				return nil, nil
			}
			return squirrel.defToTypePython(ctx, swapNode(def, name))
		}
		fmt.Println("TODO defToTypePython:", parent.Type())
		return nil, nil
	case "typed_parameter":
		ty := parent.ChildByFieldName("type")
		if ty == nil {
			return nil, nil
		}
		return squirrel.getTypeDefPython(ctx, swapNode(def, ty))
	case "class_definition":
		return (TypePython)(ClassTypePython{def: swapNode(def, parent)}), nil
	case "function_definition":
		retTyNode := parent.ChildByFieldName("return_type")
		if retTyNode == nil {
			return (TypePython)(FnTypePython{
				ret:  nil,
				noad: swapNode(def, parent),
			}), nil
		}
		retTy, err := squirrel.getTypeDefPython(ctx, swapNode(def, retTyNode))
		if err != nil {
			return nil, err
		}
		return (TypePython)(FnTypePython{
			ret:  retTy,
			noad: swapNode(def, parent),
		}), nil
	case "assignment":
		ty := parent.ChildByFieldName("type")
		if ty == nil {
			right := parent.ChildByFieldName("right")
			if right == nil {
				return nil, nil
			}
			return squirrel.getTypeDefPython(ctx, swapNode(def, right))
		}
		return squirrel.getTypeDefPython(ctx, swapNode(def, ty))
	default:
		squirrel.breadcrumb(swapNode(def, parent), fmt.Sprintf("unrecognized def parent %q", parent.Type()))
		return nil, nil
	}
}

func getSuperclassesPython(definition Node) []Node {
	supers := []Node{}
	for _, super := range children(definition.ChildByFieldName("superclasses")) {
		supers = append(supers, swapNode(definition, super))
	}
	return supers
}

type TypePython interface {
	variant() string
	node() Node
}

type FnTypePython struct {
	ret  TypePython
	noad Node
}

func (t FnTypePython) variant() string {
	return "fn"
}

func (t FnTypePython) node() Node {
	return t.noad
}

type ClassTypePython struct {
	def Node
}

func (t ClassTypePython) variant() string {
	return "class"
}

func (t ClassTypePython) node() Node {
	return t.def
}

type PrimTypePython struct {
	noad    Node
	varient string
}

func (t PrimTypePython) variant() string {
	return fmt.Sprintf("prim:%s", t.varient)
}

func (t PrimTypePython) node() Node {
	return t.noad
}

func lazyTypePythonStringer(ty *TypePython) func() fmt.Stringer {
	return func() fmt.Stringer {
		if ty != nil && *ty != nil {
			return String((*ty).variant())
		} else {
			return String("<nil>")
		}
	}
}
