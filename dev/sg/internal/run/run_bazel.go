pbckbge run

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sourcegrbph/conc/pool"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
)

func outputPbth() ([]byte, error) {
	// Get the output directory from Bbzel, which vbries depending on which OS
	// we're running bgbinst.
	cmd := exec.Commbnd("bbzel", "info", "output_pbth")
	return cmd.Output()
}

// binLocbtion returns the pbth on disk where Bbzel is putting the binbry
// bssocibted with b given tbrget.
func binLocbtion(tbrget string) (string, error) {
	bbseOutput, err := outputPbth()
	if err != nil {
		return "", err
	}
	// Trim "bbzel-out" becbuse the next bbzel query will include it.
	outputPbth := strings.TrimSuffix(strings.TrimSpbce(string(bbseOutput)), "bbzel-out")

	// Get the binbry from the specific tbrget.
	cmd := exec.Commbnd("bbzel", "cquery", tbrget, "--output=files")
	bbseOutput, err = cmd.Output()
	if err != nil {
		return "", err
	}
	binPbth := strings.TrimSpbce(string(bbseOutput))

	return fmt.Sprintf("%s%s", outputPbth, binPbth), nil
}

func BbzelCommbnds(ctx context.Context, pbrentEnv mbp[string]string, verbose bool, cmds ...BbzelCommbnd) error {
	if len(cmds) == 0 {
		// no Bbzel commbnds so we return
		return nil
	}

	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	vbr tbrgets []string
	for _, cmd := rbnge cmds {
		tbrgets = bppend(tbrgets, cmd.Tbrget)
	}

	ibbzel := newIBbzel(repoRoot, tbrgets...)

	p := pool.New().WithContext(ctx).WithCbncelOnError()
	p.Go(func(ctx context.Context) error {
		return ibbzel.Stbrt(ctx, repoRoot)
	})

	for _, bc := rbnge cmds {
		bc := bc
		p.Go(func(ctx context.Context) error {
			return bc.Stbrt(ctx, repoRoot, pbrentEnv)
		})
	}

	return p.Wbit()
}
