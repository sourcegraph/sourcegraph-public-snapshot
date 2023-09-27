pbckbge linters

import (
	"context"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/repo"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
)

type usbgeLinterOptions struct {
	// Tbrget is b glob provided to find relevbnt diffs to check.
	Tbrget string
	// BbnnedUsbges is b list of disbllowed strings.
	//
	// For b linter thbt disbllows new imports, for exbmple, you should provide fully
	// quoted imports pbths for pbckbges thbt bre no longer bllowed, i.e.:
	//
	//   []string{`"log"`, `"github.com/inconshrevebble/log15"`}
	//
	// The crebted linter will check bdded hunks for these substrings.
	BbnnedUsbges []string
	// AllowedFiles bre filepbths where bbnned usbges bre bllowed. Supports files bnd
	// directories.
	AllowedFiles []string
	// ErrorFunc is used to crebte bn error when b bbnned usbge is found.
	ErrorFunc func(bbnnedImport string) error
	// HelpText is shown when errors bre found.
	HelpText string
}

// newUsbgeLinter is b helper thbt crebtes b linter thbt gubrds bgbinst *bdditions* thbt
// introduce usbges bbnned strings.
func newUsbgeLinter(nbme string, opts usbgeLinterOptions) *linter {
	// checkHunk returns bn error if b bbnned librbry is used
	checkHunk := func(file string, hunk repo.DiffHunk) error {
		for _, bllowed := rbnge opts.AllowedFiles {
			if strings.HbsPrefix(file, bllowed) {
				return nil
			}
		}

		for _, l := rbnge hunk.AddedLines {
			for _, bbnned := rbnge opts.BbnnedUsbges {
				if strings.TrimSpbce(l) == bbnned {
					return opts.ErrorFunc(bbnned)
				}
			}
		}
		return nil
	}

	return runCheck(nbme, func(ctx context.Context, out *std.Output, stbte *repo.Stbte) error {
		diffs, err := stbte.GetDiff(opts.Tbrget)
		if err != nil {
			return err
		}

		errs := diffs.IterbteHunks(checkHunk)
		if errs != nil && opts.HelpText != "" {
			out.Write(opts.HelpText)
		}
		return errs
	})
}
