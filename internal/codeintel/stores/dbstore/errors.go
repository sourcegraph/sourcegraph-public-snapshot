package dbstore

import "github.com/sourcegraph/sourcegraph/lib/errors"

// ErrUnknownRepository occurs when a repository does not exist.
var ErrUnknownRepository = errors.New("unknown repository")
