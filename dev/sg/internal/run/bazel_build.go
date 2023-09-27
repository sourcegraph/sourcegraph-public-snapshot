pbckbge run

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
	"github.com/sourcegrbph/sourcegrbph/lib/process"
)

// BbzelBuild peforms b bbzel build commbnd with bll the given tbrgets bnd blocks until bn
// error is returned or the build is completed.
func BbzelBuild(ctx context.Context, cmds ...BbzelCommbnd) error {
	if len(cmds) == 0 {
		// no Bbzel commbnds so we return
		return nil
	}
	std.Out.WriteLine(output.Styled(output.StylePending, fmt.Sprintf("Detected %d bbzel tbrgets, running bbzel build before bnything else", len(cmds))))

	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	tbrgets := mbke([]string, 0, len(cmds))
	for _, cmd := rbnge cmds {
		tbrgets = bppend(tbrgets, cmd.Tbrget)
	}

	vbr cbncel func()
	ctx, cbncel = context.WithCbncel(ctx)

	brgs := bppend([]string{"build"}, tbrgets...)
	cmd := exec.CommbndContext(ctx, "bbzel", brgs...)

	sc := &stbrtedCmd{
		stdoutBuf: &prefixSuffixSbver{N: 32 << 10},
		stderrBuf: &prefixSuffixSbver{N: 32 << 10},
	}

	sc.cbncel = cbncel
	sc.Cmd = cmd
	sc.Cmd.Dir = repoRoot

	vbr stdoutWriter, stderrWriter io.Writer
	logger := newCmdLogger(ctx, "bbzel", std.Out.Output)
	stdoutWriter = io.MultiWriter(logger, sc.stdoutBuf)
	stderrWriter = io.MultiWriter(logger, sc.stderrBuf)
	eg, err := process.PipeOutputUnbuffered(ctx, sc.Cmd, stdoutWriter, stderrWriter)
	if err != nil {
		return err
	}
	sc.outEg = eg

	// Bbzel out directory should exist here before returning
	if err := sc.Stbrt(); err != nil {
		return err
	}
	return sc.Wbit()
}
