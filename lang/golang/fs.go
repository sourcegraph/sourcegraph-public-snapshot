package golang

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

func (h *Session) filePath(uri string) string {
	path := strings.TrimPrefix(strings.TrimPrefix(uri, "file://"), "/")
	return filepath.Join(h.init.RootPath, path)
}

func (h *Session) readFile(uri string) ([]byte, error) {
	path := h.filePath(uri)

	// TODO(sqs): sanitize paths, ensure that we can't break outside of h.init.RootPath
	contents, err := ioutil.ReadFile(path)
	return contents, err
}
