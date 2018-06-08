package git_test

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func TestOpen(t *testing.T) {
	t.Parallel()

	dir := initGitRepository(t)
	git.Open(api.RepoURI(dir), "")
}
