package squirrel

import (
	"context"

	sitter "github.com/smacker/go-tree-sitter"
)

func (squirrel *SquirrelService) getDefPython(ctx context.Context, node Node) (ret *Node, err error) {
	defer squirrel.onCall(node, String(node.Type()), lazyNodeStringer(&ret))()

	switch node.Type() {
	case "identifier":
		ident := node.Content(node.Contents)

		cur := node.Node

		for {
			cur = cur.Parent()
			if cur == nil {
				squirrel.breadcrumb(node, "getDefJava: ran out of parents")
				return nil, nil
			}

			switch cur.Type() {

			case "module":
				found := findNodeInScope(cur, node, ident)
				if found != nil {
					return found, nil
				}
				return squirrel.getDefInImportsOrCurrentModulePython(ctx, swapNode(node, cur), ident)

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
