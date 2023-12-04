package workspace

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/cmdlogger"
)

const loopDevPath = "/var/lib/firecracker/loop-devices"
const mountpointsPath = "/var/lib/firecracker/mountpoints"

// MakeLoopFile defaults to makeTemporaryLoopFile and can be replaced for testing
// with deterministic paths.
var MakeLoopFile = makeTemporaryLoopFile

func makeTemporaryLoopFile(prefix string) (*os.File, error) {
	if err := os.MkdirAll(loopDevPath, os.ModePerm); err != nil {
		return nil, err
	}
	return os.CreateTemp(loopDevPath, prefix+"-*")
}

// MakeMountDirectory defaults to makeTemporaryMountDirectory and can be replaced for testing
// with deterministic workspace/scripts directories.
var MakeMountDirectory = MakeTemporaryMountDirectory

func MakeTemporaryMountDirectory(prefix string) (string, error) {
	if err := os.MkdirAll(mountpointsPath, os.ModePerm); err != nil {
		return "", err
	}

	return os.MkdirTemp(mountpointsPath, prefix+"-*")
}

// runs the given command with args and logs the invocation and output to the provided log entry handle.
func commandLogger(ctx context.Context, cmdRunner util.CmdRunner, handle cmdlogger.LogEntry, command string, args ...string) (string, error) {
	fmt.Fprintf(handle, "$ %s %s\n", command, strings.Join(args, " "))
	out, err := cmdRunner.CombinedOutput(ctx, command, args...)
	if len(out) == 0 {
		fmt.Fprint(handle, "stderr: <no output>\n")
	} else {
		fmt.Fprintf(handle, "stderr: %s\n", strings.ReplaceAll(strings.TrimSpace(string(out)), "\n", "\nstderr: "))
	}

	return string(out), err
}
