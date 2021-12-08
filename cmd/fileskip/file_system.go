package fileskip

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
	StatSize(name string) (int64, error)
}

type InMemoryFileSystem struct {
	Files map[string]string
}

func (g *InMemoryFileSystem) RootDir() string {
	return ""
}

func (g *InMemoryFileSystem) StatSize(name string) (int64, error) {
	data, err := g.ReadRelativeFilename(name)
	return int64(len(data)), err
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

func (g *GitFileSystem) StatSize(name string) (int64, error) {
	stat, err := os.Stat(path.Join(g.Dir, name))
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
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
	Path   string
	Reader *zip.Reader
}

func NewZipFileSystem(path string) (*ZipFileSystem, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		panic(err)
		return nil, err
	}
	return &ZipFileSystem{Reader: reader}, nil
}

func (g *ZipFileSystem) Close() error {
	return nil
}

func (g *ZipFileSystem) RootDir() string {
	return ""
}

func (g *ZipFileSystem) StatSize(name string) (int64, error) {
	open, err := g.Reader.Open(name)
	if err != nil {
		return 0, err
	}
	stat, err := open.Stat()
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
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
		return []byte{}, nil
	}

	data := make([]byte, stat.Size())
	_, err = io.ReadFull(open, data)
	return data, err
}

func (g *ZipFileSystem) ListRelativeFilenames() ([]string, error) {
	var names []string
	for _, file := range g.Reader.File {
		if strings.HasSuffix(file.Name, "/") {
			continue
		}
		names = append(names, file.Name)
	}
	return names, nil
}

func (g *ZipFileSystem) GobEncode() ([]byte, error) {
	return []byte(g.Path), nil
}

func (g *ZipFileSystem) GobDecode(b []byte) error {
	fs, err := NewZipFileSystem(string(b))
	if err != nil {
		return err
	}
	g.Reader = fs.Reader
	return nil
}
