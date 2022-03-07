package squirrel

import (
	"context"
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"

	symbolsTypes "github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ReadFileFunc func(context.Context, types.RepoCommitPath) ([]byte, error)

type Squirrel struct {
	readFile     ReadFileFunc
	symbolSearch symbolsTypes.SearchFunc
	breadcrumbs  []Breadcrumb
}

func NewSquirrel(readFile ReadFileFunc, symbolSearch symbolsTypes.SearchFunc) *Squirrel {
	return &Squirrel{
		readFile:     readFile,
		symbolSearch: symbolSearch,
		breadcrumbs:  []Breadcrumb{},
	}
}

func (squirrel *Squirrel) symbolInfo(ctx context.Context, point types.RepoCommitPathPoint) (*types.SymbolInfo, error) {
	// First, find the definition.
	var def *types.RepoCommitPathRange
	{
		root, _, langSpec, err := parse(ctx, point.RepoCommitPath, squirrel.readFile)
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

		foundPkgOrNode, err := squirrel.getDef(ctx, langSpec.language, point.RepoCommitPath, startNode)
		if err != nil {
			return nil, err
		}
		if foundPkgOrNode == nil {
			return nil, nil
		}

		if foundPkgOrNode.Node != nil {
			def = &types.RepoCommitPathRange{
				RepoCommitPath: foundPkgOrNode.Node.RepoCommitPath,
				Range:          nodeToRange(foundPkgOrNode.Node.Node),
			}
		}
	}

	// Then get the hover if it exists.
	var hover *string
	{
		root, endContents, langSpec, err := parse(ctx, def.RepoCommitPath, squirrel.readFile)
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

		hover = findHover(endNode, langSpec.commentStyle, string(endContents))
	}

	return &types.SymbolInfo{
		Definition: *def,
		Hover:      hover,
	}, nil
}

func readFileFromGitserver(ctx context.Context, repoCommitPath types.RepoCommitPath) ([]byte, error) {
	cmd := gitserver.DefaultClient.Command("git", "cat-file", "blob", repoCommitPath.Commit+":"+repoCommitPath.Path)
	cmd.Repo = api.RepoName(repoCommitPath.Repo)
	stdout, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get file contents: %s\n\nstdout:\n\n%s\n\nstderr:\n\n%s", err, stdout, stderr)
	}
	return stdout, nil
}
