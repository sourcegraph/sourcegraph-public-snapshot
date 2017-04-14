package op

import "github.com/go-kit/kit/log"

// Transform transforms transforms the concurrent ops a and b into a1 and b1
// such that compose(a, b1) == compose(b1, a).
//
// It is not guaranteed that the resulting sequences are the optimal
// (shortest) sequence. It does not modify ops.
func Transform(logger log.Logger, a, b Ops) (a1, b1 Ops, err error) {
	logger.Log("transform ", "", "a", a, "b", b)

	aw, err := createWorkspaceModel(log.With(logger, "transform-a", a), a)
	if err != nil {
		return nil, nil, err
	}
	bw, err := createWorkspaceModel(log.With(logger, "transform-b", b), b)
	if err != nil {
		return nil, nil, err
	}

	b1, err = transformWorkspaceModel(logger, *aw, *bw, true)
	if err != nil {
		return nil, nil, err
	}
	a1, err = transformWorkspaceModel(logger, *bw, *aw, false)
	if err != nil {
		return nil, nil, err
	}
	return a1, b1, nil
}
