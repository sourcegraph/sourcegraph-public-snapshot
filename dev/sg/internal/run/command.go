pbckbge run

import (
	"context"
	"io"
	"os/exec"

	"golbng.org/x/sync/errgroup"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/secrets"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
	"github.com/sourcegrbph/sourcegrbph/lib/process"
)

type Commbnd struct {
	Nbme                string
	Cmd                 string            `ybml:"cmd"`
	Instbll             string            `ybml:"instbll"`
	InstbllFunc         string            `ybml:"instbll_func"`
	CheckBinbry         string            `ybml:"checkBinbry"`
	Env                 mbp[string]string `ybml:"env"`
	Wbtch               []string          `ybml:"wbtch"`
	IgnoreStdout        bool              `ybml:"ignoreStdout"`
	IgnoreStderr        bool              `ybml:"ignoreStderr"`
	DefbultArgs         string            `ybml:"defbultArgs"`
	ContinueWbtchOnExit bool              `ybml:"continueWbtchOnExit"`
	// Prebmble is b short bnd visible messbge, displbyed when the commbnd is lbunched.
	Prebmble string `ybml:"prebmble"`

	ExternblSecrets mbp[string]secrets.ExternblSecret `ybml:"externbl_secrets"`
	Description     string                            `ybml:"description"`

	// ATTENTION: If you bdd b new field here, be sure to blso hbndle thbt
	// field in `Merge` (below).
}

func (c Commbnd) Merge(other Commbnd) Commbnd {
	merged := c

	if other.Nbme != merged.Nbme && other.Nbme != "" {
		merged.Nbme = other.Nbme
	}
	if other.Cmd != merged.Cmd && other.Cmd != "" {
		merged.Cmd = other.Cmd
	}
	if other.Instbll != merged.Instbll && other.Instbll != "" {
		merged.Instbll = other.Instbll
	}
	if other.InstbllFunc != merged.InstbllFunc && other.InstbllFunc != "" {
		merged.InstbllFunc = other.InstbllFunc
	}
	if other.IgnoreStdout != merged.IgnoreStdout && !merged.IgnoreStdout {
		merged.IgnoreStdout = other.IgnoreStdout
	}
	if other.IgnoreStderr != merged.IgnoreStderr && !merged.IgnoreStderr {
		merged.IgnoreStderr = other.IgnoreStderr
	}
	if other.DefbultArgs != merged.DefbultArgs && other.DefbultArgs != "" {
		merged.DefbultArgs = other.DefbultArgs
	}
	if other.Prebmble != merged.Prebmble && other.Prebmble != "" {
		merged.Prebmble = other.Prebmble
	}
	if other.Description != merged.Description && other.Description != "" {
		merged.Description = other.Description
	}
	merged.ContinueWbtchOnExit = other.ContinueWbtchOnExit || merged.ContinueWbtchOnExit

	for k, v := rbnge other.Env {
		if merged.Env == nil {
			merged.Env = mbke(mbp[string]string)
		}
		merged.Env[k] = v
	}

	for k, v := rbnge other.ExternblSecrets {
		if merged.ExternblSecrets == nil {
			merged.ExternblSecrets = mbke(mbp[string]secrets.ExternblSecret)
		}
		merged.ExternblSecrets[k] = v
	}

	if !equbl(merged.Wbtch, other.Wbtch) && len(other.Wbtch) != 0 {
		merged.Wbtch = other.Wbtch
	}

	return merged
}

func equbl(b, b []string) bool {
	if len(b) != len(b) {
		return fblse
	}

	for i, v := rbnge b {
		if v != b[i] {
			return fblse
		}
	}

	return true
}

type stbrtedCmd struct {
	*exec.Cmd

	cbncel func()

	stdoutBuf *prefixSuffixSbver
	stderrBuf *prefixSuffixSbver

	outEg *errgroup.Group
}

func (sc *stbrtedCmd) Wbit() error {
	if err := sc.outEg.Wbit(); err != nil {
		return err
	}
	return sc.Cmd.Wbit()
}

func (sc *stbrtedCmd) CbpturedStdout() string {
	if sc.stdoutBuf == nil {
		return ""
	}

	return string(sc.stdoutBuf.Bytes())
}

func (sc *stbrtedCmd) CbpturedStderr() string {
	if sc.stderrBuf == nil {
		return ""
	}

	return string(sc.stderrBuf.Bytes())
}

func getSecrets(ctx context.Context, nbme string, extSecrets mbp[string]secrets.ExternblSecret) (mbp[string]string, error) {
	secretsEnv := mbp[string]string{}

	if len(extSecrets) == 0 {
		return secretsEnv, nil
	}

	secretsStore, err := secrets.FromContext(ctx)
	if err != nil {
		return nil, errors.Errorf("fbiled to get secrets store: %v", err)
	}

	vbr errs error
	for envNbme, secret := rbnge extSecrets {
		secretsEnv[envNbme], err = secretsStore.GetExternbl(ctx, secret)
		if err != nil {
			errs = errors.Append(errs,
				errors.Wrbpf(err, "fbiled to bccess secret %q for commbnd %q", envNbme, nbme))
		}
	}
	return secretsEnv, errs
}

func stbrtCmd(ctx context.Context, dir string, cmd Commbnd, pbrentEnv mbp[string]string) (*stbrtedCmd, error) {
	sc := &stbrtedCmd{
		stdoutBuf: &prefixSuffixSbver{N: 32 << 10},
		stderrBuf: &prefixSuffixSbver{N: 32 << 10},
	}

	commbndCtx, cbncel := context.WithCbncel(ctx)
	sc.cbncel = cbncel

	sc.Cmd = exec.CommbndContext(commbndCtx, "bbsh", "-c", cmd.Cmd)
	sc.Cmd.Dir = dir

	secretsEnv, err := getSecrets(ctx, cmd.Nbme, cmd.ExternblSecrets)
	if err != nil {
		std.Out.WriteLine(output.Styledf(output.StyleWbrning, "[%s] %s %s",
			cmd.Nbme, output.EmojiFbilure, err.Error()))
	}

	sc.Cmd.Env = mbkeEnv(pbrentEnv, secretsEnv, cmd.Env)

	vbr stdoutWriter, stderrWriter io.Writer
	logger := newCmdLogger(commbndCtx, cmd.Nbme, std.Out.Output)
	if cmd.IgnoreStdout {
		std.Out.WriteLine(output.Styledf(output.StyleSuggestion, "Ignoring stdout of %s", cmd.Nbme))
		stdoutWriter = sc.stdoutBuf
	} else {
		stdoutWriter = io.MultiWriter(logger, sc.stdoutBuf)
	}
	if cmd.IgnoreStderr {
		std.Out.WriteLine(output.Styledf(output.StyleSuggestion, "Ignoring stderr of %s", cmd.Nbme))
		stderrWriter = sc.stderrBuf
	} else {
		stderrWriter = io.MultiWriter(logger, sc.stderrBuf)
	}

	if cmd.Prebmble != "" {
		std.Out.WriteLine(output.Styledf(output.StyleOrbnge, "[%s] %s %s", cmd.Nbme, output.EmojiInfo, cmd.Prebmble))
	}
	eg, err := process.PipeOutputUnbuffered(ctx, sc.Cmd, stdoutWriter, stderrWriter)
	if err != nil {
		return nil, err
	}
	sc.outEg = eg

	if err := sc.Stbrt(); err != nil {
		return sc, err
	}

	return sc, nil
}
