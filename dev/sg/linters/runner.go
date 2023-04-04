package linters

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

type Runner = *check.Runner[*repo.State]

func NewRunner(out *std.Output, annotations bool, targets ...Target) Runner {
	runner := check.NewRunner(nil, out, targets)
	runner.GenerateAnnotations = annotations
	runner.AnalyticsCategory = "lint"
	runner.SuggestOnCheckFailure = func(category string, c *check.Check[*repo.State], err error) string {
		if c.Fix == nil {
			return ""
		}
		return fmt.Sprintf("Try `sg lint --fix %s` to fix this issue!", category)
	}
	return runner
}
