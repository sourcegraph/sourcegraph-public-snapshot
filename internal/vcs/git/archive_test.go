package git

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestArchiveReaderForRepoWithSubRepoPermissions(t *testing.T) {
	repoName := MakeGitRepository(t,
		"echo abcd > file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	)
	const commitID = "3d689662de70f9e252d4f6f1d75284e23587d670"

	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.EnabledForRepoFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName) (bool, error) {
		// sub-repo permissions are enabled only for repo with repoID = 1
		return name == repoName, nil
	})
	gitserver.ClientMocks.Archive = func(ctx context.Context, repo api.RepoName, opt gitserver.ArchiveOptions) (io.ReadCloser, error) {
		stringReader := strings.NewReader("1337")
		return io.NopCloser(stringReader), nil
	}
	defer gitserver.ResetClientMocks()

	repo := &types.Repo{Name: repoName, ID: 1}

	opts := gitserver.ArchiveOptions{
		Format:    ArchiveFormatZip,
		Treeish:   commitID,
		Pathspecs: []gitserver.Pathspec{"."},
	}
	if _, err := ArchiveReader(context.Background(), database.NewMockDB(), checker, repo.Name, opts); err == nil {
		t.Error("Error should not be null because ArchiveReader is invoked for a repo with sub-repo permissions")
	}
}

func TestArchiveReaderForRepoWithoutSubRepoPermissions(t *testing.T) {
	repoName := MakeGitRepository(t,
		"echo abcd > file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	)
	const commitID = "3d689662de70f9e252d4f6f1d75284e23587d670"

	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.EnabledForRepoFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName) (bool, error) {
		// sub-repo permissions are not present for repo with repoID = 1
		return name != repoName, nil
	})
	gitserver.ClientMocks.Archive = func(ctx context.Context, repo api.RepoName, opt gitserver.ArchiveOptions) (io.ReadCloser, error) {
		stringReader := strings.NewReader("1337")
		return io.NopCloser(stringReader), nil
	}
	defer gitserver.ResetClientMocks()

	repo := &types.Repo{Name: repoName, ID: 1}

	opts := gitserver.ArchiveOptions{
		Format:    ArchiveFormatZip,
		Treeish:   commitID,
		Pathspecs: []gitserver.Pathspec{"."},
	}
	readCloser, err := ArchiveReader(context.Background(), database.NewMockDB(), checker, repo.Name, opts)
	if err != nil {
		t.Error("Error should not be thrown because ArchiveReader is invoked for a repo without sub-repo permissions")
	}
	err = readCloser.Close()
	if err != nil {
		t.Error("Error during closing a reader")
	}
}
