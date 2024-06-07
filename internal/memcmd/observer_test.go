package memcmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	binaryPath := filepath.Join(t.TempDir(), "main")

	const bashTemplateGoBuild = `
#!/usr/bin/env bash
set -euxo pipefail

%s build -o %s %s
`

	ctx := context.Background()

	args := []string{
		"--login", // -l: login shell (so that we know that the PATH is set correctly for asdf if needed)
		"-c",
		fmt.Sprintf(bashTemplateGoBuild, goBinary, binaryPath, goFile),
	}

	goBuildCmd := exec.CommandContext(ctx, "bash", args...)
	goBuildCmd.Env = append(goBuildCmd.Env, fmt.Sprintf("GOCACHE=%s", t.TempDir()))

	{
		// Ensure that the HOME environment variable is set. This is required for
		// asdf to work correctly.
		hasHome := false
		for _, env := range goBuildCmd.Env {
			if strings.HasPrefix(env, "HOME=") {
				hasHome = true
				break
			}
		}

		if !hasHome {
			if home, err := os.UserHomeDir(); err == nil {
				goBuildCmd.Env = append(goBuildCmd.Env, fmt.Sprintf("HOME=%s", home))
			}
		}
	}

	_, err = goBuildCmd.Output()
	if err != nil {
		t.Fatalf("failed to compile test program: %v", err)
	}

	const bashTemplateRunCmd = `
#!/usr/bin/env bash
set -euxo pipefail

%s
echo "done" # force bash to fork
`
	return exec.Command("bash", "-c", fmt.Sprintf(bashTemplateRunCmd, binaryPath))
}
