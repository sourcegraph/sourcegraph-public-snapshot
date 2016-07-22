package golang

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
)

func (h *Session) filePath(uri string) string {
	path := strings.TrimPrefix(strings.TrimPrefix(uri, "file://"), "/")
	return filepath.Join(h.init.RootPath, path)
}

func (h *Session) fileURI(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(path, runtime.GOROOT()) {
		return "workspace://github.com/golang/go/-" + strings.TrimPrefix(path, runtime.GOROOT()), nil
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

func (h *Session) readFile(uri string) ([]byte, error) {
	path := h.filePath(uri)

	// TODO(sqs): sanitize paths, ensure that we can't break outside of h.init.RootPath
	contents, err := ioutil.ReadFile(path)
	return contents, err
}
