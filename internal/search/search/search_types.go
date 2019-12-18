package search

import (
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type SearchTypeParameters interface {
	SearchTypeInputValue()
}

func (c CommitParameters) SearchTypeParametersValue() {}
func (d DiffParameters) SearchTypeParametersValue()   {}

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
