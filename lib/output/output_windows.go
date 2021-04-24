package output

import (
	"github.com/hashicorp/go-multierror"
	"golang.org/x/sys/windows"
)

func init() {
	newOutputPlatformQuirks = func(o *Output) error {
		var errs *multierror.Error

		if err := setConsoleMode(windows.Stdout, windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
			errs = multierror.Append(errs, err)
		}
		if err := setConsoleMode(windows.Stderr, windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
			errs = multierror.Append(errs, err)
		}

		return errs.ErrorOrNil()
	}
}

func setConsoleMode(handle windows.Handle, flags uint32) error {
	// This is shamelessly lifted from gitlab-runner, specifically
	// https://gitlab.com/gitlab-org/gitlab-runner/blob/f8d87f1e3e3af1cc8aadcea3e40bbb069eee72ef/helpers/cli/init_cli_windows.go

	// First we have to get the current console mode so we can add the desired
	// flags.
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return err
	}

	// Now we can set the console mode.
	return windows.SetConsoleMode(handle, mode|flags)
}
