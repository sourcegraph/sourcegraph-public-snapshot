package squirrel

import (
	"context"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *SquirrelService) getDefStarlark(ctx context.Context, node Node) (ret *Node, err error) {
	defer s.onCall(node, String(node.Type()), lazyNodeStringer(&ret))()
	switch node.Type() {
	case "identifier":
		return starlarkBindingNamed(node.Node.Content(node.Contents), swapNode(node, getRoot(node.Node))), nil
	case "string":
		return s.getDefStarlarkString(ctx, node)
	default:
		return nil, nil

	}
}

func (s *SquirrelService) getDefStarlarkString(ctx context.Context, node Node) (ret *Node, err error) {
	sitterQuery, err := sitter.NewQuery([]byte(loadQuery), node.LangSpec.language)
	if err != nil {
		return nil, errors.Newf("failed to parse query: %s\n%s", err, loadQuery)
	}
	defer sitterQuery.Close()
	cursor := sitter.NewQueryCursor()
	defer cursor.Close()
	cursor.Exec(sitterQuery, getRoot(node.Node))

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			return nil, nil
		}

		if len(match.Captures) < 3 {
			return nil, errors.Newf("expected 3 captures in starlark query, got %d", len(match.Captures))
		}
		path := getStringContents(swapNode(node, match.Captures[1].Node))
		symbol := getStringContents(swapNode(node, match.Captures[2].Node))
		if nodeId(match.Captures[2].Node) != nodeId(node.Node) {
			return nil, nil
		}

		pathComponents := strings.Split(path, ":")

		if len(pathComponents) != 2 {
			return nil, nil
		}

		directory := pathComponents[0]
		filename := pathComponents[1]

		if !strings.HasPrefix(directory, "//") {
			return nil, errors.Newf("expected starlark directory to be prefixed with \"//\", got %q", directory)
		}

		destinationRepoCommitPath := types.RepoCommitPath{
			Path:   filepath.Join(strings.TrimPrefix(directory, "//"), filename),
			Repo:   node.RepoCommitPath.Repo,
			Commit: node.RepoCommitPath.Commit,
		}

		destinationRoot, err := s.parse(ctx, destinationRepoCommitPath)
		if err != nil {
			return nil, err
		}
		return starlarkBindingNamed(symbol, *destinationRoot), nil //nolint:staticcheck
	}
}

func starlarkBindingNamed(name string, node Node) *Node {
	captures := allCaptures(starlarkExportQuery, node)
	for _, capture := range captures {
		if capture.Node.Content(capture.Contents) == name {
			return swapNodePtr(node, capture.Node)
		}
	}
	return nil
}

func getStringContents(node Node) string {
	str := node.Node.Content(node.Contents)
	str = strings.TrimPrefix(str, "\"")
	str = strings.TrimSuffix(str, "\"")
	return str
}

var starlarkExportQuery = `
;;; declaration
(module (function_definition name: (identifier) @name))
(module (expression_statement (assignment left: (identifier) @name)))
;;; load_statement
(
	(module
		(expression_statement
			(call
				function: (identifier) @_funcname
				arguments: (argument_list
                  (string) @path
                  (keyword_argument name: (identifier) @named) @symbol
                )
			)
		)
	)
	(#eq? @_funcname "load")
)
`

var loadQuery = `
(
	(module
		(expression_statement
			(call
				function: (identifier) @_funcname
				arguments: (argument_list (string) @path (string) @symbol)
			)
		)
	)
	(#eq? @_funcname "load")
)
`
