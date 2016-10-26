package ctags

import "path/filepath"

func isSupportedFile(mode, filename string) bool {
	ext := filepath.Ext(filename)
	switch mode {
	case "c":
		return ext == ".c" || ext == ".h"
	case "ruby":
		return ext == ".rb"
	default:
		return false
	}
}
