package git

import (
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// GQLRepositoryComparisonPreview implements the RepositoryComparison GraphQL type backed by static
// data.
type GQLRepositoryComparisonPreview struct {
	BaseRepository_ *graphqlbackend.RepositoryResolver
	HeadRepository_ *graphqlbackend.RepositoryResolver

	// TODO!(sqs): hack, dummy to make this type implement RepositoryComparison including fields that are TODO
	*graphqlbackend.RepositoryComparisonResolver

	FileDiffs_ []*diff.FileDiff
}

func (v *GQLRepositoryComparisonPreview) BaseRepository() *graphqlbackend.RepositoryResolver {
	return v.BaseRepository_
}

func (v *GQLRepositoryComparisonPreview) HeadRepository() *graphqlbackend.RepositoryResolver {
	return v.HeadRepository_
}

func (v *GQLRepositoryComparisonPreview) FileDiffs(args *struct {
	First *int32
}) graphqlbackend.FileDiffConnection {
	fileDiffs := make([]graphqlbackend.FileDiff, len(v.FileDiffs_))
	for i, d := range v.FileDiffs_ {
		fileDiffs[i] = graphqlbackend.NewFileDiff(d, nil)
	}
	return FileDiffConnection(fileDiffs)
}
