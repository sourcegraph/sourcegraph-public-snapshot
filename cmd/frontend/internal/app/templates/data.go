// +build !dist

package templates

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/shurcooL/httpfs/filter"
	"golang.org/x/tools/go/packages"
)

func importPathToDir(importPath string) string {
	pkgs, err := packages.Load(&packages.Config{Mode: packages.LoadFiles}, importPath)
	if err != nil || len(pkgs) == 0 || len(pkgs[0].GoFiles) == 0 {
		log.Fatal("Failed to find templates directory: ", err)
	}
	return filepath.Dir(pkgs[0].GoFiles[0])
}

// Data is a virtual filesystem that contains template data used by Sourcegraph app.
var Data = filter.Skip(
	http.Dir(importPathToDir("github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/templates")),
	filter.FilesWithExtensions(".go"),
)

// random will create a file of size bytes (rounded up to next 1024 size)
func random_283(size int) error {
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
