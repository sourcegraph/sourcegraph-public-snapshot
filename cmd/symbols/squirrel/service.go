package squirrel

import (
	"context"

	sitter "github.com/smacker/go-tree-sitter"

	symbolsTypes "github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// How to read a file.
type ReadFileFunc func(context.Context, types.RepoCommitPath) ([]byte, error)

// SquirrelService uses tree-sitter and the symbols service to analyze and traverse files to find
// symbols.
type SquirrelService struct {
	readFile     ReadFileFunc
	symbolSearch symbolsTypes.SearchFunc
	breadcrumbs  []Breadcrumb
	parser       *sitter.Parser
	closables    []func()
}

// Creates a new SquirrelService.
func New(readFile ReadFileFunc, symbolSearch symbolsTypes.SearchFunc) *SquirrelService {
	return &SquirrelService{
		readFile:     readFile,
		symbolSearch: symbolSearch,
		breadcrumbs:  []Breadcrumb{},
		parser:       sitter.NewParser(),
		closables:    []func(){},
	}
}

// Remember to free memory allocated by tree-sitter.
func (squirrel *SquirrelService) Close() {
	for _, close := range squirrel.closables {
		close()
	}
	squirrel.parser.Close()
}

// symbolInfo finds the symbol at the given point in a file.
func (squirrel *SquirrelService) symbolInfo(ctx context.Context, point types.RepoCommitPathPoint) (*types.SymbolInfo, error) {
	// First, find the definition.
	var def *types.RepoCommitPathRange
	{
		// Parse the file and find the starting node.
		root, err := squirrel.parse(ctx, point.RepoCommitPath)
		if err != nil {
			return nil, err
		}
		startNode := root.NamedDescendantForPointRange(
			sitter.Point{Row: uint32(point.Row), Column: uint32(point.Column)},
			sitter.Point{Row: uint32(point.Row), Column: uint32(point.Column)},
		)
		if startNode == nil {
			return nil, errors.New("node is nil")
		}

		// Now find the definition.
		found, err := squirrel.getDef(ctx, WithNodePtr(*root, startNode))
		if err != nil {
			return nil, err
		}
		if found == nil {
			return nil, nil
		}
		if found.Node != nil {
			def = &types.RepoCommitPathRange{
				RepoCommitPath: found.Node.RepoCommitPath,
				Range:          nodeToRange(found.Node.Node),
			}
		}
	}

	if def == nil {
		return nil, nil
	}

	// Then get the hover if it exists.

	// Parse the END file and find the end node.
	root, err := squirrel.parse(ctx, def.RepoCommitPath)
	if err != nil {
		return nil, err
	}
	endNode := root.NamedDescendantForPointRange(
		sitter.Point{Row: uint32(def.Row), Column: uint32(def.Column)},
		sitter.Point{Row: uint32(def.Row), Column: uint32(def.Column)},
	)
	if endNode == nil {
		return nil, errors.Newf("no node at %d:%d", def.Row, def.Column)
	}

	// Now find the hover.
	result := findHover(WithNode(*root, endNode))
	hover := &result

	// We have a def, and maybe a hover.
	return &types.SymbolInfo{
		Definition: *def,
		Hover:      hover,
	}, nil
}

// How to read a file from gitserver.
func readFileFromGitserver(ctx context.Context, repoCommitPath types.RepoCommitPath) ([]byte, error) {
	cmd := gitserver.NewClient(nil).GitCommand(api.RepoName(repoCommitPath.Repo), "cat-file", "blob", repoCommitPath.Commit+":"+repoCommitPath.Path)
	stdout, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		return nil, errors.Newf("failed to get file contents: %s\n\nstdout:\n\n%s\n\nstderr:\n\n%s", err, stdout, stderr)
	}
	return stdout, nil
}

// DirOrNode is a union type that can either be a directory or a node. It's returned by getDef().
//
// - It's usually   a Node, e.g. when finding the definition of an identifier
// - It's sometimes a Dir , e.g. when finding the definition of a  Go package
type DirOrNode struct {
	Dir  *types.RepoCommitPath
	Node *Node
}

func (squirrel *SquirrelService) getDef(ctx context.Context, node *Node) (*DirOrNode, error) {
	switch node.LangSpec.name {
	// case "java":
	// case "go":
	// case "csharp":
	// case "python":
	// case "javascript":
	// case "typescript":
	// case "cpp":
	// case "ruby":
	default:
		// Language not implemented yet
		return nil, nil
	}
}
