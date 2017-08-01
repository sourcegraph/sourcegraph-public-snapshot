package langserver

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

func PathHasPrefix(s, prefix string) bool {
	var prefixSlash string
	if prefix != "" && !strings.HasSuffix(prefix, string(os.PathSeparator)) {
		prefixSlash = prefix + string(os.PathSeparator)
	}
	return s == prefix || strings.HasPrefix(s, prefixSlash)
}

func PathTrimPrefix(s, prefix string) string {
	if s == prefix {
		return ""
	}
	if !strings.HasSuffix(prefix, string(os.PathSeparator)) {
		prefix += string(os.PathSeparator)
	}
	return strings.TrimPrefix(s, prefix)
}

func pathEqual(a, b string) bool {
	return PathTrimPrefix(a, b) == ""
}

// IsVendorDir tells if the specified directory is a vendor directory.
func IsVendorDir(dir string) bool {
	return strings.HasPrefix(dir, "vendor/") || strings.Contains(dir, "/vendor/")
}

// isFileURI tells if s denotes an absolute file URI.
func isFileURI(s lsp.DocumentURI) bool {
	return strings.HasPrefix(string(s), "file:///")
}

// pathToURI converts given absolute path to file URI
func pathToURI(path string) lsp.DocumentURI {
	return lsp.DocumentURI("file://" + path)
}

// uriToFilePath converts given absolute file URI to path. It panics if
// uri does not begin with "file:///".
func uriToFilePath(uri lsp.DocumentURI) string {
	if !isFileURI(uri) {
		panic("not an absolute file URI: " + uri)
	}
	return strings.TrimPrefix(string(uri), "file://")
}

// panicf takes the return value of recover() and outputs data to the log with
// the stack trace appended. Arguments are handled in the manner of
// fmt.Printf. Arguments should format to a string which identifies what the
// panic code was doing. Returns a non-nil error if it recovered from a panic.
func panicf(r interface{}, format string, v ...interface{}) error {
	if r != nil {
		// Same as net/http
		const size = 64 << 10
		buf := make([]byte, size)
		buf = buf[:runtime.Stack(buf, false)]
		id := fmt.Sprintf(format, v...)
		log.Printf("panic serving %s: %v\n%s", id, r, string(buf))
		return fmt.Errorf("unexpected panic: %v", r)
	}
	return nil
}
