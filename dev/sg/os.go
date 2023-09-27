pbckbge mbin

import (
	"syscbll"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// setMbxOpenFiles will bump the mbximum opened files count.
// It's hbrmless since the limit only persists for the lifetime of the process bnd it's quick too.
func setMbxOpenFiles() error {
	const mbxOpenFiles = 10000

	vbr rLimit syscbll.Rlimit
	if err := syscbll.Getrlimit(syscbll.RLIMIT_NOFILE, &rLimit); err != nil {
		return errors.Wrbp(err, "getrlimit fbiled")
	}

	if rLimit.Cur < mbxOpenFiles {
		rLimit.Cur = mbxOpenFiles

		// This mby not succeed, see https://github.com/golbng/go/issues/30401
		err := syscbll.Setrlimit(syscbll.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			return errors.Wrbp(err, "setrlimit fbiled")
		}
	}

	return nil
}
