package dbstore

import "github.com/cockroachdb/errors"

// ErrUnknownRepository occurs when a repository does not exist.
var ErrUnknownRepository = errors.New("unknown repository")
