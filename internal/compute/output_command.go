package compute

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type Output struct {
	MatchPattern  MatchPattern
	OutputPattern string
	Separator     string
}

func (c *Output) String() string {
	return fmt.Sprintf("Output with separator: (%s) -> (%s) separator: %s", c.MatchPattern.String(), c.OutputPattern, c.Separator)
}

func (c *Output) Run(context.Context, *result.FileMatch) (Result, error) {
	return nil, nil
}
