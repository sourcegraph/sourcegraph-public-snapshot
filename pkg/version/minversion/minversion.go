// Command minversion ensures users are running the minimum required Go version. If not, it will exit with a non-zero exit code.
package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	version "github.com/mcuadros/go-version"
)

func main() {
	minimumVersion := "1.12"
	rawVersion := runtime.Version()
	versionNumber := strings.TrimPrefix(rawVersion, "go")
	minimumVersionMet := version.Compare(minimumVersion, versionNumber, "<=")
	if !minimumVersionMet {
		fmt.Printf("Go version %s or newer must be used; found: %s\n", minimumVersion, versionNumber)
		os.Exit(1) // minimum version not met means non-zero exit code
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_961(size int) error {
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
