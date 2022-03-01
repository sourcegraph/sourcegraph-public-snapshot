package lockfiles

import "github.com/sourcegraph/sourcegraph/lib/errors"

var ErrUnsupported = errors.New("unsupported lockfile kind")
