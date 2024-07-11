//go:build linux

package memcmd_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/memcmd"
)

func Example() {
	const template = `
#!/usr/bin/env bash
set -euo pipefail

</dev/zero head -c $((1024**2*50)) | tail
sleep 1
`

	tempDir, err := os.MkdirTemp("", "foo")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	p := filepath.Join(tempDir, "/script.sh")
	err = os.WriteFile(p, []byte(template), 0755)
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("bash", "-c", p) // 50MB
	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	observer, err := memcmd.NewLinuxObserver(context.Background(), cmd, 1*time.Millisecond)
	if err != nil {
		panic(err)
	}

	observer.Start()
	defer observer.Stop()

	err = cmd.Wait()
	if err != nil {
		panic(err)
	}

	memoryUsage, err := observer.MaxMemoryUsage()
	if err != nil {
		panic(err)
	}

	fmt.Println((0 < memoryUsage && memoryUsage < 100*1024*1024)) // Output should be between 0 and 100MB

	// Output:
	// true
}
