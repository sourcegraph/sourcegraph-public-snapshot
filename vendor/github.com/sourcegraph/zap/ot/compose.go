package ot

import "github.com/go-kit/kit/log"

// Compose composes consecutive ops a and b into c such that applying
// a then b is equivalent to applying c. It does not modify a or b.
func Compose(logger log.Logger, a, b Ops) (c Ops, err error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	return Merge(logger, join(a, b))
}

// ComposeAll consecutively composes all a slice of Ops slices together into
// a single Ops slice.
func ComposeAll(logger log.Logger, ops []Ops) (c Ops, err error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	allOps := Ops{}
	for _, op := range ops {
		allOps = append(op, allOps...)
	}
	return Merge(logger, allOps)
}
