package ctags

import "path/filepath"

func isSupportedFile(filenameORURI string) bool {
	if ext := filepath.Ext(filenameORURI); ext == ".c" || ext == ".h" || ext == ".rb" {
		return true
	}
	return false
}
