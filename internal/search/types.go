package search

import (
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/symbols/protocol"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type TypeParameters interface {
	typeParametersValue()
}

func (c CommitParameters) typeParametersValue()  {}
func (d DiffParameters) typeParametersValue()    {}
func (s SymbolsParameters) typeParametersValue() {}

type CommitParameters struct {
	RepoRevs           *RepositoryRevisions
	Info               *PatternInfo
	Query              *query.Query
	Diff               bool
	TextSearchOptions  git.TextSearchOptions
	ExtraMessageValues []string
}

type DiffParameters struct {
	Repo    gitserver.Repo
	Options git.RawLogDiffSearchOptions
}

type SymbolsParameters protocol.SearchArgs
