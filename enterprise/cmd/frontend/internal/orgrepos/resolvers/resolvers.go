package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewResolver(db database.DB) graphqlbackend.OrgRepositoryResolver {
	return &resolver{db: db}
}

var _ graphqlbackend.OrgRepositoryResolver = (*resolver)(nil)

type resolver struct {
	db database.DB
}

func (r *resolver) OrgRepositories(ctx context.Context, args *graphqlbackend.ListOrgRepositoriesArgs, org *types.Org) (graphqlbackend.RepositoryConnectionResolver, error) {
	if err := backend.CheckOrgExternalServices(ctx, r.db, org.ID); err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Only org members can list the org repositories.
	if err := backend.CheckOrgAccess(ctx, r.db, org.ID); err != nil {
		if err == backend.ErrNotAnOrgMember {
			return nil, errors.New("must be a member of this organization to view its repositories")
		}
		return nil, err
	}

	opt := database.ReposListOptions{}
	if args.Query != nil {
		opt.Query = *args.Query
	}
	if args.First != nil {
		opt.LimitOffset = &database.LimitOffset{Limit: int(*args.First)}
	}
	if args.After != nil {
		cursor, err := graphqlbackend.UnmarshalRepositoryCursor(args.After)
		if err != nil {
			return nil, err
		}
		opt.Cursors = append(opt.Cursors, cursor)
	} else {
		opt.Cursors = append(opt.Cursors, &types.Cursor{Direction: "next"})
	}
	if args.OrderBy == nil {
		opt.OrderBy = database.RepoListOrderBy{{
			Field:      "name",
			Descending: false,
		}}
	} else {
		opt.OrderBy = database.RepoListOrderBy{{
			Field:      graphqlbackend.ToDBRepoListColumn(*args.OrderBy),
			Descending: args.Descending,
		}}
	}

	if args.ExternalServiceIDs == nil || len(*args.ExternalServiceIDs) == 0 {
		opt.OrgID = org.ID
	} else {
		var idArray []int64
		for i, externalServiceID := range *args.ExternalServiceIDs {
			id, err := graphqlbackend.UnmarshalExternalServiceID(*externalServiceID)
			if err != nil {
				return nil, err
			}
			idArray[i] = id
		}
		opt.ExternalServiceIDs = idArray
	}

	return graphqlbackend.NewRepositoryConnectionResolver(r.db, opt, args.Cloned, args.NotCloned, args.Indexed, args.NotIndexed), nil
}
