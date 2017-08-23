package graphqlbackend

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/mattbaird/gochimp"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/notif"
)

type commentResolver struct {
	comment *sourcegraph.Comment
}

func (c *commentResolver) ID() int32 {
	return c.comment.ID
}

func (c *commentResolver) Contents() string {
	return c.comment.Contents
}

func (c *commentResolver) AuthorName() string {
	return c.comment.AuthorName
}

func (c *commentResolver) AuthorEmail() string {
	return c.comment.AuthorEmail
}

func (c *commentResolver) CreatedAt() string {
	return c.comment.CreatedAt.Format(time.RFC3339) // ISO
}

func (c *commentResolver) UpdatedAt() string {
	return c.comment.UpdatedAt.Format(time.RFC3339) // ISO
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
	repo, err := store.LocalRepos.Get(ctx, args.RemoteURI, args.AccessToken)
	if err != nil {
		return nil, err
	}

	thread, err := store.Threads.Get(ctx, int64(args.ThreadID))
	if err != nil {
		return nil, err
	}

	// Query all comments so we can send a notification to all participants.
	comments, err := store.Comments.GetAllForThread(ctx, int64(args.ThreadID))
	if err != nil {
		return nil, err
	}

	comment, err := store.Comments.Create(ctx, args.ThreadID, args.Contents, args.AuthorName, args.AuthorEmail)
	if err != nil {
		return nil, err
	}
	notifyThreadParticipants(repo, thread, comments, comment)

	return &threadResolver{thread: thread}, nil
}

// notifyThreadParticipants sends email notifications to the participants in the comment thread.
func notifyThreadParticipants(repo *sourcegraph.LocalRepo, thread *sourcegraph.Thread, previousComments []*sourcegraph.Comment, comment *sourcegraph.Comment) {
	if !notif.EmailIsConfigured() {
		return
	}
	emails := emailsToNotify(previousComments, comment)
	for _, email := range emails {
		subject := fmt.Sprintf("[%s] %s (#%d)", repo.RemoteURI, path.Base(thread.File), thread.ID)
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
			gochimp.Var{Name: "AUTHOR", Content: comment.AuthorName},
			gochimp.Var{Name: "AUTHOR_EMAIL", Content: comment.AuthorEmail},
			gochimp.Var{Name: "FILENAME", Content: thread.File},
			gochimp.Var{Name: "PREVIEW", Content: comment.Contents},
			gochimp.Var{Name: "COMMENT_URL", Content: getURL(repo, thread, comment)},
		})
	}
}

func getURL(repo *sourcegraph.LocalRepo, thread *sourcegraph.Thread, comment *sourcegraph.Comment) string {
	return fmt.Sprintf("https://about.sourcegraph.com/open-native/#open?resource=repo://%s/%s?threadID=%d&commentID=%d", repo.RemoteURI, thread.File, thread.ID, comment.ID)
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
