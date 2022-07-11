package stitch

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type gitFS struct {
	rev    string
	dir    string
	prefix string
}

// used to mock in tests
var newGitFS = defaultNewGitFS

// defaultNewGitFS returns a mock fileystem that uses `git show` to produce directory
// and file content. The given dir is used as the working directory of the git subcommand.
// The given prefix is the relative directory that roots the filesystem.
func defaultNewGitFS(rev, dir, prefix string) fs.FS {
	return &gitFS{rev, dir, prefix}
}

var gitTreePattern = regexp.MustCompile("^tree .+:.+\n")

func (fs *gitFS) Open(name string) (fs.File, error) {
	if name == "." || name == "/" {
		name = ""
	}
	revPath := makeRevPath(fs.rev, filepath.Join(fs.prefix, name))

	cmd := exec.Command("git", "show", revPath)
	cmd.Dir = fs.dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	if ok := gitTreePattern.Match(out); ok {
		lines := bytes.Split(out, []byte("\n"))

		entries := make([]string, 0, len(lines))
		for _, line := range lines {
			if len(line) == 0 {
				continue
			}

			entries = append(entries, strings.TrimRight(string(line), string(os.PathSeparator)))
		}

		return &gitFSDir{
			name:    name,
			entries: entries,
		}, nil
	}

	return &gitFSFile{
		name:       name,
		ReadCloser: io.NopCloser(bytes.NewReader(out)),
	}, nil
}

type gitFSDir struct {
	name    string
	entries []string
	offset  int
}

func (d *gitFSDir) Stat() (fs.FileInfo, error) {
	return &gitDirEntry{name: d.name}, nil
}

func (d *gitFSDir) ReadDir(count int) ([]fs.DirEntry, error) {
	n := len(d.entries) - d.offset
	if n == 0 {
		if count <= 0 {
			return nil, nil
		}
		return nil, io.EOF
	}
	if count > 0 && n > count {
		n = count
	}

	list := make([]fs.DirEntry, 0, n)
	for i := 0; i < n; i++ {
		name := d.entries[d.offset]
		list = append(list, &gitDirEntry{name: name})
		d.offset++
	}

	return list, nil
}

func (d *gitFSDir) Read(_ []byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.name, Err: errors.New("is a directory")}
}

func (d *gitFSDir) Close() error {
	return nil
}

type gitDirEntry struct {
	name string
}

func (e *gitDirEntry) Name() string               { return e.name }
func (e *gitDirEntry) Size() int64                { return 0 }
func (e *gitDirEntry) Mode() fs.FileMode          { return fs.ModeDir }
func (e *gitDirEntry) ModTime() time.Time         { return time.Time{} }
func (e *gitDirEntry) IsDir() bool                { return e.Mode().IsDir() }
func (e *gitDirEntry) Sys() any                   { return nil }
func (e *gitDirEntry) Type() fs.FileMode          { return fs.ModeDir }
func (e *gitDirEntry) Info() (fs.FileInfo, error) { return e, nil }

type gitFSFile struct {
	name string
	size int64
	io.ReadCloser
}

func (f *gitFSFile) Stat() (fs.FileInfo, error) {
	return &gitFileEntry{name: f.name, size: f.size}, nil
}

func (d *gitFSFile) ReadDir(count int) ([]fs.DirEntry, error) {
	return nil, &fs.PathError{Op: "read", Path: d.name, Err: errors.New("not a directory")}
}

type gitFileEntry struct {
	name string
	size int64
}

func (e *gitFileEntry) Name() string       { return e.name }
func (e *gitFileEntry) Size() int64        { return e.size }
func (e *gitFileEntry) Mode() fs.FileMode  { return fs.ModePerm }
func (e *gitFileEntry) ModTime() time.Time { return time.Time{} }
func (e *gitFileEntry) IsDir() bool        { return e.Mode().IsDir() }
func (e *gitFileEntry) Sys() any           { return nil }

func makeRevPath(rev, path string) string {
	return fmt.Sprintf("%s^:%s", rev, path)
}
