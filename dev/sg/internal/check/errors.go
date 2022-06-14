package check

import "errors"

// ErrSkipPostFixCheck can be returned a Fix implementation to not run a Check again
// after the Fix has concluded.
var ErrSkipPostFixCheck = errors.New("skip post-fix check")
