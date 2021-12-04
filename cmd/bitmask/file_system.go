package bitmask

import (
	"archive/zip"
	"bytes"
	"github.com/cockroachdb/errors"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

type FileSystem interface {
	ListRelativeFilenames() ([]string, error)
	ReadRelativeFilename(name string) ([]byte, error)
	RootDir() string
}

type InMemoryFileSystem struct {
	Files map[string]string
}

func (g *InMemoryFileSystem) RootDir() string {
	return ""
}

func (g *InMemoryFileSystem) ReadRelativeFilename(name string) ([]byte, error) {
	value, found := g.Files[name]
	if !found {
		return nil, errors.Errorf("no such file: %v", name)
	}
	return []byte(value), nil
}

func (g *InMemoryFileSystem) ListRelativeFilenames() ([]string, error) {
	var result []string
	for file := range g.Files {
		result = append(result, file)
	}
	return result, nil
}

type GitFileSystem struct {
	Dir string
}

func (g *GitFileSystem) RootDir() string {
	return g.Dir
}

func (g *GitFileSystem) ReadRelativeFilename(name string) ([]byte, error) {
	abspath := path.Join(g.Dir, name)
	return os.ReadFile(abspath)
}

func (g *GitFileSystem) ListRelativeFilenames() ([]string, error) {
	var branch bytes.Buffer
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchCmd.Dir = g.Dir
	branchCmd.Stdout = &branch
	err := branchCmd.Run()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to infer the default branch")
	}
	cmd := exec.Command(
		"git",
		"ls-files",
		"-z",
		"--with-tree",
		strings.Trim(branch.String(), "\n"),
	)
	cmd.Dir = g.Dir
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()

	if err != nil {
		return nil, err
	}
	stdout := string(out.Bytes())
	NUL := string([]byte{0})
	return strings.Split(stdout, NUL), nil
}

type ZipFileSystem struct {
	Reader *zip.Reader
	Closer io.Closer
}

func NewZipFileSystem(path string) (*ZipFileSystem, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	reader, err := zip.NewReader(file, stat.Size())
	if err != nil {
		return nil, err
	}
	return &ZipFileSystem{Reader: reader, Closer: file}, nil
}

func (g *ZipFileSystem) Close() error {
	return g.Closer.Close()
}

func (g *ZipFileSystem) RootDir() string {
	return ""
}

func (g *ZipFileSystem) ReadRelativeFilename(name string) ([]byte, error) {
	open, err := g.Reader.Open(name)
	if err != nil {
		return nil, err
	}
	stat, err := open.Stat()
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return nil, errors.Errorf("can't read directory %v", name)
	}
	b := make([]byte, stat.Size())
	read, err := open.Read(b)
	if err != nil {
		return nil, err
	}
	if read != len(b) {
		return nil, errors.Errorf("read %v, expected %v", read, len(b))
	}
	return b, nil
}

func (g *ZipFileSystem) ListRelativeFilenames() ([]string, error) {
	var names []string
	for _, file := range g.Reader.File {
		names = append(names, file.Name)
	}
	return names, nil
}
