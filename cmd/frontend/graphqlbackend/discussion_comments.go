package graphqlbackend

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/discussions"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/markdown"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func marshalDiscussionCommentID(dbID int64) graphql.ID {
	return relay.MarshalID("DiscussionComment", strconv.FormatInt(dbID, 36))
}

func unmarshalDiscussionCommentID(id graphql.ID) (dbID int64, err error) {
	var dbIDStr string
	err = relay.UnmarshalSpec(id, &dbIDStr)
	if err == nil {
		dbID, err = strconv.ParseInt(dbIDStr, 36, 64)
	}
	return
}

// discussionCommentByID looks up a DiscussionComment by its GraphQL ID.
func discussionCommentByID(ctx context.Context, id graphql.ID) (*discussionCommentResolver, error) {
	dbID, err := unmarshalDiscussionCommentID(id)
	if err != nil {
		return nil, err
	}
	// 🚨 SECURITY: No authentication is required to get a discussion comment. Discussion comments
	// are public unless the Sourcegraph instance itself (and inherently, the GraphQL API) is
	// private.
	comment, err := db.DiscussionComments.Get(ctx, dbID)
	if err != nil {
		return nil, err
	}
	return &discussionCommentResolver{c: comment}, nil
}

type discussionCommentResolver struct {
	c *types.DiscussionComment
}

func (r *discussionCommentResolver) ID() graphql.ID {
	return marshalDiscussionCommentID(r.c.ID)
}

func (r *discussionCommentResolver) IDWithoutKind() string {
	return strconv.FormatInt(r.c.ID, 10)
}

func (r *discussionCommentResolver) Thread(ctx context.Context) (*discussionThreadResolver, error) {
	thread, err := db.DiscussionThreads.Get(ctx, r.c.ThreadID)
	if err != nil {
		return nil, errors.Wrap(err, "DiscussionThreads.Get")
	}
	return &discussionThreadResolver{t: thread}, nil
}

func (r *discussionCommentResolver) Author(ctx context.Context) (*UserResolver, error) {
	return UserByIDInt32(ctx, r.c.AuthorUserID)
}

func (r *discussionCommentResolver) Contents(ctx context.Context) (string, error) {
	if strings.TrimSpace(r.c.Contents) != "" {
		return r.c.Contents, nil
	}
	thread, err := db.DiscussionThreads.Get(ctx, r.c.ThreadID)
	if err != nil {
		return "", errors.Wrap(err, "DiscussionThreads.Get")
	}
	return thread.Title, nil
}

func (r *discussionCommentResolver) HTML(ctx context.Context, args *struct{ Options *markdownOptions }) (string, error) {
	contents, err := r.Contents(ctx)
	if err != nil {
		return "", err
	}
	return markdown.Render(contents, nil), nil
}

func (r *discussionCommentResolver) InlineURL(ctx context.Context) (*string, error) {
	thread, err := db.DiscussionThreads.Get(ctx, r.c.ThreadID)
	if err != nil {
		return nil, errors.Wrap(err, "DiscussionThreads.Get")
	}
	url, err := discussions.URLToInlineComment(ctx, thread, r.c)
	if err != nil || url == nil {
		return nil, err
	}
	return strptr(url.String()), nil
}

func (r *discussionCommentResolver) CreatedAt(ctx context.Context) string {
	return r.c.CreatedAt.Format(time.RFC3339)
}

func (r *discussionCommentResolver) UpdatedAt(ctx context.Context) string {
	return r.c.UpdatedAt.Format(time.RFC3339)
}

func (r *discussionCommentResolver) Reports(ctx context.Context) []string {
	// 🚨 SECURITY: Only site admins can read reports.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return []string{}
	}
	if dc := conf.Get().Discussions; dc != nil && !dc.AbuseProtection {
		return []string{}
	}
	return r.c.Reports
}

func (r *discussionCommentResolver) CanReport(ctx context.Context) bool {
	if dc := conf.Get().Discussions; dc != nil && !dc.AbuseProtection {
		return false
	}
	// Only signed in users may update/report a discussion comment.
	currentUser, err := CurrentUser(ctx)
	if err != nil {
		return false
	}
	if currentUser == nil {
		return false
	}
	return true
}

func (r *discussionCommentResolver) CanClearReports(ctx context.Context) bool {
	if dc := conf.Get().Discussions; dc != nil && !dc.AbuseProtection {
		return false
	}
	// Only site admins can clear reports.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return false
	}
	return true
}

func (r *discussionCommentResolver) CanDelete(ctx context.Context) bool {
	// Only site admins can delete discussion comments.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return false
	}
	return true
}

func (*schemaResolver) DiscussionComments(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
	AuthorUserID *graphql.ID
}) (*discussionCommentsConnectionResolver, error) {
	if err := viewerCanUseDiscussions(ctx); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: No authentication is required to list the comments on a
	// discussion. They are public unless the Sourcegraph instance itself (and
	// inherently, the GraphQL API) is private.

	opt := &db.DiscussionCommentsListOptions{}
	args.ConnectionArgs.Set(&opt.LimitOffset)
	if args.AuthorUserID != nil {
		userID, err := UnmarshalUserID(*args.AuthorUserID)
		if err != nil {
			return nil, err
		}
		opt.AuthorUserID = &userID
	}
	return &discussionCommentsConnectionResolver{opt: opt}, nil
}

// checkSignedInAndEmailVerified returns an error if there is not a user signed
// in, or if that user does not have at least one verified email address.
func checkSignedInAndEmailVerified(ctx context.Context) (*UserResolver, error) {
	currentUser, err := CurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, errors.New("no current user")
	}
	emails, err := currentUser.Emails(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Emails")
	}
	for _, email := range emails {
		if email.Verified() {
			return currentUser, nil
		}
	}
	return nil, errors.New("account email must be verified to perform this action")
}

func (r *discussionsMutationResolver) AddCommentToThread(ctx context.Context, args *struct {
	ThreadID graphql.ID
	Contents string
}) (*discussionThreadResolver, error) {
	// 🚨 SECURITY: Only signed in users with a verified email may add comments
	// to a discussion thread.
	//
	// The verified email requirement for public instances is a security
	// measure to prevent spam. For private instances, it is a UX feature
	// (because we would not be able to send the author of this comment email
	// notifications anyway).
	currentUser, err := checkSignedInAndEmailVerified(ctx)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(args.Contents) == "" {
		return nil, errors.New("cannot add empty comments to threads")
	}

	// Create the comment on the thread.
	threadID, err := unmarshalDiscussionThreadID(args.ThreadID)
	if err != nil {
		return nil, err
	}

	updatedThread, err := discussions.InsecureAddCommentToThread(ctx, &types.DiscussionComment{
		ThreadID:     threadID,
		AuthorUserID: currentUser.user.ID,
		Contents:     args.Contents,
	})
	if err != nil {
		return nil, errors.Wrap(err, "AddCommentToThread")
	}
	return &discussionThreadResolver{t: updatedThread}, nil
}

func (r *discussionsMutationResolver) UpdateComment(ctx context.Context, args *struct {
	Input *struct {
		CommentID    graphql.ID
		Contents     *string
		Delete       *bool
		Report       *string
		ClearReports *bool
	}
}) (*discussionThreadResolver, error) {
	commentID, err := unmarshalDiscussionThreadID(args.Input.CommentID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only signed in users may update a discussion comment.
	currentUser, err := CurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, errors.New("no current user")
	}

	var delete bool
	if args.Input.Delete != nil && *args.Input.Delete {
		// 🚨 SECURITY: Only site admins can delete discussion comments.
		if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
			return nil, err
		}
		delete = *args.Input.Delete
	}

	var clearReports bool
	if args.Input.ClearReports != nil && *args.Input.ClearReports {
		// 🚨 SECURITY: Only site admins can clear reports.
		if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
			return nil, err
		}
		if dc := conf.Get().Discussions; dc != nil && !dc.AbuseProtection {
			return nil, errors.New("cannot clear reports; discussions.abuseProtection is disabled")
		}
		clearReports = *args.Input.ClearReports
	}

	if args.Input.Report != nil {
		if dc := conf.Get().Discussions; dc != nil && !dc.AbuseProtection {
			return nil, errors.New("cannot report comment; discussions.abuseProtection is disabled")
		}
		newReport := fmt.Sprintf(`"%s"\n\nreported by @%s`, *args.Input.Report, currentUser.user.Username)
		args.Input.Report = &newReport
	}

	if args.Input.Contents != nil {
		// 🚨 SECURITY: Only site admins and the comment author can update the contents.
		comment, err := db.DiscussionComments.Get(ctx, commentID)
		if err != nil {
			return nil, err
		}
		err = backend.CheckSiteAdminOrSameUser(ctx, comment.AuthorUserID)
		if err != nil {
			return nil, err
		}
	}

	// Resolve the thread ID of the comment first so we can return the updated
	// thread later. We must do this now because the comment may be deleted
	// below (Update may return nil).
	comment, err := db.DiscussionComments.Get(ctx, commentID)
	if err != nil {
		return nil, errors.Wrap(err, "DiscussionComments.Get")
	}
	threadID := comment.ThreadID

	updatedComment, err := db.DiscussionComments.Update(ctx, commentID, &db.DiscussionCommentsUpdateOptions{
		Contents:     args.Input.Contents,
		Delete:       delete,
		Report:       args.Input.Report,
		ClearReports: clearReports,
	})
	if err != nil {
		return nil, errors.Wrap(err, "DiscussionComments.Update")
	}
	thread, err := db.DiscussionThreads.Get(ctx, threadID)
	if err != nil {
		return nil, errors.Wrap(err, "DiscussionThreads.Get")
	}
	if args.Input.Report != nil {
		c := updatedComment
		if c == nil {
			c = comment
		}
		discussions.NotifyCommentReported(currentUser.user, thread, c)
	}
	return &discussionThreadResolver{t: thread}, nil
}

// discussionCommentsConnectionResolver resolves a list of discussion comments.
//
// 🚨 SECURITY: When instantiating an discussionCommentsConnectionResolver
// value, the caller MUST check permissions.
type discussionCommentsConnectionResolver struct {
	opt *db.DiscussionCommentsListOptions

	// cache results because they are used by multiple fields
	once     sync.Once
	comments []*types.DiscussionComment
	err      error
}

func (r *discussionCommentsConnectionResolver) compute(ctx context.Context) ([]*types.DiscussionComment, error) {
	r.once.Do(func() {
		opt2 := *r.opt
		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			opt2.Limit++ // so we can detect if there is a next page
		}

		r.comments, r.err = db.DiscussionComments.List(ctx, &opt2)
	})
	return r.comments, r.err
}

func (r *discussionCommentsConnectionResolver) Nodes(ctx context.Context) ([]*discussionCommentResolver, error) {
	comments, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	var l []*discussionCommentResolver
	for _, comment := range comments {
		l = append(l, &discussionCommentResolver{c: comment})
	}
	return l, nil
}

func (r *discussionCommentsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	withoutLimit := *r.opt
	withoutLimit.LimitOffset = nil
	count, err := db.DiscussionComments.Count(ctx, &withoutLimit)
	return int32(count), err
}

func (r *discussionCommentsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	comments, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && len(comments) > r.opt.Limit), nil
}
