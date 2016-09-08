package ctags

import (
	"strings"
)

func (h *Handler) filePath(uri string) string {
	return strings.TrimPrefix(uri, "file://")
}
