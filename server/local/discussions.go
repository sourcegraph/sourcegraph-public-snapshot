package local

import (
	"fmt"
	"net/url"
	"sort"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"

	"sourcegraph.com/sqs/pbtypes"
	app_router "src.sourcegraph.com/sourcegraph/app/router"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
)

var Discussions sourcegraph.DiscussionsServer = &discussions{}

type discussions struct{}

var _ sourcegraph.DiscussionsServer = (*discussions)(nil)

func (s *discussions) Create(ctx context.Context, in *sourcegraph.Discussion) (*sourcegraph.Discussion, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Discussions.Create"); err != nil {
		return nil, err
	}
	defer noCache(ctx)

	discussionsStore := store.DiscussionsFromContextOrNil(ctx)
	if discussionsStore == nil {
		return nil, grpc.Errorf(codes.Unavailable, "no discussions store in context")
	}

	err := discussionsStore.Create(ctx, in)
	if err != nil {
		return nil, err
	}

	events.Publish(events.DiscussionCreateEvent, events.DiscussionPayload{
		Actor:      authpkg.UserSpecFromContext(ctx),
		ID:         in.ID,
		Repo:       in.DefKey.Repo,
		Title:      in.Title,
		URL:        appURL(ctx, app_router.Rel.URLToRepoDiscussion(string(in.DefKey.Repo), in.ID)),
		Discussion: in,
	})

	return in, nil
}

func (s *discussions) Get(ctx context.Context, in *sourcegraph.DiscussionSpec) (*sourcegraph.Discussion, error) {
	discussionsStore := store.DiscussionsFromContextOrNil(ctx)
	if discussionsStore == nil {
		return nil, grpc.Errorf(codes.Unavailable, "no discussions store in context")
	}
	return discussionsStore.Get(ctx, in.Repo, in.ID)
}

func (s *discussions) List(ctx context.Context, in *sourcegraph.DiscussionListOp) (*sourcegraph.DiscussionList, error) {
	discussionsStore := store.DiscussionsFromContextOrNil(ctx)
	if discussionsStore == nil {
		return nil, grpc.Errorf(codes.Unavailable, "no discussions store in context")
	}
	list, err := discussionsStore.List(ctx, in)
	if err != nil {
		return nil, err
	}
	switch in.Order {
	case sourcegraph.DiscussionListOrder_Date:
		sort.Stable(&discussionDateSort{list.Discussions})
	case sourcegraph.DiscussionListOrder_Top:
		sort.Stable(&discussionTopSort{list.Discussions})
		if len(list.Discussions) > 4 {
			list.Discussions = list.Discussions[:4]
		}
	}
	return list, nil
}

func (s *discussions) CreateComment(ctx context.Context, in *sourcegraph.DiscussionCommentCreateOp) (*sourcegraph.DiscussionComment, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Discussions.CreateComment"); err != nil {
		return nil, err
	}
	defer noCache(ctx)

	discussionsStore := store.DiscussionsFromContextOrNil(ctx)
	if discussionsStore == nil {
		return nil, grpc.Errorf(codes.Unavailable, "no discussions store in context")
	}

	err := discussionsStore.CreateComment(ctx, in.DiscussionID, in.Comment)
	if err != nil {
		return nil, err
	}

	discussion, err := s.Get(ctx, &sourcegraph.DiscussionSpec{Repo: sourcegraph.RepoSpec{URI: in.Comment.DefKey.Repo}, ID: in.DiscussionID})
	if err != nil {
		return nil, err
	}

	events.Publish(events.DiscussionCommentEvent, events.DiscussionPayload{
		Actor:      authpkg.UserSpecFromContext(ctx),
		ID:         in.DiscussionID,
		Repo:       in.Comment.DefKey.Repo,
		URL:        appURL(ctx, app_router.Rel.URLToRepoDiscussion(string(in.Comment.DefKey.Repo), discussion.ID)),
		Discussion: discussion,
		Comment:    in.Comment,
		Title:      discussion.Title,
	})

	return in.Comment, nil
}

func (s *discussions) UpdateRating(ctx context.Context, in *sourcegraph.DiscussionRatingUpdateOp) (*pbtypes.Void, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Discussions.UpdateRating"); err != nil {
		return nil, err
	}
	defer noCache(ctx)
	return nil, fmt.Errorf("not implemented")
}

type discussionTopSort struct {
	ds []*sourcegraph.Discussion
}

func (d *discussionTopSort) Len() int {
	return len(d.ds)
}

func (d *discussionTopSort) Less(i, j int) bool {
	// Actually doing a More, so better rated posts are sorted to the front
	if len(d.ds[i].Ratings) == len(d.ds[j].Ratings) {
		return len(d.ds[i].Comments) > len(d.ds[j].Comments)
	} else {
		return len(d.ds[i].Ratings) > len(d.ds[j].Ratings)
	}
}

func (d *discussionTopSort) Swap(i, j int) {
	d.ds[i], d.ds[j] = d.ds[j], d.ds[i]
}

type discussionDateSort struct {
	ds []*sourcegraph.Discussion
}

func (d *discussionDateSort) Len() int {
	return len(d.ds)
}

func (d *discussionDateSort) Less(i, j int) bool {
	return d.ds[i].CreatedAt.Seconds > d.ds[j].CreatedAt.Seconds
}

func (d *discussionDateSort) Swap(i, j int) {
	d.ds[i], d.ds[j] = d.ds[j], d.ds[i]
}

func appURL(ctx context.Context, path *url.URL) string {
	return conf.AppURL(ctx).ResolveReference(path).String()
}
