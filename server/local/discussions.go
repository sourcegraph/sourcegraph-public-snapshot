package local

import (
	"fmt"
	"sort"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	app_router "src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/notif"
	"src.sourcegraph.com/sourcegraph/server/internal/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/mdutil"
)

// TODO(keegan) temporary override to make discussions more discoverable
var slackChannel = "#dev-bot-discussions"

var Discussions sourcegraph.DiscussionsServer = &discussions{}

type discussions struct{}

var _ sourcegraph.DiscussionsServer = (*discussions)(nil)

func (s *discussions) Create(ctx context.Context, in *sourcegraph.Discussion) (*sourcegraph.Discussion, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Discussions.Create"); err != nil {
		return nil, err
	}
	defer noCache(ctx)

	err := store.DiscussionsFromContext(ctx).Create(ctx, in)
	if err != nil {
		return nil, err
	}

	{
		// Send Slack notification.
		actor := notif.PersonFromContext(ctx)
		notif.Action(notif.ActionContext{
			Person:      actor,
			ActionType:  "created",
			ObjectURL:   appURL(ctx, app_router.Rel.URLToDef(in.DefKey)),
			ObjectRepo:  in.DefKey.Repo,
			ObjectType:  "discussion",
			ObjectID:    in.ID,
			ObjectTitle: in.Title,
		})

		// Notify mentioned people.
		ppl, err := mdutil.Mentions(ctx, []byte(in.Description))
		if err != nil {
			return nil, err
		}
		for _, p := range ppl {
			msg := fmt.Sprintf(
				"*%s* mentioned @%s in <%s|%s discussion #%d>: %s\n\n",
				actor.Login, p.Login,
				appURL(ctx, app_router.Rel.URLToDef(in.DefKey)),
				in.DefKey.Repo, in.ID, in.Title,
			)
			notif.Mention(p, notif.MentionContext{
				Mentioner:    actor.Login,
				MentionerURL: appURL(ctx, app_router.Rel.URLToUser(actor.Login)),
				Where:        fmt.Sprintf("in a discussion %s/%d", in.DefKey, in.ID),
				WhereURL:     appURL(ctx, app_router.Rel.URLToDef(in.DefKey)),
				SlackMsg:     msg,
			})
		}
	}

	return in, nil
}

func (s *discussions) Get(ctx context.Context, in *sourcegraph.DiscussionSpec) (*sourcegraph.Discussion, error) {
	return store.DiscussionsFromContext(ctx).Get(ctx, in.Repo, in.ID)
}

func (s *discussions) List(ctx context.Context, in *sourcegraph.DiscussionListOp) (*sourcegraph.DiscussionList, error) {
	list, err := store.DiscussionsFromContext(ctx).List(ctx, in)
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

	err := store.DiscussionsFromContext(ctx).CreateComment(ctx, in.DiscussionID, in.Comment)
	if err != nil {
		return nil, err
	}

	{
		// Send Slack notification.
		actor := notif.PersonFromContext(ctx)
		discussion, err := s.Get(ctx, &sourcegraph.DiscussionSpec{Repo: sourcegraph.RepoSpec{URI: in.Comment.DefKey.Repo}, ID: in.DiscussionID})
		if err != nil {
			return nil, err
		}
		var recipients []*sourcegraph.Person
		if discussion.Author.UID != actor.UID {
			recipients = append(recipients, notif.Person(ctx, sourcegraph.NewClientFromContext(ctx), &discussion.Author))
		}
		notif.Action(notif.ActionContext{
			Person:      actor,
			Recipients:  recipients,
			ActionType:  "commented on",
			ObjectURL:   appURL(ctx, app_router.Rel.URLToDef(in.Comment.DefKey)),
			ObjectRepo:  in.Comment.DefKey.Repo,
			ObjectType:  "discussion",
			ObjectID:    discussion.ID,
			ObjectTitle: discussion.Title,
		})

		// Notify mentioned people.
		ppl, err := mdutil.Mentions(ctx, []byte(in.Comment.Body))
		if err != nil {
			return nil, err
		}
		for _, p := range ppl {
			msg := fmt.Sprintf(
				"*%s* mentioned @%s in a comment on <%s|%s discussion #%d>: %s\n\n",
				actor.Login, p.Login,
				appURL(ctx, app_router.Rel.URLToDef(in.Comment.DefKey)),
				in.Comment.DefKey.Repo, discussion.ID, discussion.Title,
			)
			notif.Mention(p, notif.MentionContext{
				Mentioner:    actor.Login,
				MentionerURL: appURL(ctx, app_router.Rel.URLToUser(actor.Login)),
				Where:        fmt.Sprintf("in a comment of discussion %s/%d", in.Comment.DefKey, discussion.ID),
				WhereURL:     appURL(ctx, app_router.Rel.URLToDef(in.Comment.DefKey)),
				SlackMsg:     msg,
			})
		}
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
