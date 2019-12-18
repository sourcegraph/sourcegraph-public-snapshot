package git_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestRepository_FileSystem_Symlinks(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"touch file1",
		"mkdir dir1",
		"ln -s file1 link1",
		"ln -s ../file1 dir1/link2",
		"touch --date=2006-01-02T15:04:05Z file1 link1 dir1/link2 || touch -t " + Times[0] + " file1 link1 dir1/link2",
		"git add link1 file1 dir1/link2",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}

	symlinks := []string{"link1", "dir1/link2"}

	repo := MakeGitRepository(t, gitCommands...)
	commitID := api.CommitID(ComputeCommitHash(repo.URL, true))

	ctx := context.Background()

	// file1 should be a file.
	file1Info, err := git.Stat(ctx, repo, commitID, "file1")
	if err != nil {
		t.Fatalf("fs.Stat(file1): %s", err)
	}
	if !file1Info.Mode().IsRegular() {
		t.Errorf("file1 Stat !IsRegular (mode: %o)", file1Info.Mode())
	}

	checkSymlinkFileInfo := func(name string, link os.FileInfo) {
		t.Helper()
		if link.Mode()&os.ModeSymlink == 0 {
			t.Errorf("link mode is not symlink (mode: %o)", link.Mode())
		}
		if link.Name() != name {
			t.Errorf("got link.Name() == %q, want %q", link.Name(), name)
		}
	}

	// Check symlinks are links
	for _, symlink := range symlinks {
		fi, err := git.Lstat(ctx, repo, commitID, symlink)
		if err != nil {
			t.Fatalf("fs.Lstat(%s): %s", symlink, err)
		}
		if runtime.GOOS != "windows" {
			// TODO(alexsaveliev) make it work on Windows too
			checkSymlinkFileInfo(symlink, fi)
		}
	}

	// Also check the FileInfo returned by ReadDir to ensure it's
	// consistent with the FileInfo returned by Lstat.
	entries, err := git.ReadDir(ctx, repo, commitID, ".", false)
	if err != nil {
		t.Fatalf("fs.ReadDir(.): %s", err)
	}
	found := false
	for _, entry := range entries {
		if entry.Name() == "link1" {
			found = true
			if runtime.GOOS != "windows" {
				checkSymlinkFileInfo("link1", entry)
			}
		}
	}
	if !found {
		t.Fatal("readdir did not return link1")
	}

	// links stat should follow the link to file1.
	for _, symlink := range symlinks {
		fi, err := git.Stat(ctx, repo, commitID, symlink)
		if err != nil {
			t.Fatalf("fs.Stat(%s): %s", symlink, err)
		}
		if !fi.Mode().IsRegular() {
			t.Errorf("%s Stat !IsRegular (mode: %o)", symlink, fi.Mode())
		}
		if fi.Name() != symlink {
			t.Errorf("got Name %q, want %q", fi.Name(), symlink)
		}
		if fi.Size() != 0 {
			t.Errorf("got %s Size %d, want %d", symlink, fi.Size(), 0)
		}
	}
}

func TestRepository_FileSystem(t *testing.T) {
	t.Parallel()

	// In all tests, repo should contain three commits. The first commit
	// (whose ID is in the 'first' field) has a file at dir1/file1 with the
	// contents "myfile1" and the mtime 2006-01-02T15:04:05Z. The second
	// commit (whose ID is in the 'second' field) adds a file at file2 (in the
	// top-level directory of the repository) with the contents "infile2" and
	// the mtime 2014-05-06T19:20:21Z. The third commit contains an empty
	// tree.
	//
	// TODO(sqs): add symlinks, etc.
	gitCommands := []string{
		"mkdir dir1",
		"echo -n infile1 > dir1/file1",
		"touch --date=2006-01-02T15:04:05Z dir1 dir1/file1 || touch -t " + Times[0] + " dir1 dir1/file1",
		"git add dir1/file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"echo -n infile2 > 'file 2'",
		"touch --date=2014-05-06T19:20:21Z 'file 2' || touch -t " + Times[1] + " 'file 2'",
		"git add 'file 2'",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2014-05-06T19:20:21Z git commit -m commit2 --author='a <a@a.com>' --date 2014-05-06T19:20:21Z",
		"git rm 'dir1/file1' 'file 2'",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2018-05-06T19:20:21Z git commit -m commit3 --author='a <a@a.com>' --date 2018-05-06T19:20:21Z",
	}
	tests := map[string]struct {
		repo                 gitserver.Repo
		first, second, third api.CommitID
	}{
		"git cmd": {
			repo:   MakeGitRepository(t, gitCommands...),
			first:  "b6602ca96bdc0ab647278577a3c6edcb8fe18fb0",
			second: "c5151eceb40d5e625716589b745248e1a6c6228d",
			third:  "ba3c51080ed4a5b870952ecd7f0e15f255b24cca",
		},
	}

	for label, test := range tests {
		// notafile should not exist.
		if _, err := git.Stat(ctx, test.repo, test.first, "notafile"); !os.IsNotExist(err) {
			t.Errorf("%s: fs1.Stat(notafile): got err %v, want os.IsNotExist", label, err)
			continue
		}

		// dir1 should exist and be a dir.
		dir1Info, err := git.Stat(ctx, test.repo, test.first, "dir1")
		if err != nil {
			t.Errorf("%s: fs1.Stat(dir1): %s", label, err)
			continue
		}
		if !dir1Info.Mode().IsDir() {
			t.Errorf("%s: dir1 stat !IsDir", label)
		}
		if name := dir1Info.Name(); name != "dir1" {
			t.Errorf("%s: got dir1 name %q, want 'dir1'", label, name)
		}
		if dir1Info.Size() != 0 {
			t.Errorf("%s: got dir1 size %d, want 0", label, dir1Info.Size())
		}
		if got, want := "ab771ba54f5571c99ffdae54f44acc7993d9f115", dir1Info.Sys().(git.ObjectInfo).OID().String(); got != want {
			t.Errorf("%s: got dir1 OID %q, want %q", label, got, want)
		}

		// dir1 should contain one entry: file1.
		dir1Entries, err := git.ReadDir(ctx, test.repo, test.first, "dir1", false)
		if err != nil {
			t.Errorf("%s: fs1.ReadDir(dir1): %s", label, err)
			continue
		}
		if len(dir1Entries) != 1 {
			t.Errorf("%s: got %d dir1 entries, want 1", label, len(dir1Entries))
			continue
		}
		file1Info := dir1Entries[0]
		if got, want := file1Info.Name(), "dir1/file1"; got != want {
			t.Errorf("%s: got dir1 entry name == %q, want %q", label, got, want)
		}
		if want := int64(7); file1Info.Size() != want {
			t.Errorf("%s: got dir1 entry size == %d, want %d", label, file1Info.Size(), want)
		}
		if got, want := "a20cc2fb45631b1dd262371a058b1bf31702abaa", file1Info.Sys().(git.ObjectInfo).OID().String(); got != want {
			t.Errorf("%s: got dir1 entry OID %q, want %q", label, got, want)
		}

		// dir2 should not exist
		_, err = git.ReadDir(ctx, test.repo, test.first, "dir2", false)
		if !os.IsNotExist(err) {
			t.Errorf("%s: fs1.ReadDir(dir2): should not exist: %s", label, err)
			continue
		}

		// dir1/file1 should exist, contain "infile1", have the right mtime, and be a file.
		file1Data, err := git.ReadFile(ctx, test.repo, test.first, "dir1/file1", 0)
		if err != nil {
			t.Errorf("%s: fs1.ReadFile(dir1/file1): %s", label, err)
			continue
		}
		if !bytes.Equal(file1Data, []byte("infile1")) {
			t.Errorf("%s: got file1Data == %q, want %q", label, string(file1Data), "infile1")
		}
		file1Info, err = git.Stat(ctx, test.repo, test.first, "dir1/file1")
		if err != nil {
			t.Errorf("%s: fs1.Stat(dir1/file1): %s", label, err)
			continue
		}
		if !file1Info.Mode().IsRegular() {
			t.Errorf("%s: file1 stat !IsRegular", label)
		}
		if got, want := file1Info.Name(), "dir1/file1"; got != want {
			t.Errorf("%s: got file1 name %q, want %q", label, got, want)
		}
		if want := int64(7); file1Info.Size() != want {
			t.Errorf("%s: got file1 size == %d, want %d", label, file1Info.Size(), want)
		}

		// file 2 shouldn't exist in the 1st commit.
		_, err = git.ReadFile(ctx, test.repo, test.first, "file 2", 0)
		if !os.IsNotExist(err) {
			t.Errorf("%s: fs1.Open(file 2): got err %v, want os.IsNotExist (file 2 should not exist in this commit)", label, err)
		}

		// file 2 should exist in the 2nd commit.
		_, err = git.ReadFile(ctx, test.repo, test.second, "file 2", 0)
		if err != nil {
			t.Errorf("%s: fs2.Open(file 2): %s", label, err)
			continue
		}

		// file1 should also exist in the 2nd commit.
		if _, err := git.Stat(ctx, test.repo, test.second, "dir1/file1"); err != nil {
			t.Errorf("%s: fs2.Stat(dir1/file1): %s", label, err)
			continue
		}
		if _, err := git.ReadFile(ctx, test.repo, test.second, "dir1/file1", 0); err != nil {
			t.Errorf("%s: fs2.Open(dir1/file1): %s", label, err)
			continue
		}

		// root should exist (via Stat).
		root, err := git.Stat(ctx, test.repo, test.second, ".")
		if err != nil {
			t.Errorf("%s: fs2.Stat(.): %s", label, err)
			continue
		}
		if !root.Mode().IsDir() {
			t.Errorf("%s: got root !IsDir", label)
		}

		// root should have 2 entries: dir1 and file 2.
		rootEntries, err := git.ReadDir(ctx, test.repo, test.second, ".", false)
		if err != nil {
			t.Errorf("%s: fs2.ReadDir(.): %s", label, err)
			continue
		}
		if got, want := len(rootEntries), 2; got != want {
			t.Errorf("%s: got len(rootEntries) == %d, want %d", label, got, want)
			continue
		}
		if e0 := rootEntries[0]; !(e0.Name() == "dir1" && e0.Mode().IsDir()) {
			t.Errorf("%s: got root entry 0 %q IsDir=%v, want 'dir1' IsDir=true", label, e0.Name(), e0.Mode().IsDir())
		}
		if e1 := rootEntries[1]; !(e1.Name() == "file 2" && !e1.Mode().IsDir()) {
			t.Errorf("%s: got root entry 1 %q IsDir=%v, want 'file 2' IsDir=false", label, e1.Name(), e1.Mode().IsDir())
		}

		// dir1 should still only contain one entry: file1.
		dir1Entries, err = git.ReadDir(ctx, test.repo, test.second, "dir1", false)
		if err != nil {
			t.Errorf("%s: fs1.ReadDir(dir1): %s", label, err)
			continue
		}
		if len(dir1Entries) != 1 {
			t.Errorf("%s: got %d dir1 entries, want 1", label, len(dir1Entries))
			continue
		}
		if got, want := dir1Entries[0].Name(), "dir1/file1"; got != want {
			t.Errorf("%s: got dir1 entry name == %q, want %q", label, got, want)
		}

		// rootEntries should be empty for third commit
		rootEntries, err = git.ReadDir(ctx, test.repo, test.third, ".", false)
		if err != nil {
			t.Errorf("%s: fs3.ReadDir(.): %s", label, err)
			continue
		}
		if got, want := len(rootEntries), 0; got != want {
			t.Errorf("%s: got len(rootEntries) == %d, want %d", label, got, want)
			continue
		}
	}
}

func TestRepository_FileSystem_quoteChars(t *testing.T) {
	t.Parallel()

	// The repo contains 3 files: one whose filename includes a
	// non-ASCII char, one whose filename contains a double quote, and
	// one whose filename contains a backslash. These should be parsed
	// and unquoted properly.
	//
	// Filenames with double quotes are always quoted in some versions
	// of git, so we might encounter quoted paths even if
	// core.quotepath is off. We test twice, with it both on AND
	// off. (Note: Although
	// https://www.kernel.org/pub/software/scm/git/docs/git-config.html
	// says that double quotes, backslashes, and single quotes are
	// always quoted, this is not true on all git versions, such as
	// @sqs's current git version 2.7.0.)
	wantNames := []string{"⊗.txt", `".txt`, `\.txt`}
	sort.Strings(wantNames)
	gitCommands := []string{
		`touch ⊗.txt '".txt' \\.txt`,
		`git add ⊗.txt '".txt' \\.txt`,
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	tests := map[string]struct {
		repo gitserver.Repo
	}{
		"git cmd (quotepath=on)": {
			repo: MakeGitRepository(t, append([]string{"git config core.quotepath on"}, gitCommands...)...),
		},
		"git cmd (quotepath=off)": {
			repo: MakeGitRepository(t, append([]string{"git config core.quotepath off"}, gitCommands...)...),
		},
	}

	for label, test := range tests {
		commitID, err := git.ResolveRevision(ctx, test.repo, nil, "master", nil)
		if err != nil {
			t.Fatal(err)
		}

		entries, err := git.ReadDir(ctx, test.repo, commitID, ".", false)
		if err != nil {
			t.Errorf("%s: fs.ReadDir(.): %s", label, err)
			continue
		}
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		sort.Strings(names)

		if !reflect.DeepEqual(names, wantNames) {
			t.Errorf("%s: got names %v, want %v", label, names, wantNames)
			continue
		}

		for _, name := range wantNames {
			stat, err := git.Stat(ctx, test.repo, commitID, name)
			if err != nil {
				t.Errorf("%s: Stat(%q): %s", label, name, err)
				continue
			}
			if stat.Name() != name {
				t.Errorf("%s: got Name == %q, want %q", label, stat.Name(), name)
				continue
			}
		}
	}
}

func TestRepository_FileSystem_gitSubmodules(t *testing.T) {
	t.Parallel()

	submodDir := InitGitRepository(t,
		"touch f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	)
	const submodCommit = "94aa9078934ce2776ccbb589569eca5ef575f12e"

	gitCommands := []string{
		"git submodule add " + filepath.ToSlash(submodDir) + " submod",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m 'add submodule' --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	tests := map[string]struct {
		repo gitserver.Repo
	}{
		"git cmd": {
			repo: MakeGitRepository(t, gitCommands...),
		},
	}

	for label, test := range tests {
		commitID, err := git.ResolveRevision(ctx, test.repo, nil, "master", nil)
		if err != nil {
			t.Fatal(err)
		}

		checkSubmoduleFileInfo := func(label string, submod os.FileInfo) {
			if want := "submod"; submod.Name() != want {
				t.Errorf("%s: submod.Name(): got %q, want %q", label, submod.Name(), want)
			}
			// A submodule should have a special file mode and should
			// store information about its origin.
			if submod.Mode().IsRegular() {
				t.Errorf("%s: IsRegular", label)
			}
			if submod.Mode().IsDir() {
				t.Errorf("%s: IsDir", label)
			}
			if mode := submod.Mode(); mode&git.ModeSubmodule == 0 {
				t.Errorf("%s: submod.Mode(): got %o, want & git.ModeSubmodule (%o) != 0", label, mode, git.ModeSubmodule)
			}
			si, ok := submod.Sys().(git.Submodule)
			if !ok {
				t.Errorf("%s: submod.Sys(): got %v, want Submodule", label, si)
			}
			if want := filepath.ToSlash(submodDir); si.URL != want {
				t.Errorf("%s: (Submodule).URL: got %q, want %q", label, si.URL, want)
			}
			if si.CommitID != submodCommit {
				t.Errorf("%s: (Submodule).CommitID: got %q, want %q", label, si.CommitID, submodCommit)
			}
		}

		// Check the submodule os.FileInfo both when it's returned by
		// Stat and when it's returned in a list by ReadDir.
		submod, err := git.Stat(ctx, test.repo, commitID, "submod")
		if err != nil {
			t.Errorf("%s: fs.Stat(submod): %s", label, err)
			continue
		}
		checkSubmoduleFileInfo(label+" (Stat)", submod)
		entries, err := git.ReadDir(ctx, test.repo, commitID, ".", false)
		if err != nil {
			t.Errorf("%s: fs.ReadDir(.): %s", label, err)
			continue
		}
		// .gitmodules file is entries[0]
		checkSubmoduleFileInfo(label+" (ReadDir)", entries[1])

		_, err = git.ReadFile(ctx, test.repo, commitID, "submod", 0)
		if err != nil {
			t.Errorf("%s: fs.Open(submod): %s", label, err)
			continue
		}
	}
}
