package git

import (
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// GQLRepositoryComparisonPreview implements the RepositoryComparison GraphQL type backed by static
// data.
type GQLRepositoryComparisonPreview struct {
	BaseRepository_ *graphqlbackend.RepositoryResolver
	HeadRepository_ *graphqlbackend.RepositoryResolver

	Commits_   []*graphqlbackend.GitCommitResolver
	FileDiffs_ []*diff.FileDiff
}

func (v *GQLRepositoryComparisonPreview) BaseRepository() *graphqlbackend.RepositoryResolver {
	return v.BaseRepository_
}

func (v *GQLRepositoryComparisonPreview) HeadRepository() *graphqlbackend.RepositoryResolver {
	return v.HeadRepository_
}

func (v *GQLRepositoryComparisonPreview) Range() graphqlbackend.GitRevisionRange {
	return nil
}

func (GQLRepositoryComparisonPreview) IsPreview() bool { return true }

func (v *GQLRepositoryComparisonPreview) Commits(*graphqlutil.ConnectionArgs) graphqlbackend.GitCommitConnection {
	return GitCommitConnection(v.Commits_)
}

func (v *GQLRepositoryComparisonPreview) FileDiffs(args *graphqlutil.ConnectionArgs) graphqlbackend.FileDiffConnection {
	fileDiffs := make([]graphqlbackend.FileDiff, len(v.FileDiffs_))
	for i, d := range v.FileDiffs_ {
		fileDiffs[i] = graphqlbackend.NewFileDiff(d, nil, nil)
	}
	return FileDiffConnection(fileDiffs)
}
