package langserver

import (
	"os"
	"strings"
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
