package op

import "github.com/go-kit/kit/log"

// Compose composes consecutive ops a and b into c such that applying
// a then b is equivalent to applying c. It does not modify a or b.
func Compose(logger log.Logger, a, b Ops) (c Ops, err error) {
	return Merge(logger, join(a, b))
}
