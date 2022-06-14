package squirrel

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

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

		path := getStringContents(swapNode(node, match.Captures[0].Node))
		symbol := getStringContents(swapNode(node, match.Captures[1].Node))

		pathComponents := strings.Split(path, ":")

		if len(pathComponents) != 2 {
			return nil, nil
		}

		directory := pathComponents[0]
		filename := pathComponents[1]

		if directory == "" {
			// dir of the file itself
			// TODO
			panic("blame olaf")
		}

		if !strings.HasPrefix(directory, "//") {
			fmt.Println("skipping", directory)
			continue
		}

		qualifier := strings.TrimPrefix(path, "//")

	}

	return nil, nil
}

func getStringContents(node Node) string {
	str := node.Node.Content(node.Contents)
	str = strings.TrimPrefix(str, "\"")
	str = strings.TrimSuffix(str, "\"")
	return str
}

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
