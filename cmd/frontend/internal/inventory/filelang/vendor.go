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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_363(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
