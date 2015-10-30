package local

import (
	"fmt"
	"net/url"
	"sort"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	app_router "src.sourcegraph.com/sourcegraph/app/router"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/notif"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/mdutil"
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

	{
		// Build list of recipients
		recipients, err := mdutil.Mentions(ctx, []byte(in.Description))
		if err != nil {
			return nil, err
		}

		// Send notification.
		permalink := appURL(ctx, app_router.Rel.URLToRepoDiscussion(string(in.DefKey.Repo), in.ID))
		cl := sourcegraph.NewClientFromContext(ctx)
		cl.Notify.GenericEvent(ctx, &sourcegraph.NotifyGenericEvent{
			Actor:       notif.UserFromContext(ctx),
			Recipients:  recipients,
			ActionType:  "created",
			ObjectURL:   permalink,
			ObjectRepo:  in.DefKey.Repo,
			ObjectType:  "discussion",
			ObjectID:    in.ID,
			ObjectTitle: in.Title,
		})
	}

	{
		events.Publish(events.Event{
			EventID: notif.DiscussionCreateEvent,
			Payload: notif.Payload{
				Type:        notif.DiscussionCreateEvent,
				UserSpec:    authpkg.UserSpecFromContext(ctx),
				ActionType:  "created",
				ObjectID:    in.ID,
				ObjectRepo:  in.DefKey.Repo,
				ObjectTitle: in.Title,
				ObjectType:  "discussion",
				ObjectURL:   appURL(ctx, app_router.Rel.URLToDef(in.DefKey)),
				Object:      in,
			},
		})
	}

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

	{
		// Build list of recipients
		cl := sourcegraph.NewClientFromContext(ctx)
		actor := notif.UserFromContext(ctx)
		discussion, err := s.Get(ctx, &sourcegraph.DiscussionSpec{Repo: sourcegraph.RepoSpec{URI: in.Comment.DefKey.Repo}, ID: in.DiscussionID})
		if err != nil {
			return nil, err
		}
		var recipients []*sourcegraph.UserSpec
		if discussion.Author.UID != actor.UID {
			recipients = append(recipients, &discussion.Author)
		}
		ppl, err := mdutil.Mentions(ctx, []byte(in.Comment.Body))
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, ppl...)

		// Send notification
		permalink := appURL(ctx, app_router.Rel.URLToRepoDiscussion(string(in.Comment.DefKey.Repo), discussion.ID))
		cl.Notify.GenericEvent(ctx, &sourcegraph.NotifyGenericEvent{
			Actor:       actor,
			Recipients:  recipients,
			ActionType:  "commented on",
			ObjectURL:   permalink,
			ObjectRepo:  in.Comment.DefKey.Repo,
			ObjectType:  "discussion",
			ObjectID:    discussion.ID,
			ObjectTitle: discussion.Title,
		})
	}

	{
		events.Publish(events.Event{
			EventID: notif.DiscussionCommentEvent,
			Payload: notif.Payload{
				Type:       notif.DiscussionCommentEvent,
				UserSpec:   authpkg.UserSpecFromContext(ctx),
				ActionType: "commented on",
				ObjectID:   in.DiscussionID,
				ObjectRepo: in.Comment.DefKey.Repo,
				ObjectType: "discussion",
				ObjectURL:  appURL(ctx, app_router.Rel.URLToDef(in.Comment.DefKey)),
				Object:     in.Comment,
			},
		})
	}

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
