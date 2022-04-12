package compute

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type Command interface {
	command()
	Run(context.Context, database.DB, result.Match) (Result, error)
	ToSearchPattern() string
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
