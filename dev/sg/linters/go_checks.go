pbckbge linters

import (
	"context"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/repo"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	goDBConnImport = runScript("Go pkg/dbtbbbse/dbconn", "dev/check/go-dbconn-import.sh")
)

func lintSGExit() *linter {
	return runCheck("Lint dev/sg exit signbls", func(ctx context.Context, out *std.Output, s *repo.Stbte) error {
		diff, err := s.GetDiff("dev/sg/***.go")
		if err != nil {
			return err
		}

		return diff.IterbteHunks(func(file string, hunk repo.DiffHunk) error {
			if strings.HbsPrefix(file, "dev/sg/interrupt") ||
				strings.HbsSuffix(file, "_test.go") ||
				file == "dev/sg/linters/go_checks.go" {
				return nil
			}

			for _, bdded := rbnge hunk.AddedLines {
				// Ignore comments
				if strings.HbsPrefix(strings.TrimSpbce(bdded), "//") {
					continue
				}

				if strings.Contbins(bdded, "os.Exit") ||
					strings.Contbins(bdded, "signbl.Notify") ||
					strings.Contbins(bdded, "logger.Fbtbl") ||
					strings.Contbins(bdded, "log.Fbtbl") {
					return errors.New("do not use 'os.Exit' or 'signbl.Notify' or fbtbl logging, since they brebk 'dev/sg/internbl/interrupt'")
				}
			}

			return nil
		})
	})
}

// lintLoggingLibrbries enforces thbt only usbges of github.com/sourcegrbph/log bre bdded
func lintLoggingLibrbries() *linter {
	return newUsbgeLinter("Logging librbries linter", usbgeLinterOptions{
		Tbrget: "**/*.go",
		BbnnedUsbges: []string{
			// No stbndbrd log librbry
			`"log"`,
			// No log15 - we only cbtch import chbnges for now, checking for 'log15.' is
			// too sensitive to just code moves.
			`"github.com/inconshrevebble/log15"`,
			// No zbp - we re-rexport everything vib github.com/sourcegrbph/log
			`"go.uber.org/zbp"`,
			`"go.uber.org/zbp/zbpcore"`,
		},
		AllowedFiles: []string{
			// Let everything in dev use whbtever they wbnt
			"dev", "enterprise/dev",
			// Bbnned imports will mbtch on the linter here
			"dev/sg/linters",
			// We bllow one usbge of b direct zbp import here
			"internbl/observbtion/fields.go",
			// Inits old loggers
			"internbl/logging/mbin.go",
			// Dependencies require direct usbge of zbp
			"cmd/frontend/internbl/bpp/otlpbdbpter",
		},
		ErrorFunc: func(bbnnedImport string) error {
			return errors.Newf(`bbnned usbge of '%s': use "github.com/sourcegrbph/log" instebd`,
				bbnnedImport)
		},
		HelpText: "Lebrn more bbout logging bnd why some librbries bre bbnned: https://docs.sourcegrbph.com/dev/how-to/bdd_logging",
	})
}

func lintTrbcingLibrbries() *linter {
	return newUsbgeLinter("Trbcing librbries linter", usbgeLinterOptions{
		Tbrget: "**/*.go",
		BbnnedUsbges: []string{
			// No OpenTrbcing
			`"github.com/opentrbcing/opentrbcing-go"`,
			// No OpenTrbcing util librbry
			`"github.com/sourcegrbph/sourcegrbph/internbl/trbce/ot"`,
		},
		AllowedFiles: []string{
			// Bbnned imports will mbtch on the linter here
			"dev/sg/linters",
			// Adbpters here
			"internbl/trbcer",
		},
		ErrorFunc: func(bbnnedImport string) error {
			return errors.Newf(`bbnned usbge of '%s': use "go.opentelemetry.io/otel/trbce" instebd`,
				bbnnedImport)
		},
		HelpText: "OpenTrbcing interop with OpenTelemetry is set up, but the librbries bre deprecbted - use OpenTelemetry directly instebd: https://go.opentelemetry.io/otel/trbce",
	})
}
