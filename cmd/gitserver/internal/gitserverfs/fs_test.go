package gitserverfs

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestGitserverFS_RepoDir(t *testing.T) {
	fs := New(observation.TestContextTB(t), "/data/repos")

	tts := []struct {
		repoName api.RepoName
		want     string
	}{
		{
			repoName: "github.com/sourcegraph/sourcegraph",
			want:     "/data/repos/github.com/sourcegraph/sourcegraph/.git",
		},
		{
			repoName: "github.com/sourcegraph/sourcegraph.git",
			want:     "/data/repos/github.com/sourcegraph/sourcegraph.git/.git",
		},
		{
			repoName: "DELETED-123123.123123-github.com/sourcegraph/sourcegraph",
			want:     "/data/repos/github.com/sourcegraph/sourcegraph/.git",
		},
		{
			// This is invalid, but as a protection make sure that we still don't
			// allow a path outside of /data/repos.
			repoName: "github.com/sourcegraph/sourcegraph/../../../../../src-cli",
			want:     "/data/repos/src-cli/.git",
		},
	}

	for i, tt := range tts {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := fs.RepoDir(tt.repoName).Path()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIgnorePath(t *testing.T) {
	reposDir := "/data/repos"

	for _, tc := range []struct {
		path         string
		shouldIgnore bool
	}{
		{path: filepath.Join(reposDir, tempDirName), shouldIgnore: true},
		{path: filepath.Join(reposDir, p4HomeName), shouldIgnore: true},
		// Double check handling of trailing space
		{path: filepath.Join(reposDir, p4HomeName+"   "), shouldIgnore: true},
		{path: filepath.Join(reposDir, "sourcegraph/sourcegraph"), shouldIgnore: false},
	} {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tc.shouldIgnore, ignorePath(reposDir, tc.path))
		})
	}
}

func TestRemoveRepoDirectory(t *testing.T) {
	logger := logtest.Scoped(t)
	root := t.TempDir()

	mkFiles(t, root,
		"github.com/foo/baz/.git/HEAD",
		"github.com/foo/survivor/.git/HEAD",
		"github.com/bam/bam/.git/HEAD",
		"example.com/repo/.git/HEAD",
	)

	// Remove everything but github.com/foo/survivor
	for _, d := range []string{
		"github.com/foo/baz/.git",
		"github.com/bam/bam/.git",
		"example.com/repo/.git",
	} {
		if err := removeRepoDirectory(logger, root, common.GitDir(filepath.Join(root, d))); err != nil {
			t.Fatalf("failed to remove %s: %s", d, err)
		}
	}

	// Removing them a second time is safe
	for _, d := range []string{
		"github.com/foo/baz/.git",
		"github.com/bam/bam/.git",
		"example.com/repo/.git",
	} {
		if err := removeRepoDirectory(logger, root, common.GitDir(filepath.Join(root, d))); err != nil {
			t.Fatalf("failed to remove %s: %s", d, err)
		}
	}

	assertPaths(t, root,
		"github.com/foo/survivor/.git/HEAD",
		".tmp",
	)
}

func TestRemoveRepoDirectory_Empty(t *testing.T) {
	root := t.TempDir()

	mkFiles(t, root,
		"github.com/foo/baz/.git/HEAD",
	)
	logger := logtest.Scoped(t)

	if err := removeRepoDirectory(logger, root, common.GitDir(filepath.Join(root, "github.com/foo/baz/.git"))); err != nil {
		t.Fatal(err)
	}

	assertPaths(t, root,
		".tmp",
	)
}
