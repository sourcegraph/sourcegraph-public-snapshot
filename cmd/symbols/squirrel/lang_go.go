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

type PkgOrNode struct {
	Pkg  *types.RepoCommitPath
	Node *NodeWithRepoCommitPath
}

func (s *Squirrel) getDef(ctx context.Context, lang *sitter.Language, repoCommitPath types.RepoCommitPath, node *sitter.Node) (*PkgOrNode, error) {
	if node == nil {
		return nil, nil
	}

	s.breadcrumbs = append(s.breadcrumbs, Breadcrumb{
		RepoCommitPathRange: types.RepoCommitPathRange{
			RepoCommitPath: repoCommitPath,
			Range:          nodeToRange(node),
		},
		length:  nodeLength(node),
		message: "getDef",
	})

	contents, err := s.readFile(ctx, repoCommitPath)
	if err != nil {
		return nil, err
	}

	switch node.Type() {
	case "identifier":
		for cur := node; cur != nil; cur = cur.Parent() {
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
							if found.Content(contents) == node.Content(contents) {
								return &PkgOrNode{Node: &NodeWithRepoCommitPath{RepoCommitPath: repoCommitPath, Node: found}}, nil
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
			return s.getField(ctx, lang, repoCommitPath, parent.ChildByFieldName("package"), node.Content(contents))
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
			return s.getField(ctx, lang, repoCommitPath, parent.ChildByFieldName("operand"), node.Content(contents))
		default:
			return nil, errors.Newf("unexpected parent type %s", parent.Type())
		}
	case "package_identifier":
		top := getRoot(node)
		pkg := node.Content(contents)
		dir := ""
		forEachCapture("(import_spec path: (interpreted_string_literal) @import)", top, lang, func(name string, node *sitter.Node) {
			path := node.Content(contents)
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
			Repo:   repoCommitPath.Repo,
			Commit: repoCommitPath.Commit,
			Path:   dir,
		}}, nil
	}

	return nil, nil
}

func (s *Squirrel) getField(ctx context.Context, lang *sitter.Language, repoCommitPath types.RepoCommitPath, node *sitter.Node, field string) (*PkgOrNode, error) {
	if node == nil {
		return nil, nil
	}

	s.breadcrumbs = append(s.breadcrumbs, Breadcrumb{
		RepoCommitPathRange: types.RepoCommitPathRange{
			RepoCommitPath: repoCommitPath,
			Range:          nodeToRange(node),
		},
		length:  nodeLength(node),
		message: fmt.Sprintf("getField(%s)", field),
	})

	typePkgOrNode, err := s.getTypeDef(ctx, lang, repoCommitPath, node)
	if err != nil {
		return nil, err
	}
	if typePkgOrNode == nil {
		return nil, nil
	}

	if typePkgOrNode.Pkg != nil {
		result, err := s.getDefInRepoDir(ctx, typePkgOrNode.Pkg.Repo, typePkgOrNode.Pkg.Commit, field, typePkgOrNode.Pkg.Path)
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
		forEachCapture("(method_declaration name: (field_identifier) @method)", getRoot(ty), lang, func(captureName string, node *sitter.Node) {
			if node.Content(contents) == field {
				foundMethod = node
			}
		})
		if foundMethod == nil {
			return nil, nil
		}
		return &PkgOrNode{Node: &NodeWithRepoCommitPath{RepoCommitPath: typePkgOrNode.Node.RepoCommitPath, Node: foundMethod}}, nil
	default:
		return nil, errors.Newf("unrecognized type %s", typeDef.Type())
	}
}

func (s *Squirrel) getTypeDef(ctx context.Context, lang *sitter.Language, repoCommitPath types.RepoCommitPath, node *sitter.Node) (*PkgOrNode, error) {
	if node == nil {
		return nil, nil
	}

	s.breadcrumbs = append(s.breadcrumbs, Breadcrumb{
		RepoCommitPathRange: types.RepoCommitPathRange{
			RepoCommitPath: repoCommitPath,
			Range:          nodeToRange(node),
		},
		length:  nodeLength(node),
		message: "getTypeDef",
	})

	_, err := s.readFile(ctx, repoCommitPath)
	if err != nil {
		return nil, err
	}

	defPkgOrNode, err := s.getDef(ctx, lang, repoCommitPath, node)
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
			return s.getTypeDef(ctx, lang, defPkgOrNode.Node.RepoCommitPath, ty.ChildByFieldName("name"))
		}
	case "type_spec":
		return defPkgOrNode, nil
	default:
		return nil, errors.Newf("unrecognized parent type %s", parent.Type())
	}

	return nil, nil
}

func (s *Squirrel) getDefInRepoDir(ctx context.Context, repo, commit, symbolName, dir string) (*NodeWithRepoCommitPath, error) {
	defSymbols, err := s.symbolSearch(ctx, symbolsTypes.SearchArgs{
		Repo:            api.RepoName(repo),
		CommitID:        api.CommitID(commit),
		Query:           fmt.Sprintf("^%s$", symbolName),
		IsRegExp:        true,
		IsCaseSensitive: true,
		IncludePatterns: []string{"^" + dir},
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

	root, _, _, err := parse(ctx, def.RepoCommitPath, s.readFile)
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

	return &NodeWithRepoCommitPath{RepoCommitPath: def.RepoCommitPath, Node: node}, nil
}
