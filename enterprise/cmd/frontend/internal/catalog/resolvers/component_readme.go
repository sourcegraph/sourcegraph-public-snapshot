package resolvers

import (
	"context"
	"path"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (r *componentResolver) Readme(ctx context.Context) (gql.FileResolver, error) {
	sourceLocations, err := r.sourceLocations(ctx)
	if err != nil {
		return nil, err
	}

	for _, sloc := range sourceLocations {
		file, err := sloc.commit.File(ctx, &struct{ Path string }{Path: path.Join(sloc.path, "README.md")})
		if file != nil || err != nil {
			return file, err
		}
	}

	return nil, nil
}
