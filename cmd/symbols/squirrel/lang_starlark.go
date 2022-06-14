package squirrel

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (squirrel *SquirrelService) getDefStarlark(ctx context.Context, node Node) (ret *Node, err error) {
	defer squirrel.onCall(node, String(node.Type()), lazyNodeStringer(&ret))()
	switch node.Type() {
	case "identifier":
		return squirrel.getDefStarlarkIdentifier(ctx, node)
	case "string":
		return squirrel.getDefStarlarkString(ctx, node)
	default:
		return nil, nil

	}
}

func (squirrel *SquirrelService) getDefStarlarkIdentifier(ctx context.Context, node Node) (ret *Node, err error) {
	symbol := node.Node.Content(node.Contents)
	root := Node{
		RepoCommitPath: node.RepoCommitPath,
		Node:           getRoot(node.Node),
		Contents:       node.Contents,
		LangSpec:       node.LangSpec,
	}
	exports, err := starlarkBindingNamed(symbol, root)
	if err != nil {
		return nil, err
	}
	if exports != nil {
		return &Node{
			RepoCommitPath: node.RepoCommitPath,
			Node:           exports,
			Contents:       node.Contents,
			LangSpec:       node.LangSpec,
		}, nil
	}
	fmt.Println(node)
	return nil, nil
}

func (squirrel *SquirrelService) getDefStarlarkString(ctx context.Context, node Node) (ret *Node, err error) {
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
			panic(match.Captures)
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
			panic(directory)
		}

		destinationRepoCommitPath := types.RepoCommitPath{
			Path:   filepath.Join(strings.TrimPrefix(directory, "//"), filename),
			Repo:   node.RepoCommitPath.Repo,
			Commit: node.RepoCommitPath.Commit,
		}

		destinationRoot, err := squirrel.parse(ctx, destinationRepoCommitPath)
		if err != nil {
			return nil, err
		}
		exports, err := starlarkBindingNamed(symbol, *destinationRoot)
		if err != nil {
			return nil, err
		}
		return &Node{
			RepoCommitPath: destinationRepoCommitPath,
			Node:           exports,
			//Node:           nil,
			Contents: node.Contents,
			LangSpec: node.LangSpec,
		}, nil
	}
}

func starlarkBindingNamed(name string, node Node) (*sitter.Node, error) {
	captures, err := allCaptures(starlarkExportQuery, node)
	if err != nil {
		return nil, err
	}
	for _, capture := range captures {
		if capture.Node.Content(capture.Contents) == name {
			return capture.Node, nil
		}
	}
	return nil, nil
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

var loadQueryKeywordArgument = `
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
