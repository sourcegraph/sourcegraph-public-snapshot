package golang

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// TODO(keegancsmith) Find a more reliable way to get this commit
const stdlibVersion = "0d818588685976407c81c60d2fda289361cbc8ec" // go1.7

func (h *Handler) filePath(uri string) string {
	path := strings.TrimPrefix(uri, "file://")
	if strings.HasPrefix(path, "/gopath/") {
		path = strings.TrimPrefix(path, "/")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(h.init.RootPath, path)
	}
	return path
}

func (h *Handler) fileURI(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(path, runtime.GOROOT()) {
		return "stdlib://" + stdlibVersion + strings.TrimPrefix(path, runtime.GOROOT()), nil
	}
	root, err := filepath.Abs(h.init.RootPath)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(path, root) {
		return "", fmt.Errorf("%s is not a descendent of %s", path, root)
	}
	f, err := filepath.Rel(root, path)
	return "file:///" + f, err
}

func (h *Handler) readFile(uri string) ([]byte, error) {
	path := h.filePath(uri)

	// TODO(sqs): sanitize paths, ensure that we can't break outside of h.init.RootPath
	contents, err := ioutil.ReadFile(path)
	return contents, err
}

func (h *Handler) goEnv() []string {
	env := []string{"GOPATH=" + h.filePath("gopath")}
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "GOPATH=") {
			env = append(env, e)
		}
	}
	return env
}
