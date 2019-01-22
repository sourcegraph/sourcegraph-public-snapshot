// Command minversion prints a boolean stating whether users are running the minimum required Go version.
package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	version "github.com/mcuadros/go-version"
)

func main() {
	minimumVersion := "1.11.0"
	rawVersion := runtime.Version()
	versionNumber := strings.TrimPrefix(rawVersion, "go")
	minimumVersionMet := version.Compare(minimumVersion, versionNumber, "<=")
	if !minimumVersionMet {
		fmt.Printf("Go version %s or newer must be used; found: %s", minimumVersion, versionNumber)
		os.Exit(1) // minimum version not met means non-zero exit code
	}
}
