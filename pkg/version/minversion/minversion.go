// Ensure that users are running Go 1.11 or newer

// Package minversion prints a boolean stating whether users are running the minimum required Go version.
package main

import (
	"fmt"
	"runtime"
	"strings"

	version "github.com/mcuadros/go-version"
)

func main() {
	rawVersion := runtime.Version()
	versionNumber := strings.TrimPrefix(rawVersion, "go")
	fmt.Println(version.Compare("1.11.0", versionNumber, "<="))
}
