package pathutil

import (
	"fmt"
	"strings"
)

func panicIfFileOrBufferPath(path string) {
	if strings.HasPrefix(path, "#") || strings.HasPrefix(path, "/") {
		panic(fmt.Sprintf("unexpected file or buffer path %q", path))
	}
}

func panicIfNotFileOrBufferPath(path string) {
	if !strings.HasPrefix(path, "#") && !strings.HasPrefix(path, "/") {
		panic(fmt.Sprintf("path %q is not a buffer or file path", path))
	}
}

func ToBufferPath(path string) string {
	if strings.HasPrefix(path, "/") {
		panic(fmt.Sprintf("path %q is a file path, wanted a non-buffer, non-file path", path))
	}
	return "#" + path
}

func IsBufferPath(path string) bool {
	panicIfNotFileOrBufferPath(path)
	return strings.HasPrefix(path, "#")
}

func IsFilePath(path string) bool {
	panicIfNotFileOrBufferPath(path)
	return strings.HasPrefix(path, "/")
}

func StripBufferPath(path string) string {
	if !strings.HasPrefix(path, "#") {
		panic(fmt.Sprintf("expected path %q to have '#' prefix", path))
	}
	return strings.TrimPrefix(path, "#")
}

func StripFilePath(path string) string {
	if !strings.HasPrefix(path, "/") {
		panic(fmt.Sprintf("expected path %q to have '/' prefix", path))
	}
	return strings.TrimPrefix(path, "/")
}

func StripFileOrBufferPath(path string) string {
	panicIfNotFileOrBufferPath(path)
	return path[1:]
}
