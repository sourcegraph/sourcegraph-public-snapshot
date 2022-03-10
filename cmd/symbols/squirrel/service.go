package squirrel

import (
	"context"
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"

	symbolsTypes "github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
func NewSquirrelService(readFile ReadFileFunc, symbolSearch symbolsTypes.SearchFunc) *SquirrelService {
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

// How to read a file from gitserver.
func readFileFromGitserver(ctx context.Context, repoCommitPath types.RepoCommitPath) ([]byte, error) {
	cmd := gitserver.DefaultClient.Command("git", "cat-file", "blob", repoCommitPath.Commit+":"+repoCommitPath.Path)
	cmd.Repo = api.RepoName(repoCommitPath.Repo)
	stdout, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get file contents: %s\n\nstdout:\n\n%s\n\nstderr:\n\n%s", err, stdout, stderr)
	}
	return stdout, nil
}
