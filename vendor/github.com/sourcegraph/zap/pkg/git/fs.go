package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type OSFS struct{ Root string }

var _ FileSystem = OSFS{}

func (fs OSFS) absPath(name string) string {
	if strings.HasPrefix(name, "#") {
		panic(fmt.Sprintf("buffer path %q is not expected here, expecting a file path or relative path", name))
	}
	if strings.HasPrefix(name, "/") {
		name = strings.TrimPrefix(name, "/")
	}
	if fs.Root == "" {
		panic("no root")
	}
	return filepath.Join(fs.Root, name)
}

func (fs OSFS) ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(fs.absPath(name))
}

func (fs OSFS) WriteFile(name string, data []byte, mode os.FileMode) error {
	name = fs.absPath(name)
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(name, data, mode)
}

func (fs OSFS) Rename(oldpath, newpath string) error {
	oldpath = fs.absPath(oldpath)
	newpath = fs.absPath(newpath)
	if err := os.MkdirAll(filepath.Dir(newpath), 0700); err != nil {
		return err
	}
	return os.Rename(oldpath, newpath)
}

func (fs OSFS) Exists(name string) error {
	_, err := os.Stat(fs.absPath(name))
	return err
}

func (fs OSFS) Remove(name string) error {
	return os.Remove(fs.absPath(name))
}
