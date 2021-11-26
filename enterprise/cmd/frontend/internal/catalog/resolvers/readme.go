package resolvers

import (
	"context"
	"path"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (r *componentResolver) Readme(ctx context.Context) (gql.FileResolver, error) {
	slocs, err := r.sourceLocationSetResolver(ctx)
	if err != nil {
		return nil, err
	}
	return slocs.Readme(ctx)
}

func (r *rootResolver) GitTreeEntryReadme(ctx context.Context, treeEntry *gql.GitTreeEntryResolver) (gql.FileResolver, error) {
	return sourceLocationSetResolverFromTreeEntry(treeEntry, r.db).Readme(ctx)
}

func (r *sourceLocationSetResolver) Readme(ctx context.Context) (gql.FileResolver, error) {
	for _, sloc := range r.slocs {
		file, err := sloc.commit.File(ctx, &struct{ Path string }{Path: path.Join(sloc.path, "README.md")})
		if file != nil || err != nil {
			return file, err
		}
	}
	return nil, nil
}
