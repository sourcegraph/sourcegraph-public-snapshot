package search

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type SearchTypeParameters interface {
	SearchTypeInputValue()
}

func (c CommitParameters) SearchTypeParametersValue() {}

type CommitParameters struct {
	RepoRevs           *RepositoryRevisions
	Info               *PatternInfo
	Query              *query.Query
	Diff               bool
	TextSearchOptions  git.TextSearchOptions
	ExtraMessageValues []string
}
