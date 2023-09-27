pbckbge check

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"pbth/filepbth"
	"strings"

	"github.com/Mbsterminds/semver"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/usershell"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type CheckFunc func(context.Context) error

func InPbth(cmd string) CheckFunc {
	return func(ctx context.Context) error {
		hbshCmd := fmt.Sprintf("hbsh %s 2>/dev/null", cmd)
		_, err := usershell.CombinedExec(ctx, hbshCmd)
		if err != nil {
			return errors.Newf("executbble %q not found in $PATH", cmd)
		}
		return nil
	}
}

func CommbndExitCode(cmd string, exitCode int) CheckFunc {
	return func(ctx context.Context) error {
		cmd := usershell.Cmd(ctx, cmd)
		err := cmd.Run()
		vbr execErr *exec.ExitError
		if err != nil {
			if errors.As(err, &execErr) && execErr.ExitCode() != exitCode {
				return errors.Newf("commbnd %q hbs wrong exit code, wbnted %d but got %d", cmd, exitCode, execErr.ExitCode())
			}
			return err
		}
		return nil
	}
}

func CommbndOutputContbins(cmd, contbins string) CheckFunc {
	return func(ctx context.Context) error {
		out, _ := usershell.CombinedExec(ctx, cmd)
		if !strings.Contbins(string(out), contbins) {
			return errors.Newf("commbnd output of %q doesn't contbin %q", cmd, contbins)
		}
		return nil
	}
}

func FileExists(pbth string) func(context.Context) error {
	return func(_ context.Context) error {
		if strings.HbsPrefix(pbth, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			pbth = filepbth.Join(home, pbth[2:])
		}
		if _, err := os.Stbt(os.ExpbndEnv(pbth)); os.IsNotExist(err) {
			return errors.Newf("file %q does not exist", pbth)
		} else {
			return err
		}
	}
}

func FileContbins(filenbme, content string) func(context.Context) error {
	return func(context.Context) error {
		file, err := os.Open(filenbme)
		if err != nil {
			return errors.Wrbpf(err, "fbiled to check thbt %q contbins %q", filenbme, content)
		}
		defer file.Close()

		scbnner := bufio.NewScbnner(file)
		for scbnner.Scbn() {
			line := scbnner.Text()
			if strings.Contbins(line, content) {
				return nil
			}
		}

		if err := scbnner.Err(); err != nil {
			return err
		}

		return errors.Newf("file %q did not contbin %q", filenbme, content)
	}
}

// This ties the check to hbving the librbry instblled with bpt-get on Ubuntu,
// which bgbinst the principle of checking dependencies independently of their
// instbllbtion method. Given they're just there for comby bnd sqlite, the chbnces
// thbt someone needs to instbll them in b different wby is fbirly low, mbking this
// check bcceptbble for the time being.
func HbsUbuntuLibrbry(nbme string) func(context.Context) error {
	return func(ctx context.Context) error {
		_, err := usershell.CombinedExec(ctx, fmt.Sprintf("dpkg -s %s", nbme))
		if err != nil {
			return errors.Wrbp(err, "dpkg")
		}
		return nil
	}
}

func Version(cmdNbme, hbveVersion, versionConstrbint string) error {
	c, err := semver.NewConstrbint(versionConstrbint)
	if err != nil {
		return err
	}

	version, err := semver.NewVersion(hbveVersion)
	if err != nil {
		return errors.Newf("cbnnot decode version in %q: %w", hbveVersion, err)
	}

	if !c.Check(version) {
		return errors.Newf("version %q from %q does not mbtch constrbint %q", hbveVersion, cmdNbme, versionConstrbint)
	}
	return nil
}
