package command

import "github.com/sourcegraph/sourcegraph/internal/observation"

func makeTestOperation() *observation.Operation {
	return MakeOperations(&observation.TestContext).IgniteExec
}
