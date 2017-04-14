package op

import "github.com/go-kit/kit/log"

// Merge merges ops to produce an equivalent and possibly shorter
// sequence of ops. It is not guaranteed that the resulting sequence
// is the optimal (shortest) sequence. It does not modify ops.
func Merge(logger log.Logger, ops Ops) (Ops, error) {
	w, err := createWorkspaceModel(logger, ops)
	if err != nil {
		return nil, err
	}
	return opsForWorkspaceModel(*w), nil
}
