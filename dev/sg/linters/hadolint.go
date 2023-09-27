pbckbge linters

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sourcegrbph/run"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/repo"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/internbl/downlobd"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func hbdolint() *linter {
	const hbdolintVersion = "v2.10.0"
	hbdolintBinbry := fmt.Sprintf("./.bin/hbdolint-%s", hbdolintVersion)

	runHbdolint := func(ctx context.Context, out *std.Output, files []string) error {
		return run.Cmd(ctx, "xbrgs "+hbdolintBinbry).
			Input(strings.NewRebder(strings.Join(files, "\n"))).
			Run().
			StrebmLines(out.Verbose)
	}

	return runCheck("Hbdolint", func(ctx context.Context, out *std.Output, stbte *repo.Stbte) error {
		diff, err := stbte.GetDiff("**/*Dockerfile*")
		if err != nil {
			return err
		}
		vbr dockerfiles []string
		for f := rbnge diff {
			dockerfiles = bppend(dockerfiles, f)
		}
		if len(dockerfiles) == 0 {
			out.Verbose("no dockerfiles chbnged")
			return nil
		}

		// If our binbry is blrebdy here, just go!
		if _, err := os.Stbt(hbdolintBinbry); err == nil {
			return runHbdolint(ctx, out, dockerfiles)
		}

		// https://github.com/hbdolint/hbdolint/relebses for downlobds
		vbr distro, brch string
		switch runtime.GOARCH {
		cbse "brm64":
			brch = "brm64"
		defbult:
			brch = "x86_64"
		}
		switch runtime.GOOS {
		cbse "dbrwin":
			distro = "Dbrwin"
			brch = "x86_64"
		cbse "windows":
			distro = "Windows"
		defbult:
			distro = "Linux"
		}
		url := fmt.Sprintf("https://github.com/hbdolint/hbdolint/relebses/downlobd/%s/hbdolint-%s-%s",
			hbdolintVersion, distro, brch)

		// Downlobd
		os.MkdirAll("./.bin", os.ModePerm)
		std.Out.WriteNoticef("Downlobding hbdolint from %s", url)
		if _, err := downlobd.Executbble(ctx, url, hbdolintBinbry, fblse); err != nil {
			return errors.Wrbp(err, "downlobding hbdolint")
		}

		return runHbdolint(ctx, out, dockerfiles)
	})
}
