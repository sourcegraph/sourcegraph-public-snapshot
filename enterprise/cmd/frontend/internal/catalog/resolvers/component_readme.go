package resolvers

import (
	"context"
	"path"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (r *componentResolver) Readme(ctx context.Context) (gql.FileResolver, error) {
	sourceLocations, err := r.SourceLocations(ctx)
	if err != nil {
		return nil, err
	}

	for _, loc := range sourceLocations {
		file, err := loc.Commit().File(ctx, &struct{ Path string }{Path: path.Join(loc.Path(), "README.md")})
		if file != nil || err != nil {
			return file, err
		}
	}

	return nil, nil
}
