package embedded

import (
	"embed"
	"os"
	"path"
)

// embeddedRootDir is the name of the root directory for the embeddedFS variable.
const embeddedRootDir string = "resources"

//go:embed resources/*
var embeddedFS embed.FS

// entrypointName is the path to the entry point relative to the embeddedRootDir.
var entrypointName = path.Join("bin", "jsii-runtime.js")

// ExtractRuntime extracts a copy of the embedded runtime library into
// the designated directory, and returns the fully qualified path to the entry
// point to be used when starting the child process.
func ExtractRuntime(into string) (entrypoint string, err error) {
	err = extractRuntime(into, embeddedRootDir)
	if err == nil {
		entrypoint = path.Join(into, entrypointName)
	}
	return
}

// extractRuntime copies the contents of embeddedFS at "from" to the provided
// "into" directory, recursively.
func extractRuntime(into string, from string) error {
	files, err := embeddedFS.ReadDir(from)
	if err != nil {
		return err
	}
	for _, file := range files {
		src := path.Join(from, file.Name())
		dest := path.Join(into, file.Name())
		if file.IsDir() {
			if err = os.Mkdir(dest, 0o700); err != nil {
				return err
			}
			if err = extractRuntime(dest, src); err != nil {
				return err
			}
		} else {
			data, err := embeddedFS.ReadFile(src)
			if err != nil {
				return err
			}
			if err = os.WriteFile(dest, data, 0o600); err != nil {
				return err
			}
		}
	}
	return nil
}
