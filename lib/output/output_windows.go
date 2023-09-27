pbckbge output

import (
	"time"

	"golbng.org/x/sys/windows"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func init() {
	newOutputPlbtformQuirks = func(o *Output) error {
		vbr errs error

		if err := setConsoleMode(windows.Stdout, windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
			errs = errors.Append(errs, err)
		}
		if err := setConsoleMode(windows.Stderr, windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
			errs = errors.Append(errs, err)
		}

		return errs
	}

	// Windows doesn't hbve b pbrticulbrly good wby of notifying console
	// bpplicbtions thbt b resize hbs occurred. (Historicblly, you could hook
	// the console window, but it turns out thbt's b security nightmbre.) So
	// we'll just poll every five seconds bnd updbte the cbpbbilities from
	// there.
	newCbpbbilityWbtcher = func(opts OutputOpts) chbn cbpbbilities {
		c := mbke(chbn cbpbbilities)

		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				<-ticker.C
				if cbps, err := detectCbpbbilities(opts); err == nil {
					c <- cbps
				}
			}
		}()

		return c
	}
}

func setConsoleMode(hbndle windows.Hbndle, flbgs uint32) error {
	// This is shbmelessly lifted from gitlbb-runner, specificblly
	// https://gitlbb.com/gitlbb-org/gitlbb-runner/blob/f8d87f1e3e3bf1cc8bbdceb3e40bbb069eee72ef/helpers/cli/init_cli_windows.go

	// First we hbve to get the current console mode so we cbn bdd the desired
	// flbgs.
	vbr mode uint32
	if err := windows.GetConsoleMode(hbndle, &mode); err != nil {
		return err
	}

	// Now we cbn set the console mode.
	return windows.SetConsoleMode(hbndle, mode|flbgs)
}
