package graphqlbackend

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/mattbaird/gochimp"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/notif"
)

// emailMentionPattern is a regex that matches an email mention in a comment, of
// the form "+alice@example.com". This is a simplified pattern that does not
// ensure the email is valid.
var emailMentionPattern = regexp.MustCompile(`\B\+[^\s]+@[^\s]+`)

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
	_, err := store.LocalRepos.Get(ctx, args.RemoteURI, args.AccessToken)
	if err != nil {
		return nil, err
	}

	thread, err := store.Threads.Get(ctx, int64(args.ThreadID))
	if err != nil {
		return nil, err
	}

	comment, err := store.Comments.Create(ctx, &sourcegraph.Comment{
		ThreadID:    args.ThreadID,
		Contents:    args.Contents,
		AuthorName:  args.AuthorName,
		AuthorEmail: args.AuthorEmail,
	})
	if err != nil {
		return nil, err
	}
	notifyCommentMentions(thread, comment)

	return &threadResolver{thread: thread}, nil
}

func notifyCommentMentions(thread *sourcegraph.Thread, comment *sourcegraph.Comment) {
	if notif.EmailIsConfigured() {
		const sendLimit = 50
		emails := parseEmailsFromComment(comment.Contents)
		if len(emails) > sendLimit {
			// Limit number of emails we send per comment to prevent spamming.
			emails = emails[:sendLimit]
		}
		for _, email := range emails {
			notif.SendMandrillTemplate("new-comment", "", email, "New comment from "+comment.AuthorName, []gochimp.Var{}, []gochimp.Var{
				gochimp.Var{Name: "AUTHOR", Content: comment.AuthorName},
				gochimp.Var{Name: "AUTHOR_EMAIL", Content: comment.AuthorEmail},
				gochimp.Var{Name: "FILENAME", Content: thread.File},
				gochimp.Var{Name: "PREVIEW", Content: comment.Contents},
			})
		}
	}
}

func parseEmailsFromComment(contents string) []string {
	matches := emailMentionPattern.FindAll([]byte(contents), -1)
	emails := []string{}
	added := map[string]interface{}{}

	for _, m := range matches {
		e := strings.TrimPrefix(string(m), "+")
		if _, ok := added[e]; !ok {
			emails = append(emails, e)
			added[e] = struct{}{}
		}
	}

	return emails
}
