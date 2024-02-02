package gitcli

import (
	"archive/tar"
	"context"
	"io"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/stretchr/testify/require"
)

func readFileContentsFromTar(t *testing.T, tr *tar.Reader, name string) string {
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		if h.Name == name {
			contents, err := io.ReadAll(tr)
			require.NoError(t, err)
			return string(contents)
		}
	}

	t.Fatalf("File %q not found in tar archive", name)
	return ""
}

func TestGitCLIBackend_buildArchiveArgs(t *testing.T) {
	t.Run("no pathspecs", func(t *testing.T) {
		args := buildArchiveArgs(git.ArchiveFormatTar, "HEAD", nil)
		require.Equal(t, []string{"archive", "--worktree-attributes", "--format=tar", "HEAD", "--"}, args)
	})

	t.Run("with pathspecs", func(t *testing.T) {
		args := buildArchiveArgs(git.ArchiveFormatTar, "HEAD", []string{"file1", "file2"})
		require.Equal(t, []string{"archive", "--worktree-attributes", "--format=tar", "HEAD", "--", "file1", "file2"}, args)
	})

	t.Run("zip adds -0", func(t *testing.T) {
		args := buildArchiveArgs(git.ArchiveFormatZip, "HEAD", nil)
		require.Equal(t, []string{"archive", "--worktree-attributes", "--format=zip", "-0", "HEAD", "--"}, args)
	})
}

func TestGitCLIBackend_ArchiveReader(t *testing.T) {
	ctx := context.Background()

	backend := BackendWithRepoCommands(t,
		"echo abcd > file1",
		"git add file1",
		"git commit -m commit --author='Foo Author <foo@sourcegraph.com>'",
	)

	commitID, err := backend.RevParseHead(ctx)
	require.NoError(t, err)

	t.Run("read simple archive", func(t *testing.T) {
		r, err := backend.ArchiveReader(ctx, "tar", string(commitID), nil)
		require.NoError(t, err)
		t.Cleanup(func() { r.Close() })
		tr := tar.NewReader(r)
		contents := readFileContentsFromTar(t, tr, "file1")
		require.Equal(t, "abcd\n", contents)
	})
}
