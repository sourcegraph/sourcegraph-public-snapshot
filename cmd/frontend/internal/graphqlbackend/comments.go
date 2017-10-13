package graphqlbackend

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mattbaird/gochimp"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/slack"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/notif"
)

type commentResolver struct {
	org     *sourcegraph.Org
	repo    *sourcegraph.OrgRepo
	thread  *sourcegraph.Thread
	comment *sourcegraph.Comment
}

func (c *commentResolver) ID() int32 {
	return c.comment.ID
}

func (c *commentResolver) Contents() string {
	return c.comment.Contents
}

func (c *commentResolver) CreatedAt() string {
	return c.comment.CreatedAt.Format(time.RFC3339) // ISO
}

func (c *commentResolver) UpdatedAt() string {
	return c.comment.UpdatedAt.Format(time.RFC3339) // ISO
}

func (c *commentResolver) Author(ctx context.Context) (*orgMemberResolver, error) {
	member, err := store.OrgMembers.GetByOrgIDAndUserID(ctx, c.org.ID, c.comment.AuthorUserID)
	if err != nil {
		return nil, err
	}
	return &orgMemberResolver{c.org, member, nil}, nil
}

func (*schemaResolver) AddCommentToThread(ctx context.Context, args *struct {
	ThreadID int32
	Contents string
}) (*threadResolver, error) {
	thread, err := store.Threads.Get(ctx, args.ThreadID)
	if err != nil {
		return nil, err
	}

	repo, err := store.OrgRepos.GetByID(ctx, thread.OrgRepoID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: verify that the user is in the org.
	actor := actor.FromContext(ctx)
	_, err = store.OrgMembers.GetByOrgIDAndUserID(ctx, repo.OrgID, actor.UID)
	if err != nil {
		return nil, err
	}

	user, err := store.Users.GetByAuth0ID(actor.UID)
	if err != nil {
		return nil, err
	}

	org, err := store.Orgs.GetByID(ctx, repo.OrgID)
	if err != nil {
		return nil, err
	}

	// Query all comments so we can send a notification to all participants.
	comments, err := store.Comments.GetAllForThread(ctx, args.ThreadID)
	if err != nil {
		return nil, err
	}

	comment, err := store.Comments.Create(ctx, args.ThreadID, args.Contents, "", actor.Email, actor.UID)
	if err != nil {
		return nil, err
	}

	results, err := notifyNewComment(ctx, repo, thread, comments, comment, user.DisplayName)
	if err != nil {
		log15.Error("notifyNewComment failed", "error", err)
	}

	t := &threadResolver{org, repo, thread}

	title, err := t.Title(ctx)
	if err != nil {
		// errors swallowed because title is only needed for Slack notifications
		log15.Error("threadResolver.Title failed", "error", err)
	}
	if user, err := currentUser(ctx); err != nil {
		// errors swallowed because user is only needed for Slack notifications
		log15.Error("graphqlbackend.AddCommentToThread: currentUser failed", "error", err)
	} else {
		// TODO(Dan): replace sourcegraphOrgWebhookURL with any customer/org-defined webhook
		client := slack.New(org.SlackWebhookURL, true)
		go client.NotifyOnComment(user, org, repo, thread, comment, results.emails, getURL(repo, thread, "slack"), title)
	}

	return t, nil
}

func (*schemaResolver) ShareComment(ctx context.Context, args *struct {
	CommentID int32
}) (string, error) {
	comment, err := store.Comments.GetByID(ctx, args.CommentID)
	if err != nil {
		return "", err
	}

	thread, err := store.Threads.Get(ctx, comment.ThreadID)
	if err != nil {
		return "", err
	}

	repo, err := store.OrgRepos.GetByID(ctx, thread.OrgRepoID)
	if err != nil {
		return "", err
	}

	// 🚨 SECURITY: verify that the user is in the org.
	actor := actor.FromContext(ctx)
	_, err = store.OrgMembers.GetByOrgIDAndUserID(ctx, repo.OrgID, actor.UID)
	if err != nil {
		return "", err
	}
	return store.SharedItems.Create(ctx, &sourcegraph.SharedItem{
		AuthorUserID: actor.UID,
		CommentID:    &args.CommentID,
	})
}

type commentResults struct {
	emails     []string
	commentURL string
}

func notifyNewComment(ctx context.Context, repo *sourcegraph.OrgRepo, thread *sourcegraph.Thread, previousComments []*sourcegraph.Comment, comment *sourcegraph.Comment, commentAuthorName string) (*commentResults, error) {
	commentURL := getURL(repo, thread, "email")
	if !notif.EmailIsConfigured() {
		return &commentResults{emails: []string{}, commentURL: commentURL}, nil
	}

	var first *sourcegraph.Comment
	if len(previousComments) > 0 {
		first = previousComments[0]
	} else {
		first = comment
	}

	emails, err := allEmailsForOrg(ctx, repo.OrgID, []string{comment.AuthorUserID})
	if err != nil {
		return nil, err
	}
	repoName := repoNameFromURI(repo.RemoteURI)
	contents := strings.Replace(comment.Contents, "\n", "<br>", -1)
	for _, email := range emails {
		subject := fmt.Sprintf("[%s] %s (#%d)", repoName, titleFromContents(first.Contents), thread.ID)
		if len(previousComments) > 0 {
			subject = "Re: " + subject
		}
		config := &notif.EmailConfig{
			Template:  "new-comment",
			FromName:  commentAuthorName,
			FromEmail: "noreply@sourcegraph.com", // Remember to update this once we allow replies to these emails.
			ToName:    "",                        // We don't know names right now.
			ToEmail:   email,
			Subject:   subject,
		}

		notif.SendMandrillTemplate(config, []gochimp.Var{}, []gochimp.Var{
			gochimp.Var{Name: "CONTENTS", Content: contents},
			gochimp.Var{Name: "COMMENT_URL", Content: commentURL},
			gochimp.Var{Name: "LOCATION", Content: fmt.Sprintf("%s/%s:L%d", repoName, thread.File, thread.StartLine)},
		})
	}
	return &commentResults{emails: emails, commentURL: commentURL}, nil
}

func repoNameFromURI(remoteURI string) string {
	m := strings.SplitN(remoteURI, "/", 2)
	if len(m) < 2 {
		return remoteURI
	}
	return m[1]
}

func getURL(repo *sourcegraph.OrgRepo, thread *sourcegraph.Thread, utmSource string) string {
	aboutValues := url.Values{}
	if utmSource != "" {
		aboutValues.Set("utm_source", utmSource)
	}

	cloneURL := fmt.Sprintf("https://%s", repo.RemoteURI)
	values := url.Values{}
	values.Set("repo", cloneURL)
	values.Set("vcs", "git")
	values.Set("revision", thread.Revision)
	values.Set("path", thread.File)
	values.Set("thread", strconv.FormatInt(int64(thread.ID), 10))
	return fmt.Sprintf("https://about.sourcegraph.com/open/?%s#open?%s", aboutValues.Encode(), values.Encode())
}
