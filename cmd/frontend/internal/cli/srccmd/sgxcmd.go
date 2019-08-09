// Package srccmd contains the src command name. It's split from package cli to avoid import cycles.
package srccmd

import (
	"log"
	"path/filepath"

	"github.com/kardianos/osext"
)

// Path is the path at which the binary can be found for execution purposes.
// There is no guarantee that the binary will be on the system's path, so you
// must always use this path instead for executing the command.
// Path uses UNIX-style file separators to ensure it suitable both for running
// from system environment and Makefiles/shell scripts
var Path string

func init() {
	// Grab the absolute path to the executable. Do not use os.Args[0] for
	// reasons outlined in osext README.
	var err error
	Path, err = osext.Executable()
	if err != nil {
		log.Fatal(err)
	}
	// Enforce Unix-style path, because this variable may be used in Makefiles
	Path = filepath.ToSlash(Path)

	// Detect if we are a test binary by looking at our extension. If we are
	// then we do not specify the absolute path to the binary, instead we leave
	// it simply as 'src' because several tests will attempt to perform self
	// invocation and will instead attempt to run the test binary itself, which
	// will fail because the test binary expects CLI flags unrelated to ours.
	// It is for this reason that `src` must be on the system path during
	// testing.
	if filepath.Ext(Path) == ".test" {
		Path = "src"
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_336(size int) error {
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
