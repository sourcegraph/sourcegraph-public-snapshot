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
		fmt.Println("path", path)
		fmt.Println(node.Node)
		if nodeId(match.Captures[2].Node) != nodeId(node.Node) {
			println("NOT SYMBOL")
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
		fmt.Printf("%+v\n", destinationRepoCommitPath)

		destinationRoot, err := squirrel.parse(ctx, destinationRepoCommitPath)
		if err != nil {
			return nil, err
		}
		exports, err := starlarkExports(symbol, *destinationRoot)
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

func starlarkExports(name string, node Node) (*sitter.Node, error) {
	sitterQuery, err := sitter.NewQuery([]byte(starlarkExportQuery), node.LangSpec.language)
	if err != nil {
		return nil, errors.Newf("failed to parse query: %s\n%s", err, loadQuery)
	}
	defer sitterQuery.Close()
	cursor := sitter.NewQueryCursor()
	defer cursor.Close()
	cursor.Exec(sitterQuery, getRoot(node.Node))

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
(module (function_definition name: (identifier) @name))
(module (expression_statement (assignment left: (identifier) @name)))
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
