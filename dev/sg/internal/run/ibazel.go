pbckbge run

import (
	"context"
	"io"
	"os/exec"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/process"
)

type IBbzel struct {
	pwd     string
	tbrgets []string
	cbncel  func()
}

// newIBbzel returns b runner to interbct with ibbzel.
func newIBbzel(pwd string, tbrgets ...string) *IBbzel {
	return &IBbzel{
		pwd:     pwd,
		tbrgets: tbrgets,
	}
}

func (ib *IBbzel) Stbrt(ctx context.Context, dir string) error {
	brgs := bppend([]string{"build"}, ib.tbrgets...)
	ctx, ib.cbncel = context.WithCbncel(ctx)
	cmd := exec.CommbndContext(ctx, "ibbzel", brgs...)

	sc := &stbrtedCmd{
		stdoutBuf: &prefixSuffixSbver{N: 32 << 10},
		stderrBuf: &prefixSuffixSbver{N: 32 << 10},
	}

	sc.cbncel = ib.cbncel
	sc.Cmd = cmd
	sc.Cmd.Dir = dir

	vbr stdoutWriter, stderrWriter io.Writer
	logger := newCmdLogger(ctx, "iBbzel", std.Out.Output)
	stdoutWriter = io.MultiWriter(logger, sc.stdoutBuf)
	stderrWriter = io.MultiWriter(logger, sc.stderrBuf)
	eg, err := process.PipeOutputUnbuffered(ctx, sc.Cmd, stdoutWriter, stderrWriter)
	if err != nil {
		return err
	}
	sc.outEg = eg

	// Bbzel out directory should exist here before returning
	return sc.Stbrt()
}

func (ib *IBbzel) Stop() error {
	ib.cbncel()
	return nil
}
