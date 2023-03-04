package run

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/sourcegraph/sourcegraph/lib/process"
)

// BazelBuild peforms a bazel build command with all the given targets and blocks until an
// error is returned or the build is completed.
func BazelBuild(ctx context.Context, cmds ...BazelCommand) error {
	if len(cmds) == 0 {
		// no Bazel commands so we return
		return nil
	}
	std.Out.WriteLine(output.Styled(output.StylePending, fmt.Sprintf("Detected %d bazel targets, running bazel build before anything else", len(cmds))))

	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	targets := make([]string, 0, len(cmds))
	for _, cmd := range cmds {
		targets = append(targets, cmd.Target)
	}

	var cancel func()
	ctx, cancel = context.WithCancel(ctx)

	args := append([]string{"build"}, targets...)
	cmd := exec.CommandContext(ctx, "bazel", args...)

	sc := &startedCmd{
		stdoutBuf: &prefixSuffixSaver{N: 32 << 10},
		stderrBuf: &prefixSuffixSaver{N: 32 << 10},
	}

	sc.cancel = cancel
	sc.Cmd = cmd
	sc.Cmd.Dir = repoRoot

	var stdoutWriter, stderrWriter io.Writer
	logger := newCmdLogger(ctx, "bazel", std.Out.Output)
	stdoutWriter = io.MultiWriter(logger, sc.stdoutBuf)
	stderrWriter = io.MultiWriter(logger, sc.stderrBuf)
	eg, err := process.PipeOutputUnbuffered(ctx, sc.Cmd, stdoutWriter, stderrWriter)
	if err != nil {
		return err
	}
	sc.outEg = eg

	// Bazel out directory should exist here before returning
	if err := sc.Start(); err != nil {
		return err
	}
	return sc.Wait()
}
