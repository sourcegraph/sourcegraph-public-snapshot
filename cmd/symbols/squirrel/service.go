package squirrel

import (
	"context"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// How to read a file.
type ReadFileFunc func(context.Context, types.RepoCommitPath) ([]byte, error)

// SquirrelService uses tree-sitter to analyze code and collect symbols.
type SquirrelService struct {
	readFile  ReadFileFunc
	parser    *sitter.Parser
	closables []func()
}

// Creates a new SquirrelService.
func NewSquirrelService(readFile ReadFileFunc) *SquirrelService {
	return &SquirrelService{
		readFile:  readFile,
		parser:    sitter.NewParser(),
		closables: []func(){},
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
	cmd := gitserver.NewClient(nil).Command("git", "cat-file", "blob", repoCommitPath.Commit+":"+repoCommitPath.Path)
	cmd.Repo = api.RepoName(repoCommitPath.Repo)
	stdout, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		return nil, errors.Newf("failed to get file contents: %s\n\nstdout:\n\n%s\n\nstderr:\n\n%s", err, stdout, stderr)
	}
	return stdout, nil
}
