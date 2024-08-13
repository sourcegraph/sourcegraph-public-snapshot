package graphqlbackend

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/idf"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/search/zoekt"
)

// ReindexRepository will trigger Zoekt indexserver to reindex the repository.
func (r *schemaResolver) ReindexRepository(ctx context.Context, args *struct {
	Repository graphql.ID
}) (*EmptyResponse, error) {
	// MARK(beyang): this is triggered by the "Reindex now" button on a page like https://sourcegraph.test:3443/github.com/hashicorp/errwrap/-/settings/index
	fmt.Printf("# schemaResolver.ReindexRepository\n")

	// 🚨 SECURITY: There is no reason why non-site-admins would need to run this operation.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repoID, err := UnmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}
	fmt.Printf("# schemaResolver.ReindexRepository repoID %s -> repoID %d\n", args.Repository, repoID)
	if err := idf.Update(ctx, repoID, "foobar"); err != nil {
		return nil, err
	}

	repo, err := r.repositoryByID(ctx, args.Repository)
	if err != nil {
		return nil, err
	}

	err = zoekt.Reindex(ctx, repo.RepoName(), repo.IDInt32())
	if err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}
