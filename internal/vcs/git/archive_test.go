package git

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
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

	repo := &types.Repo{Name: repoName, ID: 1}

	opts := gitserver.ArchiveOptions{
		Format:  ArchiveFormatZip,
		Treeish: commitID,
		Paths:   []string{"."},
	}
	if _, err := ArchiveReader(context.Background(), checker, repo.Name, opts); err == nil {
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

	repo := &types.Repo{Name: repoName, ID: 1}

	opts := gitserver.ArchiveOptions{
		Format:  ArchiveFormatZip,
		Treeish: commitID,
		Paths:   []string{"."},
	}
	readCloser, err := ArchiveReader(context.Background(), checker, repo.Name, opts)
	if err != nil {
		t.Error("Error should not be thrown because ArchiveReader is invoked for a repo without sub-repo permissions")
	}
	err = readCloser.Close()
	if err != nil {
		t.Error("Error during closing a reader")
	}
}

func TestArchiveReaderSubRepoFiltersFiles(t *testing.T) {
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{
		UID: 1,
	})
	repoName := MakeGitRepository(t,
		"echo abcd > file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"mkdir subdir",
		"echo abcd > subdir/file2",
		"git add subdir/file2",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:06Z git commit -m commit2 --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
		"mkdir subdir_no_access",
		"echo foo > subdir_no_access/file3",
		"git add subdir_no_access/file3",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit -m commit3 --author='a <a@a.com>' --date 2006-01-02T15:04:07Z",
	)
	lastCommit := "4b3dd6935d913243907b327fc6393aff194de354"

	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.EnabledForRepoFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName) (bool, error) {
		return true, nil
	})
	checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, userID int32, content authz.RepoContent) (authz.Perms, error) {
		return authz.Read, nil
	})

	testCases := []struct {
		label                  string
		subRepoPerms           authz.SubRepoPermissions
		expectedFilesInArchive []string
		requestedPaths         []string
		permissionsFunc        func(ctx context.Context, userID int32, content authz.RepoContent) (authz.Perms, error)
	}{
		{
			label: "path includes",
			subRepoPerms: authz.SubRepoPermissions{
				PathIncludes: []string{"**/file2"},
				PathExcludes: nil,
			},
			expectedFilesInArchive: []string{"subdir/", "subdir/file2"},
		},
		{
			label: "path excludes",
			subRepoPerms: authz.SubRepoPermissions{
				PathIncludes: nil,
				PathExcludes: []string{"subdir_no_access/**"},
			},
			expectedFilesInArchive: []string{"file1", "subdir/", "subdir/file2"},
		},
		{
			label:                  "with files requested",
			subRepoPerms:           authz.SubRepoPermissions{},
			requestedPaths:         []string{"subdir/file2", "subdir_no_access/file3"},
			expectedFilesInArchive: []string{"subdir/", "subdir/file2"},
			permissionsFunc: func(ctx context.Context, userID int32, content authz.RepoContent) (authz.Perms, error) {
				if content.Path == "subdir_no_access/file3" {
					return authz.None, nil
				}
				return authz.Read, nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			checker.RawPermissionsFunc.SetDefaultHook(func(ctx context.Context, userID int32, name api.RepoName) (authz.SubRepoPermissions, error) {
				return tc.subRepoPerms, nil
			})

			if tc.permissionsFunc != nil {
				checker.PermissionsFunc.SetDefaultHook(tc.permissionsFunc)
			}
			repo := &types.Repo{Name: repoName, ID: 1}

			opts := gitserver.ArchiveOptions{
				Format:  ArchiveFormatZip,
				Treeish: lastCommit,
				Paths:   tc.requestedPaths,
			}
			readCloser, err := ArchiveReader(ctx, checker, repo.Name, opts)
			if err != nil {
				t.Fatalf("Error should not be thrown because ArchiveReader is invoked for a repo without sub-repo permissions, error: %s", err)
			}
			validateFilesInArchive(readCloser, t, tc.expectedFilesInArchive)
			err = readCloser.Close()
			if err != nil {
				t.Error("Error during closing a reader")
			}
		})
	}
}

func validateFilesInArchive(rc io.ReadCloser, t *testing.T, expectedFiles []string) {
	t.Helper()
	buff := bytes.NewBuffer([]byte{})
	size, err := io.Copy(buff, rc)
	if err != nil {
		t.Fatalf("error copying file contents: %s", err)
	}
	reader := bytes.NewReader(buff.Bytes())
	zipReader, err := zip.NewReader(reader, size)
	if err != nil {
		t.Fatalf("error creating zip reader: %s", err)
	}
	if len(zipReader.File) != len(expectedFiles) {
		t.Errorf("expected zip archive to have %d files, got %d files instead", len(expectedFiles), len(zipReader.File))
	}
	for _, zf := range zipReader.File {
		match := false
		for _, ef := range expectedFiles {
			if zf.Name == ef {
				match = true
			}
		}
		if match == false {
			t.Errorf("zip archive missing file: %s", zf.Name)
		}
	}
}
