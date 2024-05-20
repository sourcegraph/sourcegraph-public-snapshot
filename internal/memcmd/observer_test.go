package memcmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func allocatingGoProgram(t testing.TB, allocationSizeBytes uint64) *exec.Cmd {
	t.Helper()

	const goTemplate = `
package main

import (
	"fmt"
	"time"
	"os"
)

func main() {
	var slice []byte

	if len(os.Args) > 0 { // Conditional that's always true to force the slice to be allocated on the heap
		slice = make([]byte, %d)
		for i := 0; i < len(slice); i++ {
			slice[i] = byte(i & 0xff)
		}
	}

	time.Sleep(500 * time.Millisecond)
	fmt.Println(len(slice)) // Don't optimize the slice away
}`

	goSource := fmt.Sprintf(goTemplate, allocationSizeBytes)

	goFile := filepath.Join(t.TempDir(), "main.go")
	err := os.WriteFile(goFile, []byte(goSource), 0644) // permissions: -rw-r--r--
	if err != nil {
		t.Fatalf("failed to write test program: %v", err)
	}

	const bashTemplate = `
#!/usr/bin/env bash
set -euxo pipefail
go run %s
echo "done" # force bash to fork
`

	bashCommand := fmt.Sprintf(bashTemplate, goFile)

	cmd := exec.CommandContext(context.Background(), "bash", "-c", bashCommand)
	return cmd
}
