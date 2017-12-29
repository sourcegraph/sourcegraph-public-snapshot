package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func (r *siteResolver) Threads(args *struct {
	connectionArgs
}) *siteThreadConnectionResolver {
	return &siteThreadConnectionResolver{
		connectionResolverCommon: newConnectionResolverCommon(args.connectionArgs),
	}
}

type siteThreadConnectionResolver struct {
	connectionResolverCommon
}

func (r *siteThreadConnectionResolver) Nodes(ctx context.Context) ([]*threadResolver, error) {
	// 🚨 SECURITY: Only site admins can list threads.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	threadsList, err := localstore.Threads.List(ctx)
	if err != nil {
		return nil, err
	}

	var l []*threadResolver
	for _, thread := range threadsList {
		orgRepo, err := localstore.OrgRepos.GetByID(ctx, thread.OrgRepoID)
		if err != nil {
			return nil, err
		}
		org, err := localstore.Orgs.GetByID(ctx, orgRepo.OrgID)
		if err != nil {
			return nil, err
		}

		l = append(l, &threadResolver{
			thread: thread,
			repo:   orgRepo,
			org:    org,
		})
	}
	return l, nil
}

func (r *siteThreadConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	// 🚨 SECURITY: Only site admins can count threads.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return 0, err
	}

	count, err := localstore.Threads.Count(ctx)
	return int32(count), err
}
