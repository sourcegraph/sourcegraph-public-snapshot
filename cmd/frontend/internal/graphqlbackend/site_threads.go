package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

func (r *siteResolver) Threads(args *struct {
	connectionArgs
}) *siteThreadConnectionResolver {
	var opt db.ThreadsListOptions
	args.connectionArgs.set(&opt.ListOptions)
	return &siteThreadConnectionResolver{opt: opt}
}

type siteThreadConnectionResolver struct {
	opt db.ThreadsListOptions
}

func (r *siteThreadConnectionResolver) Nodes(ctx context.Context) ([]*threadResolver, error) {
	// 🚨 SECURITY: Only site admins can list threads.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	threads, err := db.Threads.List(ctx, &r.opt)
	if err != nil {
		return nil, err
	}

	var l []*threadResolver
	for _, thread := range threads {
		orgRepo, err := db.OrgRepos.GetByID(ctx, thread.OrgRepoID)
		if err != nil {
			return nil, err
		}
		org, err := db.Orgs.GetByID(ctx, orgRepo.OrgID)
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

	count, err := db.Threads.Count(ctx, r.opt)
	return int32(count), err
}
