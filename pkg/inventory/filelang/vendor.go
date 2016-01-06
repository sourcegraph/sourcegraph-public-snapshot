package filelang

import (
	"os"
	"regexp"
	"strings"
)

func init() {
	vendorPatterns = append(vendorPatterns,
		regexp.MustCompile(`^\.git/`),
		regexp.MustCompile(`^\.hg/`),
		regexp.MustCompile(`^\.srclib-cache/`),
		regexp.MustCompile(`^\.srclib-store/`),
	)
}

// IsVendored returns whether a path (and everything underneath it) is
// vendored.
func IsVendored(path string, isDir bool) bool {
	path = strings.TrimPrefix(path, string(os.PathSeparator))
	if isDir {
		path += "/"
	}
	for _, re := range vendorPatterns {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}
