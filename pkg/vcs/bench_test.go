package vcs_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
)

const (
	benchFileSystemCommits = 15
	benchGetCommitCommits  = 15
	benchCommitsCommits    = 15
)

func BenchmarkFileSystem_GitCmd(b *testing.B) {
	defer func() {
		b.StopTimer()
		b.StartTimer()
	}()

	cmds, files := makeGitCommandsAndFiles(benchFileSystemCommits)
	r := gitcmd.Open(context.Background(), initGitRepository(b, cmds...))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchFileSystem(b, r, "mytag", files)
	}
}

func BenchmarkGetCommit_GitCmd(b *testing.B) {
	defer func() {
		b.StopTimer()
		b.StartTimer()
	}()

	cmds, _ := makeGitCommandsAndFiles(benchGetCommitCommits)
	openRepo := func() vcs.Repository {
		return gitcmd.Open(context.Background(), initGitRepository(b, cmds...))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGetCommit(b, openRepo, "mytag")
	}
}

func BenchmarkCommits_GitCmd(b *testing.B) {
	defer func() {
		b.StopTimer()
		b.StartTimer()
	}()

	cmds, _ := makeGitCommandsAndFiles(benchCommitsCommits)
	openRepo := func() vcs.Repository {
		return gitcmd.Open(context.Background(), initGitRepository(b, cmds...))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchCommits(b, openRepo, "mytag")
	}
}

func makeGitCommandsAndFiles(n int) (cmds, files []string) {
	for i := 0; i < n; i++ {
		name := benchFilename(i)
		files = append(files, name)
		cmds = append(cmds,
			fmt.Sprintf("mkdir -p %s", filepath.Dir(name)),
			fmt.Sprintf("echo hello%d >> %s", i, name),
			fmt.Sprintf("git add %s", name),
			fmt.Sprintf("GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2014-05-06T19:20:21Z git commit -m commit%d --author='a <a@a.com>' --date 2014-05-06T19:20:21Z", i),
		)
	}
	cmds = append(cmds, "git tag mytag")
	return cmds, files
}

func benchFilename(i int) string {
	switch i % 4 {
	case 0:
		return fmt.Sprintf("file%d", i)
	case 1:
		return fmt.Sprintf("dir%d/file%d", i%10, i)
	case 2:
		return fmt.Sprintf("dir%d/subdir%d/file%d", i%7, i%3, i)
	case 3:
		return fmt.Sprintf("file%d", i%2)
	}
	panic("unreachable")
}

func benchFileSystem(b *testing.B, r vcs.Repository, tag string, files []string) {
	commitID, err := r.ResolveRevision(tag)
	if err != nil {
		b.Errorf("ResolveRevision: %s", err)
		return
	}

	fs := vcs.FileSystem(r, commitID)

	for _, f := range files {
		dir := filepath.Dir(f)

		if dir != "." {
			// dir should exist and be a dir.
			dir1Info, err := fs.Stat(dir)
			if err != nil {
				b.Errorf("fs.Stat(%q): %s", dir, err)
				return
			}
			if !dir1Info.Mode().IsDir() {
				b.Errorf("dir %q stat !IsDir", dir)
			}

			// dir should contain an entry file1.
			dirEntries, err := fs.ReadDir(dir)
			if err != nil {
				b.Errorf("fs.ReadDir(dir): %s", err)
				return
			}
			if len(dirEntries) == 0 {
				b.Errorf("dir should contain file1")
				return
			}
		}

		// file should exist, and be a file.
		file, err := fs.Open(f)
		if err != nil {
			b.Errorf("fs.Open(%q): %s", f, err)
			return
		}
		_, err = ioutil.ReadAll(file)
		if err != nil {
			b.Errorf("ReadAll(%q): %s", f, err)
			return
		}
		file.Close()

		fi, err := fs.Stat(f)
		if err != nil {
			b.Errorf("fs.Stat(%q): %s", f, err)
			return
		}
		if !fi.Mode().IsRegular() {
			b.Errorf("file %q stat !IsRegular", f)
		}
	}
}

func benchGetCommit(b *testing.B, openRepo func() vcs.Repository, tag string) {
	r := openRepo()

	commitID, err := r.ResolveRevision(tag)
	if err != nil {
		b.Errorf("ResolveRevision: %s", err)
		return
	}

	_, err = r.GetCommit(commitID)
	if err != nil {
		b.Errorf("GetCommit: %s", err)
		return
	}
}

func benchCommits(b *testing.B, openRepo func() vcs.Repository, tag string) {
	r := openRepo()

	commitID, err := r.ResolveRevision(tag)
	if err != nil {
		b.Errorf("ResolveRevision: %s", err)
		return
	}

	_, _, err = r.Commits(vcs.CommitsOptions{Head: commitID})
	if err != nil {
		b.Errorf("Commits: %s", err)
		return
	}
}
