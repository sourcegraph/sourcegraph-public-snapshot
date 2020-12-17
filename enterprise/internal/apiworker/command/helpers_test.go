package command

import (
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func makeTestOperation() *observation.Operation {
	return MakeOperations(&observation.TestContext).Exec
}

var commandComparer = cmp.Comparer(func(x, y command) bool {
	return x.Key == y.Key && x.Dir == y.Dir && compareStrings(x.Env, y.Env) && compareStrings(x.Command, y.Command)
})

func compareStrings(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}

	for i, v1 := range x {
		if y[i] != v1 {
			return false
		}
	}

	return true
}
