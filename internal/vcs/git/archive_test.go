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

func TestArchiveReader(t *testing.T) {
	ctx := context.Background()
	repoName := MakeGitRepository(t,
		"echo abcd > file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"echo foo > file2",
		"git add file2",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:06Z git commit -m commit2 --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
		"mkdir subdir",
		"echo foo > subdir/file3",
		"git add subdir/file3",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit -m commit3 --author='a <a@a.com>' --date 2006-01-02T15:04:07Z",
	)
	commit2 := "dc2e05a9a473784075628f42ba9925643d163ec6"
	commit3 := "6a1c4f03b7ed384391b55b46633bdd2d3a323220"

	testCases := []struct {
		label                  string
		expectedFilesInArchive []string
		archiveOptions         gitserver.ArchiveOptions
	}{
		{
			label:                  "basic",
			expectedFilesInArchive: []string{"file1", "file2", "subdir/", "subdir/file3"},
			archiveOptions: gitserver.ArchiveOptions{
				Format:  ArchiveFormatZip,
				Treeish: commit3,
			},
		},
		{
			label:                  "archive at second commit only contains first two files",
			expectedFilesInArchive: []string{"file1", "file2"},
			archiveOptions: gitserver.ArchiveOptions{
				Format:  ArchiveFormatZip,
				Treeish: commit2,
			},
		},
		{
			label:                  "archive requesting subdir only contains those files",
			expectedFilesInArchive: []string{"subdir/", "subdir/file3"},
			archiveOptions: gitserver.ArchiveOptions{
				Format:  ArchiveFormatZip,
				Treeish: commit3,
				Paths:   []string{"subdir"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			readCloser, err := ArchiveReader(ctx, nil, repoName, tc.archiveOptions)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			defer readCloser.Close()
			validateFilesInZipArchive(readCloser, t, tc.expectedFilesInArchive)
		})
	}
}

func TestArchiveReaderForRepoWithSubRepoPermissionsFiltersFile(t *testing.T) {
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})
	repoName := MakeGitRepository(t,
		"echo abcd > file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"echo foo > file2",
		"git add file2",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:06Z git commit -m commit2 --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
	)
	const commit1 = "3d689662de70f9e252d4f6f1d75284e23587d670"
	const commit2 = "dc2e05a9a473784075628f42ba9925643d163ec6"

	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
		if content.Path == "file2" {
			return authz.None, nil
		}
		return authz.Read, nil
	})
	repo := &types.Repo{Name: repoName, ID: 1}

	opts := gitserver.ArchiveOptions{
		Format: ArchiveFormatZip,
		Paths:  []string{"."},
	}
	testCases := []struct {
		label                  string
		expectedFilesInArchive []string
		subrepoEnabledForRepo  bool
		commit                 string
	}{
		{
			label:                  "Filters file for repo w/ sub-repo enabled",
			expectedFilesInArchive: []string{"file1"},
			subrepoEnabledForRepo:  true,
			commit:                 commit2,
		},
		{
			label:                  "Doesn't filter file for repo w/o sub-repo enabled",
			expectedFilesInArchive: []string{"file1", "file2"},
			subrepoEnabledForRepo:  false,
			commit:                 commit2,
		},
		{
			label:                  "Archiving at first commit only contains file1",
			expectedFilesInArchive: []string{"file1"},
			subrepoEnabledForRepo:  true,
			commit:                 commit1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			opts.Treeish = tc.commit
			checker.EnabledForRepoFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName) (bool, error) {
				return tc.subrepoEnabledForRepo, nil
			})
			readCloser, err := ArchiveReader(ctx, checker, repo.Name, opts)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			defer readCloser.Close()
			validateFilesInZipArchive(readCloser, t, tc.expectedFilesInArchive)
		})
	}
}

func validateFilesInZipArchive(rc io.ReadCloser, t *testing.T, expectedFiles []string) {
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

func TestArchiveReaderForRepoWithoutSubRepoPermissions(t *testing.T) {
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})
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

	repo := &types.Repo{Name: repoName, ID: 1}

	opts := gitserver.ArchiveOptions{
		Format:  ArchiveFormatZip,
		Treeish: commitID,
		Paths:   []string{"."},
	}
	readCloser, err := ArchiveReader(ctx, checker, repo.Name, opts)
	if err != nil {
		t.Fatalf("Error should not be thrown because ArchiveReader is invoked for a repo without sub-repo permissions, got error: %s", err)
	}
	err = readCloser.Close()
	if err != nil {
		t.Error("Error during closing a reader")
	}
}
