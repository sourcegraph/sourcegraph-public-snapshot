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

	// First consult overlay.
	if contents, found := h.readOverlayFile(path); found {
		return contents, nil
	}

	// TODO(sqs): sanitize paths, ensure that we can't break outside of h.init.RootPath
	contents, err := ioutil.ReadFile(path)
	return contents, err
}

func (h *Session) addOverlayFile(uri string, contents []byte) {
	if h.overlayFiles == nil {
		h.overlayFiles = map[string][]byte{}
	}
	h.overlayFiles[h.filePath(uri)] = contents
}

func (h *Session) removeOverlayFile(uri string) {
	delete(h.overlayFiles, h.filePath(uri))
}

func (h *Session) readOverlayFile(uri string) (contents []byte, found bool) {
	contents, found = h.overlayFiles[h.filePath(uri)]
	return
}
