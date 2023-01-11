package workspace

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
)

// MakeTempFile defaults to makeTemporaryFile and can be replaced for testing
// with determinstic workspace/scripts directories.
var MakeTempFile = makeTemporaryFile

func makeTemporaryFile(prefix string) (*os.File, error) {
	if tempdir := os.Getenv("TMPDIR"); tempdir != "" {
		if err := os.MkdirAll(tempdir, os.ModePerm); err != nil {
			return nil, err
		}
		return os.CreateTemp(tempdir, prefix+"-*")
	}

	return os.CreateTemp("", prefix+"-*")
}

// MakeTempDirectory defaults to makeTemporaryDirectory and can be replaced for testing
// with determinstic workspace/scripts directories.
var MakeTempDirectory = MakeTemporaryDirectory

func MakeTemporaryDirectory(prefix string) (string, error) {
	if tempdir := os.Getenv("TMPDIR"); tempdir != "" {
		if err := os.MkdirAll(tempdir, os.ModePerm); err != nil {
			return "", err
		}
		return os.MkdirTemp(tempdir, prefix+"-*")
	}

	return os.MkdirTemp("", prefix+"-*")
}

// runs the given command with args and logs the invocation and output to the provided log entry handle.
func commandLogger(ctx context.Context, handle command.LogEntry, command string, args ...string) (string, error) {
	fmt.Fprintf(handle, "$ %s %s\n", command, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, command, args...)
	out, err := cmd.CombinedOutput()
	if len(out) == 0 {
		fmt.Fprint(handle, "stderr: <no output>\n")
	} else {
		fmt.Fprintf(handle, "stderr: %s\n", strings.ReplaceAll(strings.TrimSpace(string(out)), "\n", "\nstderr: "))
	}

	return string(out), err
}
