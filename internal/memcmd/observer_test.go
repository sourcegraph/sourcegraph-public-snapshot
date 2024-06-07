package memcmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

var goBinary = "go"

func init() {
	if path := os.Getenv("GO_RLOCATIONPATH"); path != "" {
		var err error
		goBinary, err = runfiles.Rlocation(path)
		if err != nil {
			panic(err)
		}
	}
}

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
	err := os.WriteFile(goFile, []byte(goSource), 0o644) // permissions: -rw-r--r--
	if err != nil {
		t.Fatalf("failed to write test program: %v", err)
	}

	const bashTemplate = `
#!/usr/bin/env bash
set -euxo pipefail
%s run %s
echo "done" # force bash to fork
`

	bashCommand := fmt.Sprintf(bashTemplate, goBinary, goFile)

	gocacheDir := t.TempDir()

	cmd := exec.CommandContext(context.Background(), "bash", "-c", bashCommand)
	cmd.Env = append(cmd.Env, "GOCACHE="+gocacheDir)
	return cmd
}
