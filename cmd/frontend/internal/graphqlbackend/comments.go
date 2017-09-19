package graphqlbackend

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mattbaird/gochimp"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/tracking/slack"
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

// deprecated
func (c *commentResolver) AuthorName() string {
	return c.comment.AuthorName
}

// deprecated
func (c *commentResolver) AuthorEmail() string {
	return c.comment.AuthorEmail
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
	return &orgMemberResolver{c.org, member}, nil
}

func (*schemaResolver) AddCommentToThread(ctx context.Context, args *struct {
	RemoteURI   string
	AccessToken string
	ThreadID    int32
	Contents    string
	AuthorName  string
	AuthorEmail string
}) (*threadResolver, error) {
	// 🚨 SECURITY: DO NOT REMOVE THIS CHECK! LocalRepos.Get is responsible for 🚨
	// ensuring the user has permissions to access the repository.
	repo, err := store.OrgRepos.Get(ctx, args.RemoteURI, args.AccessToken)
	if err != nil {
		return nil, err
	}

	thread, err := store.Threads.Get(ctx, args.ThreadID)
	if err != nil {
		return nil, err
	}

	// Query all comments so we can send a notification to all participants.
	comments, err := store.Comments.GetAllForThread(ctx, args.ThreadID)
	if err != nil {
		return nil, err
	}

	actor := actor.FromContext(ctx)
	comment, err := store.Comments.Create(ctx, args.ThreadID, args.Contents, args.AuthorName, args.AuthorEmail, actor.UID)
	if err != nil {
		return nil, err
	}
	emails := notifyThreadParticipants(repo, thread, comments, comment)
	err = slack.NotifyOnComment(args.AuthorName, args.AuthorEmail, fmt.Sprintf("%s (%d)", repo.RemoteURI, repo.ID), strings.Join(emails, ", "))
	if err != nil {
		log15.Error("slack.NotifyOnComment failed", "error", err)
	}

	return &threadResolver{nil, repo, thread}, nil
}

func (*schemaResolver) AddCommentToThread2(ctx context.Context, args *struct {
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
	member, err := store.OrgMembers.GetByOrgIDAndUserID(ctx, repo.OrgID, actor.UID)
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

	comment, err := store.Comments.Create(ctx, args.ThreadID, args.Contents, "", "", actor.UID)
	if err != nil {
		return nil, err
	}

	emails := notifyThreadParticipants(repo, thread, comments, comment)
	err = slack.NotifyOnComment(member.DisplayName, member.Email, fmt.Sprintf("%s (%d)", repo.RemoteURI, repo.ID), strings.Join(emails, ", "))
	if err != nil {
		log15.Error("slack.NotifyOnComment failed", "error", err)
	}
	return &threadResolver{org, repo, thread}, nil
}

// notifyThreadParticipants sends email notifications to the participants in the comment thread.
func notifyThreadParticipants(repo *sourcegraph.OrgRepo, thread *sourcegraph.Thread, previousComments []*sourcegraph.Comment, comment *sourcegraph.Comment) []string {
	if !notif.EmailIsConfigured() {
		return []string{}
	}

	var first *sourcegraph.Comment
	if len(previousComments) > 0 {
		first = previousComments[0]
	} else {
		first = comment
	}
	emails := emailsToNotify(previousComments, comment)
	repoName := repoNameFromURI(repo.RemoteURI)
	contents := strings.Replace(comment.Contents, "\n", "<br>", -1)
	for _, email := range emails {
		subject := fmt.Sprintf("[%s] %s (#%d)", repoName, titleFromContents(first.Contents), thread.ID)
		if len(previousComments) > 0 {
			subject = "Re: " + subject
		}
		config := &notif.EmailConfig{
			Template:  "new-comment",
			FromName:  comment.AuthorName,
			FromEmail: "noreply@sourcegraph.com", // Remember to update this once we allow replies to these emails.
			ToName:    "",                        // We don't know names right now.
			ToEmail:   email,
			Subject:   subject,
		}

		notif.SendMandrillTemplate(config, []gochimp.Var{}, []gochimp.Var{
			gochimp.Var{Name: "CONTENTS", Content: contents},
			gochimp.Var{Name: "COMMENT_URL", Content: getURL(repo, thread, comment)},
			gochimp.Var{Name: "LOCATION", Content: fmt.Sprintf("%s/%s:L%d", repoName, thread.File, thread.StartLine)},
		})
	}
	return emails
}

func repoNameFromURI(remoteURI string) string {
	m := strings.SplitN(remoteURI, "/", 2)
	if len(m) < 2 {
		return remoteURI
	}
	return m[1]
}

func getURL(repo *sourcegraph.OrgRepo, thread *sourcegraph.Thread, comment *sourcegraph.Comment) string {
	cloneURL := fmt.Sprintf("https://%s", repo.RemoteURI)
	values := url.Values{}
	values.Set("repo", cloneURL)
	values.Set("vcs", "git")
	values.Set("revision", thread.Revision)
	values.Set("path", thread.File)
	values.Set("thread", strconv.FormatInt(int64(thread.ID), 10))
	return fmt.Sprintf("https://about.sourcegraph.com/open-native/#open?%s", values.Encode())
}

// maxEmails is a limit on the number of email notifications
// that we will send per comment to mitigate potential spam abuse.
const maxEmails = 50

// emailMentionPattern is a regex that matches an email mention in a comment, of
// the form "+alice@example.com". This is a simplified pattern that does not
// ensure the email is valid.
//
// TODO: will not match emails with non-alphanumeric TLDs (e.g. user@foo.みんな)
var emailMentionPattern = regexp.MustCompile(`\B\+[^\s]+@[^\s]+\.[A-Za-z0-9]+`)

// emailsToNotify returns all emails that should be notified of the new comment in the thread of previous comments.
func emailsToNotify(previousComments []*sourcegraph.Comment, newComment *sourcegraph.Comment) []string {
	unique := map[string]struct{}{}
	var emails []string

	// Notify everyone already in the conversation, except for the author of the new comment.
	for _, c := range previousComments {
		if c.AuthorEmail != newComment.AuthorEmail {
			emails = appendUnique(unique, emails, c.AuthorEmail)
		}
		emails = appendUniqueEmailsFromMentions(unique, emails, c.Contents, newComment.AuthorEmail)
	}

	// Notify all mentions in the new comment (including the original author if they mentioned themself).
	emails = appendUniqueEmailsFromMentions(unique, emails, newComment.Contents, "")

	if len(emails) > maxEmails {
		emails = emails[:maxEmails]
	}
	return emails
}

// appendUniqueEmailsFromMentions parses email mentions from comment and returns
// the ones not already in unique appended to emails.
func appendUniqueEmailsFromMentions(unique map[string]struct{}, emails []string, comment, excludeAuthor string) []string {
	matches := emailMentionPattern.FindAll([]byte(comment), -1)
	for _, m := range matches {
		email := strings.TrimPrefix(string(m), "+")
		if email != excludeAuthor {
			emails = appendUnique(unique, emails, email)
		}
	}
	return emails
}

// appendUnique returns value appended to values if value is not a key in unique.
func appendUnique(unique map[string]struct{}, values []string, value string) []string {
	if _, ok := unique[value]; ok {
		return values
	}
	unique[value] = struct{}{}
	return append(values, value)
}
