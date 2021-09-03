//go:build debug
// +build debug

package batches

import (
	"log"
	"os"
	"path/filepath"
)

// In builds with the debug flag (i.e. `go build -tags debug -o src ./cmd/src`)
// init() sets up the default logger to log to a file in ~/.sourcegraph.
func init() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("getting user home directory: %s", err)
	}

	fullPath := filepath.Join(homedir, ".sourcegraph", "src-cli.debug.log")

	f, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("setting debug log file failed: %s", err)
	}

	log.SetOutput(f)
}
