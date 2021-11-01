package compute

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type Command interface {
	command()
	Run(context.Context, *result.FileMatch) (Result, error)
	String() string
}

var (
	_ Command = (*MatchOnly)(nil)
	_ Command = (*Replace)(nil)
	_ Command = (*Output)(nil)
)

func (MatchOnly) command() {}
func (Replace) command()   {}
func (Output) command()    {}
