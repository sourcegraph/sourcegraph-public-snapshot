package squirrel

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Nominal type for the ID of a tree-sitter node.
type NodeId string

// walk walks every node in the tree-sitter tree, calling f(node) on each node.
func walk(node *sitter.Node, f func(node *sitter.Node)) {
	walkFilter(node, func(n *sitter.Node) bool { f(n); return true })
}

// walkFilter walks every node in the tree-sitter tree, calling f(node) on each node and descending into
// children if it returns true.
func walkFilter(node *sitter.Node, f func(node *sitter.Node) bool) {
	if f(node) {
		for i := 0; i < int(node.ChildCount()); i++ {
			walkFilter(node.Child(i), f)
		}
	}
}

// nodeId returns the ID of the node.
func nodeId(node *sitter.Node) NodeId {
	return NodeId(fmt.Sprint(nodeToRange(node)))
}

// getRoot returns the root node of the tree-sitter tree, given any node inside it.
func getRoot(node *sitter.Node) *sitter.Node {
	var top *sitter.Node
	for cur := node; cur != nil; cur = cur.Parent() {
		top = cur
	}
	return top
}

// isLessRange compares ranges.
func isLessRange(a, b types.Range) bool {
	if a.Row == b.Row {
		return a.Column < b.Column
	}
	return a.Row < b.Row
}

// tabsToSpaces converts tabs to spaces.
func tabsToSpaces(s string) string {
	return strings.ReplaceAll(s, "\t", "    ")
}

const TAB_SIZE = 4

// lengthInSpaces returns the length of the string in spaces (using TAB_SIZE).
func lengthInSpaces(s string) int {
	total := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\t' {
			total += TAB_SIZE
		} else {
			total++
		}
	}
	return total
}

// spacesToColumn measures the length in spaces from the start of the string to the given column.
func spacesToColumn(s string, column int) int {
	total := 0
	for i := 0; i < len(s); i++ {
		if total >= column {
			return i
		}

		if s[i] == '\t' {
			total += TAB_SIZE
		} else {
			total++
		}
	}
	return total
}

// colorSprintfFunc is a color printing function.
type colorSprintfFunc func(a ...any) string

// bracket prefixes all the lines of the given string with pretty brackets.
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

// forEachCapture runs the given tree-sitter query on the given node and calls f(captureName, node) for
// each capture.
func forEachCapture(query string, node Node, f func(captureName string, node Node)) error {
	sitterQuery, err := sitter.NewQuery([]byte(query), node.LangSpec.language)
	if err != nil {
		return errors.Newf("failed to parse query: %s\n%s", err, query)
	}
	defer sitterQuery.Close()
	cursor := sitter.NewQueryCursor()
	defer cursor.Close()
	cursor.Exec(sitterQuery, node.Node)

	match, _, hasCapture := cursor.NextCapture()
	for hasCapture {
		for _, capture := range match.Captures {
			captureName := sitterQuery.CaptureNameForId(capture.Index)
			f(captureName, Node{
				RepoCommitPath: node.RepoCommitPath,
				Node:           capture.Node,
				Contents:       node.Contents,
				LangSpec:       node.LangSpec,
			})
		}
		match, _, hasCapture = cursor.NextCapture()
	}

	return nil
}

// nodeToRange returns the range of the node.
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

// nodeLength returns the length of the node.
func nodeLength(node *sitter.Node) int {
	length := 1
	if node.StartPoint().Row == node.EndPoint().Row {
		length = int(node.EndPoint().Column - node.StartPoint().Column)
	}
	return length
}

// Of course.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// When generic?
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// A sitter.Node plus convenient info.
type Node struct {
	RepoCommitPath types.RepoCommitPath
	*sitter.Node
	Contents []byte
	LangSpec LangSpec
}

func WithNode(other Node, newNode *sitter.Node) Node {
	return Node{
		RepoCommitPath: other.RepoCommitPath,
		Node:           newNode,
		Contents:       other.Contents,
		LangSpec:       other.LangSpec,
	}
}

func WithNodePtr(other Node, newNode *sitter.Node) *Node {
	return &Node{
		RepoCommitPath: other.RepoCommitPath,
		Node:           newNode,
		Contents:       other.Contents,
		LangSpec:       other.LangSpec,
	}
}

var unrecognizedFileExtensionError = errors.New("unrecognized file extension")
var unsupportedLanguageError = errors.New("unsupported language")

// Parses a file and returns info about it.
func (s *SquirrelService) parse(ctx context.Context, repoCommitPath types.RepoCommitPath) (*Node, error) {
	ext := strings.TrimPrefix(filepath.Ext(repoCommitPath.Path), ".")

	langName, ok := extToLang[ext]
	if !ok {
		return nil, unrecognizedFileExtensionError
	}

	langSpec, ok := langToLangSpec[langName]
	if !ok {
		return nil, unsupportedLanguageError
	}

	s.parser.SetLanguage(langSpec.language)

	contents, err := s.readFile(ctx, repoCommitPath)
	if err != nil {
		return nil, err
	}

	tree, err := s.parser.ParseCtx(ctx, nil, contents)
	if err != nil {
		return nil, errors.Newf("failed to parse file contents: %s", err)
	}
	s.closables = append(s.closables, tree.Close)

	root := tree.RootNode()
	if root == nil {
		return nil, errors.New("root is nil")
	}

	return &Node{RepoCommitPath: repoCommitPath, Node: root, Contents: contents, LangSpec: langSpec}, nil
}

func fatalIfError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func fatalIfErrorLabel(t *testing.T, err error, label string) {
	if err != nil {
		t.Fatalf("%s: %s", label, err)
	}
}
