package run

import (
	"context"
	"io"
	"os/exec"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/process"
)

type IBazel struct {
	pwd     string
	targets []string
	cancel  func()
}

// newIBazel returns a runner to interact with ibazel.
func newIBazel(pwd string, targets ...string) *IBazel {
	return &IBazel{
		pwd:     pwd,
		targets: targets,
	}
}

func (ib *IBazel) Start(ctx context.Context, dir string) error {
	args := append([]string{"build"}, ib.targets...)
	ctx, ib.cancel = context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, "ibazel", args...)

	sc := &startedCmd{
		stdoutBuf: &prefixSuffixSaver{N: 32 << 10},
		stderrBuf: &prefixSuffixSaver{N: 32 << 10},
	}

	sc.cancel = ib.cancel
	sc.Cmd = cmd
	sc.Cmd.Dir = dir

	var stdoutWriter, stderrWriter io.Writer
	logger := newCmdLogger(ctx, "iBazel", std.Out.Output)
	stdoutWriter = io.MultiWriter(logger, sc.stdoutBuf)
	stderrWriter = io.MultiWriter(logger, sc.stderrBuf)
	eg, err := process.PipeOutputUnbuffered(ctx, sc.Cmd, stdoutWriter, stderrWriter)
	if err != nil {
		return err
	}
	sc.outEg = eg

	// Bazel out directory should exist here before returning
	return sc.Start()
}

func (ib *IBazel) Stop() error {
	ib.cancel()
	return nil
}
