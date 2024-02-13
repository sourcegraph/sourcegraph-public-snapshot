package gitcli

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func readFileContentsFromZip(t *testing.T, zr *zip.Reader, name string) string {
	f, err := zr.Open(name)
	if err != nil {
		t.Fatalf("File %q not found in zip archive", name)
	}

	contents, err := io.ReadAll(f)
	require.NoError(t, err)
	return string(contents)
}

func TestBuildArchiveArgs(t *testing.T) {
	t.Run("no paths", func(t *testing.T) {
		args := buildArchiveArgs(git.ArchiveFormatTar, "HEAD", nil)
		require.Equal(t, []string{"archive", "--worktree-attributes", "--format=tar", "HEAD", "--"}, args)
	})

	t.Run("with paths", func(t *testing.T) {
		args := buildArchiveArgs(git.ArchiveFormatTar, "HEAD", []string{"file1", "file2"})
		require.Equal(t, []string{"archive", "--worktree-attributes", "--format=tar", "HEAD", "--", ":(literal)file1", ":(literal)file2"}, args)
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
		"mkdir dir1",
		"echo efgh > dir1/file2",
		"git add file1",
		"git add dir1",
		"git commit -m commit --author='Foo Author <foo@sourcegraph.com>'",
	)

	commitID, err := backend.RevParseHead(ctx)
	require.NoError(t, err)

	t.Run("read simple tar archive", func(t *testing.T) {
		r, err := backend.ArchiveReader(ctx, "tar", string(commitID), nil)
		require.NoError(t, err)
		t.Cleanup(func() { r.Close() })
		tr := tar.NewReader(r)
		contents := readFileContentsFromTar(t, tr, "file1")
		require.Equal(t, "abcd\n", contents)
	})

	t.Run("read simple zip archive", func(t *testing.T) {
		r, err := backend.ArchiveReader(ctx, "zip", string(commitID), nil)
		require.NoError(t, err)
		t.Cleanup(func() { r.Close() })
		contents, err := io.ReadAll(r)
		require.NoError(t, err)
		zr, err := zip.NewReader(bytes.NewReader([]byte(contents)), int64(len(contents)))
		require.NoError(t, err)
		fileContents := readFileContentsFromZip(t, zr, "file1")
		require.Equal(t, "abcd\n", fileContents)
	})

	t.Run("read multiple files from tar archive using paths", func(t *testing.T) {
		r, err := backend.ArchiveReader(ctx, "tar", string(commitID), []string{"file1", "dir1/file2"})
		require.NoError(t, err)
		t.Cleanup(func() { r.Close() })
		tr := tar.NewReader(r)
		contents := readFileContentsFromTar(t, tr, "dir1/file2")
		require.Equal(t, "efgh\n", contents)
		r, err = backend.ArchiveReader(ctx, "tar", string(commitID), []string{"file1", "dir1/file2"})
		require.NoError(t, err)
		t.Cleanup(func() { r.Close() })
		tr = tar.NewReader(r)
		contents = readFileContentsFromTar(t, tr, "file1")
		require.Equal(t, "abcd\n", contents)
	})

	t.Run("read file in directory", func(t *testing.T) {
		r, err := backend.ArchiveReader(ctx, "tar", string(commitID), nil)
		require.NoError(t, err)
		t.Cleanup(func() { r.Close() })
		tr := tar.NewReader(r)
		contents := readFileContentsFromTar(t, tr, "dir1/file2")
		require.Equal(t, "efgh\n", contents)
	})

	t.Run("non existent commit", func(t *testing.T) {
		_, err := backend.ArchiveReader(ctx, "tar", "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", nil)
		require.Error(t, err)
		require.True(t, errors.HasType(err, &gitdomain.RevisionNotFoundError{}))
	})

	t.Run("non existent ref", func(t *testing.T) {
		_, err := backend.ArchiveReader(ctx, "tar", "head-2", nil)
		require.Error(t, err)
		require.True(t, errors.HasType(err, &gitdomain.RevisionNotFoundError{}))
	})

	t.Run("non existent file", func(t *testing.T) {
		_, err := backend.ArchiveReader(ctx, "tar", string(commitID), []string{"no-file"})
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("invalid path pattern", func(t *testing.T) {
		_, err := backend.ArchiveReader(ctx, "tar", string(commitID), []string{"dir1/*"})
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})

	// Verify that if the context is canceled, the reader returns an error.
	t.Run("context cancelation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		t.Cleanup(cancel)

		r, err := backend.ArchiveReader(ctx, git.ArchiveFormatTar, string(commitID), nil)
		require.NoError(t, err)

		cancel()

		tr := tar.NewReader(r)
		_, err = tr.Next()
		require.Error(t, err)
		require.True(t, errors.Is(err, context.Canceled), "unexpected error: %v", err)

		require.NoError(t, r.Close())
	})
}

// This repo was synthesized by hand to contain a file whose path is `.git/mydir/file2` (the Git
// CLI will not let you create a file with a `.git` path component).
//
// The synthesized bad commit is:
//
// commit aa600fc517ea6546f31ae8198beb1932f13b0e4c (HEAD -> master)
// Author: Quinn Slack <qslack@qslack.com>
//
//	Date:   Tue Jun 5 16:17:20 2018 -0700
//
// wip
//
// diff --git a/.git/mydir/file2 b/.git/mydir/file2
// new file mode 100644
// index 0000000..82b919c
// --- /dev/null
// +++ b/.git/mydir/file2
// @@ -0,0 +1 @@
// +milton
func createRepoWithDotGitDir(t *testing.T, root string) string {
	t.Helper()
	b64 := func(s string) string {
		t.Helper()
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	dir := filepath.Join(root, "remotes", "repo-with-dot-git-dir")

	files := map[string]string{
		"config": `
[core]
repositoryformatversion=0
filemode=true
`,
		"HEAD":              `ref: refs/heads/master`,
		"refs/heads/master": `aa600fc517ea6546f31ae8198beb1932f13b0e4c`,
		"objects/e7/9c5e8f964493290a409888d5413a737e8e5dd5": b64("eAFLyslPUrBgyMzLLMlMzOECACgtBOw="),
		"objects/ce/013625030ba8dba906f756967f9e9ca394464a": b64("eAFLyslPUjBjyEjNycnnAgAdxQQU"),
		"objects/82/b919c9c565d162c564286d9d6a2497931be47e": b64("eAFLyslPUjBnyM3MKcnP4wIAIw8ElA=="),
		"objects/e5/231c1d547df839dce09809e43608fe6c537682": b64("eAErKUpNVTAzYTAxAAIFvfTMEgbb8lmsKdJ+zz7ukeMOulcqZqOllmloYGBmYqKQlpmTashwjtFMlZl7xe2VbN/DptXPm7N4ipsXACOoGDo="),
		"objects/da/5ecc846359eaf23e8abe907b3125fdd7abdbc0": b64("eAErKUpNVTA2ZjA0MDAzMVFIy8xJNWJo2il58mjqxaSjKRq5c7NUpk+WflIHABZRD2I="),
		"objects/d0/01d287018593691c36042e1c8089fde7415296": b64("eAErKUpNVTA2ZjA0MDAzMVFIy8xJNWQ4x2imysy94vZKtu9h0+rnzVk8xc0LAP2TDiQ="),
		"objects/b4/009ecbf1eba01c5279f25840e2afc0d15f5005": b64("eAGdjdsJAjEQRf1OFdOAMpPN5gEitiBWEJIRBzcJu2b7N2IHfh24nMtJrRTpQA4PfWOGjEhZe4fk5zDZQGmyaDRT8ujDI7MzNOtgVdz7s21w26VWuC8xveC8vr+8/nBKrVxgyF4bJBfgiA5RjXUEO/9xVVKlS1zUB/JxNbA="),
		"objects/3d/779a05641b4ee6f1bc1e0b52de75163c2a2669": b64("eAErKUpNVTA2YjAxAAKF3MqUzCKGW3FnWpIjX32y69o3odpQ9e/11bcPAAAipRGQ"),
		"objects/aa/600fc517ea6546f31ae8198beb1932f13b0e4c": b64("eAGdjlkKAjEQBf3OKfoCSmfpLCDiFcQTZDodHHQWxwxe3xFv4FfBKx4UT8PQNzDa7doiAkLGataFXCg12lRYMEVM4qzHWMUz2eCjUXNeZGzQOdwkd1VLl1EzmZCqoehQTK6MRVMlRFJ5bbdpgcvajyNcH5nvcHy+vjz/cOBpOIEmE41D7xD2GBDVtm6BTf64qnc/qw9c4UKS"),
		"objects/e6/9de29bb2d1d6434b8b29ae775ad8c2e48c5391": b64("eAFLyslPUjBgAAAJsAHw"),
	}
	for name, data := range files {
		name = filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(name), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}
