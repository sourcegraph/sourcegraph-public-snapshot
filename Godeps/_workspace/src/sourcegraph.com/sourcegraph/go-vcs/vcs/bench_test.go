package vcs_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/git"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/gitcmd"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/hg"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/hgcmd"
)

const (
	benchFileSystemCommits = 15
	benchGetCommitCommits  = 15
	benchCommitsCommits    = 15
)

func BenchmarkFileSystem_GitLibGit2(b *testing.B) {
	defer func() {
		b.StopTimer()
		b.StartTimer()
	}()

	cmds, files := makeGitCommandsAndFiles(benchFileSystemCommits)
	r := makeGitRepositoryLibGit2(b, cmds...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchFileSystem(b, r, "mytag", files)
	}
}

func BenchmarkFileSystem_GitCmd(b *testing.B) {
	defer func() {
		b.StopTimer()
		b.StartTimer()
	}()

	cmds, files := makeGitCommandsAndFiles(benchFileSystemCommits)
	r, err := gitcmd.Open(initGitRepository(b, cmds...))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchFileSystem(b, r, "mytag", files)
	}
}

func BenchmarkFileSystem_HgNative(b *testing.B) {
	defer func() {
		b.StopTimer()
		b.StartTimer()
	}()

	cmds, files := makeHgCommandsAndFiles(benchFileSystemCommits)
	r := makeHgRepositoryNative(b, cmds...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchFileSystem(b, r, "mytag", files)
	}
}

func BenchmarkFileSystem_HgCmd(b *testing.B) {
	defer func() {
		b.StopTimer()
		b.StartTimer()
	}()

	cmds, files := makeHgCommandsAndFiles(benchFileSystemCommits)
	r, err := hgcmd.Open(initHgRepository(b, cmds...))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchFileSystem(b, r, "mytag", files)
	}
}

func BenchmarkGetCommit_GitLibGit2(b *testing.B) {
	defer func() {
		b.StopTimer()
		b.StartTimer()
	}()

	cmds, _ := makeGitCommandsAndFiles(benchGetCommitCommits)
	r := makeGitRepositoryLibGit2(b, cmds...)
	openRepo := func() benchRepository {
		r, err := git.Open(r.Dir)
		if err != nil {
			b.Fatal(err)
		}
		return r
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGetCommit(b, openRepo, "mytag")
	}
}

func BenchmarkGetCommit_GitCmd(b *testing.B) {
	defer func() {
		b.StopTimer()
		b.StartTimer()
	}()

	cmds, _ := makeGitCommandsAndFiles(benchGetCommitCommits)
	openRepo := func() benchRepository {
		r, err := gitcmd.Open(initGitRepository(b, cmds...))
		if err != nil {
			b.Fatal(err)
		}
		return r
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGetCommit(b, openRepo, "mytag")
	}
}

func BenchmarkGetCommit_HgNative(b *testing.B) {
	defer func() {
		b.StopTimer()
		b.StartTimer()
	}()

	cmds, _ := makeHgCommandsAndFiles(benchGetCommitCommits)
	r := makeHgRepositoryNative(b, cmds...)
	openRepo := func() benchRepository {
		r, err := hg.Open(r.Dir)
		if err != nil {
			b.Fatal(err)
		}
		return r
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGetCommit(b, openRepo, "mytag")
	}
}

func BenchmarkGetCommit_HgCmd(b *testing.B) {
	defer func() {
		b.StopTimer()
		b.StartTimer()
	}()

	cmds, _ := makeHgCommandsAndFiles(benchGetCommitCommits)
	r := makeHgRepositoryCmd(b, cmds...)
	openRepo := func() benchRepository {
		r, err := hg.Open(r.Dir)
		if err != nil {
			b.Fatal(err)
		}
		return r
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGetCommit(b, openRepo, "mytag")
	}
}

func BenchmarkCommits_GitLibGit2(b *testing.B) {
	defer func() {
		b.StopTimer()
		b.StartTimer()
	}()

	cmds, _ := makeGitCommandsAndFiles(benchCommitsCommits)
	r := makeGitRepositoryLibGit2(b, cmds...)
	openRepo := func() benchRepository {
		r, err := git.Open(r.Dir)
		if err != nil {
			b.Fatal(err)
		}
		return r
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchCommits(b, openRepo, "mytag")
	}
}

func BenchmarkCommits_GitCmd(b *testing.B) {
	defer func() {
		b.StopTimer()
		b.StartTimer()
	}()

	cmds, _ := makeGitCommandsAndFiles(benchCommitsCommits)
	openRepo := func() benchRepository {
		r, err := gitcmd.Open(initGitRepository(b, cmds...))
		if err != nil {
			b.Fatal(err)
		}
		return r
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchCommits(b, openRepo, "mytag")
	}
}

func BenchmarkCommits_HgNative(b *testing.B) {
	defer func() {
		b.StopTimer()
		b.StartTimer()
	}()

	cmds, _ := makeHgCommandsAndFiles(benchCommitsCommits)
	r := makeHgRepositoryNative(b, cmds...)
	openRepo := func() benchRepository {
		r, err := hg.Open(r.Dir)
		if err != nil {
			b.Fatal(err)
		}
		return r
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchCommits(b, openRepo, "mytag")
	}
}

func BenchmarkCommits_HgCmd(b *testing.B) {
	defer func() {
		b.StopTimer()
		b.StartTimer()
	}()

	cmds, _ := makeHgCommandsAndFiles(benchCommitsCommits)
	openRepo := func() benchRepository {
		r, err := hgcmd.Open(initHgRepository(b, cmds...))
		if err != nil {
			b.Fatal(err)
		}
		return r
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

func makeHgCommandsAndFiles(n int) (cmds []string, files []string) {
	for i := 0; i < n; i++ {
		name := benchFilename(i)
		files = append(files, name)
		cmds = append(cmds,
			fmt.Sprintf("mkdir -p %s", filepath.Dir(name)),
			fmt.Sprintf("echo hello%d >> %s", i, name),
			fmt.Sprintf("hg add %s", name),
			fmt.Sprintf("hg commit -m hello%d --user 'a <a@a.com>' --date '2014-05-06 19:20:21 UTC'", i),
		)
	}
	cmds = append(cmds, "hg tag mytag")
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

type benchRepository interface {
	ResolveRevision(string) (vcs.CommitID, error)
	ResolveTag(string) (vcs.CommitID, error)
	GetCommit(vcs.CommitID) (*vcs.Commit, error)
	Commits(vcs.CommitsOptions) ([]*vcs.Commit, uint, error)
	FileSystem(vcs.CommitID) (vfs.FileSystem, error)
}

func benchFileSystem(b *testing.B, r benchRepository, tag string, files []string) {
	commitID, err := r.ResolveTag(tag)
	if err != nil {
		b.Errorf("ResolveTag: %s", err)
		return
	}

	fs, err := r.FileSystem(commitID)
	if err != nil {
		b.Errorf("FileSystem: %s", err)
		return
	}

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

func benchGetCommit(b *testing.B, openRepo func() benchRepository, tag string) {
	r := openRepo()

	commitID, err := r.ResolveTag(tag)
	if err != nil {
		b.Errorf("ResolveTag: %s", err)
		return
	}

	_, err = r.GetCommit(commitID)
	if err != nil {
		b.Errorf("GetCommit: %s", err)
		return
	}
}

func benchCommits(b *testing.B, openRepo func() benchRepository, tag string) {
	r := openRepo()

	commitID, err := r.ResolveTag(tag)
	if err != nil {
		b.Errorf("ResolveTag: %s", err)
		return
	}

	_, _, err = r.Commits(vcs.CommitsOptions{Head: commitID})
	if err != nil {
		b.Errorf("Commits: %s", err)
		return
	}
}
