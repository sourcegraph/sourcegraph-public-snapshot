package squirrel

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	symbolsTypes "github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// PkgOrNode is a union type that can either be a package or a node.
//
// - It's usually   a Node, e.g. when finding the definition of an "identifier"
// - It's sometimes a Pkg , e.g. when finding the definition of a  "package_identifier"
//
// It's the return type of definition calls.
type PkgOrNode struct {
	Pkg  *types.RepoCommitPath
	Node *Node
}

// getDef returns the definition of the given node.
func (s *SquirrelService) getDef(ctx context.Context, node *Node) (*PkgOrNode, error) {
	if node == nil {
		return nil, nil
	}

	addBreadcrumb(&s.breadcrumbs, *node, "getDef")

	switch node.Type() {
	case "identifier":
		for cur := node.Node; cur != nil; cur = cur.Parent() {
			parent := cur.Parent()
			if parent == nil {
				break
			}
			switch parent.Type() {
			case "block":
				for cur2 := cur; cur2 != nil; cur2 = cur2.PrevNamedSibling() {
					if cur2.Type() == "var_declaration" {
						if cur2.NamedChild(0).Type() == "var_spec" {
							found := cur2.NamedChild(0).ChildByFieldName("name")
							if found.Content(node.Contents) == node.Content(node.Contents) {
								return &PkgOrNode{Node: WithNodePtr(*node, found)}, nil
							}
						}
					}
				}
			}
		}
	case "type_identifier":
		parent := node.Parent()
		if parent == nil {
			break
		}
		switch parent.Type() {
		case "qualified_type":
			return s.getField(ctx, WithNodePtr(*node, parent.ChildByFieldName("package")), node.Content(node.Contents))
		default:
			return nil, errors.Newf("unrecognized parent type %s", parent.Type())
		}
	case "field_identifier":
		parent := node.Parent()
		if parent == nil {
			return nil, nil
		}

		switch parent.Type() {
		case "selector_expression":
			return s.getField(ctx, WithNodePtr(*node, parent.ChildByFieldName("operand")), node.Content(node.Contents))
		default:
			return nil, errors.Newf("unexpected parent type %s", parent.Type())
		}
	case "package_identifier":
		pkg := node.Content(node.Contents)
		dir := ""
		forEachCapture("(import_spec path: (interpreted_string_literal) @import)", WithNode(*node, getRoot(node.Node)), func(name string, node Node) {
			path := node.Content(node.Contents)
			path = strings.TrimPrefix(path, `"`)
			path = strings.TrimSuffix(path, `"`)

			if !strings.HasSuffix(path, "/"+pkg) {
				return
			}

			components := strings.Split(path, "/")
			if len(components) < 3 {
				return
			}

			dir = strings.Join(components[3:], "/")
		})
		if dir == "" {
			return nil, nil
		}
		return &PkgOrNode{Pkg: &types.RepoCommitPath{
			Repo:   node.RepoCommitPath.Repo,
			Commit: node.RepoCommitPath.Commit,
			Path:   dir,
		}}, nil
	}

	return nil, nil
}

// getField returns the definition of the field on the given node.
func (s *SquirrelService) getField(ctx context.Context, node *Node, field string) (*PkgOrNode, error) {
	if node == nil {
		return nil, nil
	}

	addBreadcrumb(&s.breadcrumbs, *node, fmt.Sprintf("getField(%s)", field))

	typePkgOrNode, err := s.getTypeDef(ctx, node)
	if err != nil {
		return nil, err
	}
	if typePkgOrNode == nil {
		return nil, nil
	}

	if typePkgOrNode.Pkg != nil {
		result, err := s.getDefInPkg(ctx, typePkgOrNode.Pkg.Repo, typePkgOrNode.Pkg.Commit, field, typePkgOrNode.Pkg.Path)
		if err != nil {
			return nil, err
		}
		return &PkgOrNode{Node: result}, nil
	}

	typeDef := typePkgOrNode.Node.Node

	parent := typeDef.Parent()
	if parent == nil {
		return nil, nil
	}
	switch parent.Type() {
	case "type_spec":
		ty := parent.ChildByFieldName("type")
		if ty == nil {
			return nil, nil
		}

		contents, err := s.readFile(ctx, typePkgOrNode.Node.RepoCommitPath)
		if err != nil {
			return nil, err
		}

		var foundMethod *sitter.Node
		forEachCapture("(method_declaration name: (field_identifier) @method)", WithNode(*node, getRoot(ty)), func(captureName string, node Node) {
			if node.Content(contents) == field {
				foundMethod = node.Node
			}
		})
		if foundMethod == nil {
			return nil, nil
		}
		return &PkgOrNode{Node: WithNodePtr(*typePkgOrNode.Node, foundMethod)}, nil
	default:
		return nil, errors.Newf("unrecognized type %s", typeDef.Type())
	}
}

// getTypeDef returns the definition of the type of the given node.
func (s *SquirrelService) getTypeDef(ctx context.Context, node *Node) (*PkgOrNode, error) {
	if node == nil {
		return nil, nil
	}

	addBreadcrumb(&s.breadcrumbs, *node, "getTypeDef")

	defPkgOrNode, err := s.getDef(ctx, node)
	if err != nil {
		return nil, err
	}
	if defPkgOrNode == nil {
		return nil, nil
	}

	if defPkgOrNode.Pkg != nil {
		return defPkgOrNode, nil
	}

	def := defPkgOrNode.Node.Node

	parent := def.Parent()
	if parent == nil {
		return nil, nil
	}

	switch parent.Type() {
	case "var_spec":
		ty := parent.ChildByFieldName("type")
		if ty == nil {
			return nil, nil
		}
		if ty.Type() == "pointer_type" {
			ty = ty.NamedChild(0)
			if ty == nil {
				return nil, nil
			}
		}
		switch ty.Type() {
		case "qualified_type":
			return s.getTypeDef(ctx, WithNodePtr(*defPkgOrNode.Node, ty.ChildByFieldName("name")))
		}
	case "type_spec":
		return defPkgOrNode, nil
	default:
		return nil, errors.Newf("unrecognized parent type %s", parent.Type())
	}

	return nil, nil
}

// getDefInPkg returns the definition of the symbol within the given package.
func (s *SquirrelService) getDefInPkg(ctx context.Context, repo, commit, symbolName, pkg string) (*Node, error) {
	if s.symbolSearch == nil {
		return nil, nil
	}

	defSymbols, err := s.symbolSearch(ctx, symbolsTypes.SearchArgs{
		Repo:            api.RepoName(repo),
		CommitID:        api.CommitID(commit),
		Query:           fmt.Sprintf("^%s$", symbolName),
		IsRegExp:        true,
		IsCaseSensitive: true,
		IncludePatterns: []string{"^" + pkg},
		ExcludePattern:  "",
		First:           1,
	})
	if err != nil {
		return nil, err
	}

	if len(defSymbols) == 0 {
		return nil, nil
	}

	defSymbol := defSymbols[0]

	def := types.RepoCommitPathRange{
		RepoCommitPath: types.RepoCommitPath{
			Repo:   repo,
			Commit: commit,
			Path:   defSymbol.Path,
		},
		Range: types.Range{
			Row:    int(defSymbol.Line - 1),
			Column: 0, // TODO symbol search should also return the character
			Length: len(defSymbol.Name),
		},
	}

	contents, err := s.readFile(ctx, def.RepoCommitPath)
	if err != nil {
		return nil, err
	}

	root, err := s.parse(ctx, def.RepoCommitPath, s.readFile)
	lines := strings.Split(string(contents), "\n")
	column := strings.Index(lines[def.Range.Row], defSymbol.Name)
	if column == -1 {
		return nil, nil
	}

	node := root.NamedDescendantForPointRange(
		sitter.Point{Row: uint32(def.Range.Row), Column: uint32(column)},
		sitter.Point{Row: uint32(def.Range.Row), Column: uint32(column)},
	)

	if node == nil {
		return nil, nil
	}

	return WithNodePtr(*root, node), nil
}
