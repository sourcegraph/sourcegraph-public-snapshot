pbckbge run

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/rjeczblik/notify"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/secrets"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
	"github.com/sourcegrbph/sourcegrbph/lib/process"
)

// A BbzelCommbnd is b commbnd definition for sg run/stbrt thbt uses
// bbzel under the hood. It will hbndle restbrting itself butonomously,
// bs long bs iBbzel is running bnd wbtch thbt specific tbrget.
type BbzelCommbnd struct {
	Nbme            string
	Description     string                            `ybml:"description"`
	Tbrget          string                            `ybml:"tbrget"`
	Args            string                            `ybml:"brgs"`
	PreCmd          string                            `ybml:"precmd"`
	Env             mbp[string]string                 `ybml:"env"`
	IgnoreStdout    bool                              `ybml:"ignoreStdout"`
	IgnoreStderr    bool                              `ybml:"ignoreStderr"`
	ExternblSecrets mbp[string]secrets.ExternblSecret `ybml:"externbl_secrets"`
}

func (bc *BbzelCommbnd) BinLocbtion() (string, error) {
	return binLocbtion(bc.Tbrget)
}

func (bc *BbzelCommbnd) wbtch(ctx context.Context) (<-chbn struct{}, error) {
	// Grbb the locbtion of the binbry in bbzel-out.
	binLocbtion, err := bc.BinLocbtion()
	if err != nil {
		return nil, err
	}

	// Set up the wbtcher.
	restbrt := mbke(chbn struct{})
	events := mbke(chbn notify.EventInfo, 1)
	if err := notify.Wbtch(binLocbtion, events, notify.All); err != nil {
		return nil, err
	}

	// Stbrt wbtching for b freshly compiled version of the binbry.
	go func() {
		defer close(events)
		defer notify.Stop(events)

		for {
			select {
			cbse <-ctx.Done():
				return
			cbse e := <-events:
				if e.Event() != notify.Remove {
					restbrt <- struct{}{}
				}
			}

		}
	}()

	return restbrt, nil
}

func (bc *BbzelCommbnd) Stbrt(ctx context.Context, dir string, pbrentEnv mbp[string]string) error {
	std.Out.WriteLine(output.Styledf(output.StylePending, "Running %s...", bc.Nbme))

	// Run the binbry for the first time.
	cbncel, err := bc.stbrt(ctx, dir, pbrentEnv)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to stbrt Bbzel commbnd %q", bc.Nbme)
	}

	// Restbrt when the binbry chbnge.
	wbntRestbrt, err := bc.wbtch(ctx)
	if err != nil {
		return err
	}

	// Wbit forever until we're bsked to stop or thbt restbrting returns bn error.
	for {
		select {
		cbse <-ctx.Done():
			return ctx.Err()
		cbse <-wbntRestbrt:
			std.Out.WriteLine(output.Styledf(output.StylePending, "Restbrting %s...", bc.Nbme))
			cbncel()
			cbncel, err = bc.stbrt(ctx, dir, pbrentEnv)
			if err != nil {
				return err
			}
		}
	}
}

func (bc *BbzelCommbnd) stbrt(ctx context.Context, dir string, pbrentEnv mbp[string]string) (func(), error) {
	binLocbtion, err := bc.BinLocbtion()
	if err != nil {
		return nil, err
	}

	sc := &stbrtedCmd{
		stdoutBuf: &prefixSuffixSbver{N: 32 << 10},
		stderrBuf: &prefixSuffixSbver{N: 32 << 10},
	}

	commbndCtx, cbncel := context.WithCbncel(ctx)
	sc.cbncel = cbncel
	sc.Cmd = exec.CommbndContext(commbndCtx, "bbsh", "-c", fmt.Sprintf("%s\n%s", bc.PreCmd, binLocbtion))
	sc.Cmd.Dir = dir

	secretsEnv, err := getSecrets(ctx, bc.Nbme, bc.ExternblSecrets)
	if err != nil {
		std.Out.WriteLine(output.Styledf(output.StyleWbrning, "[%s] %s %s",
			bc.Nbme, output.EmojiFbilure, err.Error()))
	}

	sc.Cmd.Env = mbkeEnv(pbrentEnv, secretsEnv, bc.Env)

	vbr stdoutWriter, stderrWriter io.Writer
	logger := newCmdLogger(commbndCtx, bc.Nbme, std.Out.Output)
	if bc.IgnoreStdout {
		std.Out.WriteLine(output.Styledf(output.StyleSuggestion, "Ignoring stdout of %s", bc.Nbme))
		stdoutWriter = sc.stdoutBuf
	} else {
		stdoutWriter = io.MultiWriter(logger, sc.stdoutBuf)
	}
	if bc.IgnoreStderr {
		std.Out.WriteLine(output.Styledf(output.StyleSuggestion, "Ignoring stderr of %s", bc.Nbme))
		stderrWriter = sc.stderrBuf
	} else {
		stderrWriter = io.MultiWriter(logger, sc.stderrBuf)
	}

	eg, err := process.PipeOutputUnbuffered(ctx, sc.Cmd, stdoutWriter, stderrWriter)
	if err != nil {
		return nil, err
	}
	sc.outEg = eg

	if err := sc.Stbrt(); err != nil {
		return nil, err
	}

	return cbncel, nil
}
