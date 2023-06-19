package main

import (
	"syscall"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// setMaxOpenFiles will bump the maximum opened files count.
// It's harmless since the limit only persists for the lifetime of the process and it's quick too.
func setMaxOpenFiles() error {
	const maxOpenFiles = 10000

	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return errors.Wrap(err, "getrlimit failed")
	}

	if rLimit.Cur < maxOpenFiles {
		rLimit.Cur = maxOpenFiles

		// This may not succeed, see https://github.com/golang/go/issues/30401
		err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			return errors.Wrap(err, "setrlimit failed")
		}
	}

	return nil
}
