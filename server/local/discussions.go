package local

import (
	"fmt"
	"sort"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	app_router "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/ext/slack"
	"sourcegraph.com/sourcegraph/sourcegraph/notif"
	"sourcegraph.com/sourcegraph/sourcegraph/server/internal/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sourcegraph/sourcegraph/util/mdutil"
	"sourcegraph.com/sqs/pbtypes"
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
		userStr, err := getUserDisplayName(ctx)
		if err != nil {
			return nil, err
		}
		msg := fmt.Sprintf("*%s* created a discussion <%s|%s discussion #%d>: %s",
			userStr,
			appURL(ctx, app_router.Rel.URLToDef(in.DefKey)),
			in.DefKey.Repo, in.ID, in.Title,
		)
		go slack.PostMessage(slack.PostOpts{Msg: msg, Channel: slackChannel})

		// Notify mentioned people.
		ppl, err := mdutil.Mentions(ctx, []byte(in.Description))
		if err != nil {
			return nil, err
		}
		for _, p := range ppl {
			msg := fmt.Sprintf(
				"*%s* mentioned @%s in <%s|%s discussion #%d>: %s\n\n",
				userStr, p.Login,
				appURL(ctx, app_router.Rel.URLToDef(in.DefKey)),
				in.DefKey.Repo, in.ID, in.Title,
			)
			notif.Mention(p, notif.Context{
				Mentioner:    userStr,
				MentionerURL: appURL(ctx, app_router.Rel.URLToUser(userStr)),
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
		userStr, err := getUserDisplayName(ctx)
		if err != nil {
			return nil, err
		}
		discussion, err := s.Get(ctx, &sourcegraph.DiscussionSpec{Repo: sourcegraph.RepoSpec{URI: in.Comment.DefKey.Repo}, ID: in.DiscussionID})
		if err != nil {
			return nil, err
		}
		msg := fmt.Sprintf("*%s* commented on a discussion <%s|%s discussion #%d>: %s",
			userStr,
			appURL(ctx, app_router.Rel.URLToDef(in.Comment.DefKey)),
			in.Comment.DefKey.Repo, discussion.ID, discussion.Title,
		)
		go slack.PostMessage(slack.PostOpts{Msg: msg, Channel: slackChannel})

		// Notify mentioned people.
		ppl, err := mdutil.Mentions(ctx, []byte(in.Comment.Body))
		if err != nil {
			return nil, err
		}
		for _, p := range ppl {
			msg := fmt.Sprintf(
				"*%s* mentioned @%s in a comment on <%s|%s discussion #%d>: %s\n\n",
				userStr, p.Login,
				appURL(ctx, app_router.Rel.URLToDef(in.Comment.DefKey)),
				in.Comment.DefKey.Repo, discussion.ID, discussion.Title,
			)
			notif.Mention(p, notif.Context{
				Mentioner:    userStr,
				MentionerURL: appURL(ctx, app_router.Rel.URLToUser(userStr)),
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
