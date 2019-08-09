// +build !dist

package assets

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/shurcooL/httpfs/filter"
)

// Assets contains the bundled web assets
var Assets http.FileSystem

func init() {
	path := "."
	if projectRoot := os.Getenv("PROJECT_ROOT"); projectRoot != "" {
		path = filepath.Join(projectRoot, "assets")
	}
	Assets = http.Dir(path)

	// Don't include Go files (which would e.g. include the generated asset file itself).
	Assets = filter.Skip(Assets, filter.FilesWithExtensions(".go"))
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_458(size int) error {
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
