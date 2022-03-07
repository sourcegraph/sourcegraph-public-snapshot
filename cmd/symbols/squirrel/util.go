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

// The ID of a tree-sitter node.
type Id = string

// walk walks every node in the tree-sitter tree, calling f on each node.
func walk(node *sitter.Node, f func(node *sitter.Node)) {
	f(node)
	for i := 0; i < int(node.ChildCount()); i++ {
		walk(node.Child(i), f)
	}
}

func nodeId(node *sitter.Node) Id {
	return fmt.Sprint(nodeToRange(node))
}

func getRoot(node *sitter.Node) *sitter.Node {
	var top *sitter.Node
	for cur := node; cur != nil; cur = cur.Parent() {
		top = cur
	}
	return top
}

func isLessRange(a, b types.Range) bool {
	if a.Row == b.Row {
		return a.Column < b.Column
	}
	return a.Row < b.Row
}

func tabsToSpaces(s string) string {
	return strings.Replace(s, "\t", "    ", -1)
}

func lengthInSpaces(s string) int {
	total := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\t' {
			total += 4
		} else {
			total++
		}
	}
	return total
}

func spacesToColumn(s string, ix int) int {
	total := 0
	for i := 0; i < len(s); i++ {
		if total >= ix {
			return i
		}

		if s[i] == '\t' {
			total += 4
		} else {
			total++
		}
	}
	return total
}

type colorSprintfFunc func(a ...interface{}) string

func bracket(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) == 1 {
		return "- " + text
	}

	for i, line := range lines {
		if i == 0 {
			lines[i] = "┌ " + line
		} else if i < len(lines)-1 {
			lines[i] = "│ " + line
		} else {
			lines[i] = "└ " + line
		}
	}

	return strings.Join(lines, "\n")
}

func forEachCapture(query string, root *sitter.Node, lang *sitter.Language, f func(captureName string, node *sitter.Node)) error {
	sitterQuery, err := sitter.NewQuery([]byte(query), lang)
	if err != nil {
		return errors.Newf("failed to parse query: %s\n%s", err, query)
	}
	cursor := sitter.NewQueryCursor()
	cursor.Exec(sitterQuery, root)

	match, _, hasCapture := cursor.NextCapture()
	for hasCapture {
		for _, capture := range match.Captures {
			captureName := sitterQuery.CaptureNameForId(capture.Index)
			f(captureName, capture.Node)
		}
		// Next capture
		match, _, hasCapture = cursor.NextCapture()
	}

	return nil
}

func nodeToRange(node *sitter.Node) types.Range {
	length := 1
	if node.StartPoint().Row == node.EndPoint().Row {
		length = int(node.EndPoint().Column - node.StartPoint().Column)
	}
	return types.Range{
		Row:    int(node.StartPoint().Row),
		Column: int(node.StartPoint().Column),
		Length: length,
	}
}

func nodeLength(node *sitter.Node) int {
	length := 1
	if node.StartPoint().Row == node.EndPoint().Row {
		length = int(node.EndPoint().Column - node.StartPoint().Column)
	}
	return length
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

type NodeWithRepoCommitPath struct {
	RepoCommitPath types.RepoCommitPath
	Node           *sitter.Node
}

func parse(ctx context.Context, repoCommitPath types.RepoCommitPath, readFile ReadFileFunc) (*sitter.Node, []byte, *LangSpec, error) {
	ext := strings.TrimPrefix(filepath.Ext(repoCommitPath.Path), ".")

	langName, ok := extToLang[ext]
	if !ok {
		return nil, nil, nil, errors.Newf("unrecognized file extension %s", ext)
	}

	langSpec, ok := langToLangSpec[langName]
	if !ok {
		return nil, nil, nil, errors.Newf("unsupported language %s", langName)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(langSpec.language)

	contents, err := readFile(ctx, repoCommitPath)
	if err != nil {
		return nil, nil, nil, err
	}

	tree, err := parser.ParseCtx(context.Background(), nil, contents)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse file contents: %s", err)
	}

	root := tree.RootNode()
	if root == nil {
		return nil, nil, nil, errors.New("root is nil")
	}

	return root, contents, &langSpec, nil
}
