package main

import (
	"bytes"
	"github.com/cockroachdb/errors"
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
