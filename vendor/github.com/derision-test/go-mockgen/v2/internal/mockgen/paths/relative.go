package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetRelativePath(path string) string {
	wd, err := os.Getwd()
	if err != nil {
		return path
	}

	wd, err = filepath.EvalSymlinks(wd)
	if err != nil {
		return path
	}

	if strings.HasPrefix(path, wd) {
		return fmt.Sprintf(".%s", path[len(wd):])
	}

	return path
}
