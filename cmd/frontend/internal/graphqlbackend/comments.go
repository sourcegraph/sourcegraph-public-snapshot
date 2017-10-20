package graphqlbackend

import (
	"context"
	"fmt"
	"html"
	"net/url"
	"regexp"
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

func (c *commentResolver) Title(ctx context.Context) string {
	return TitleFromContents(c.comment.Contents)
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

func (c *commentResolver) Author(ctx context.Context) (*userResolver, error) {
	user, err := store.Users.GetByAuth0ID(c.comment.AuthorUserID)
	if err != nil {
		return nil, err
	}
	return &userResolver{user, actor.FromContext(ctx)}, nil
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

	user, err := store.Users.GetByCurrentAuthUser(ctx)
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

	results, err := notifyNewComment(ctx, *repo, *thread, comments, *comment, *user, *org)
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
		go client.NotifyOnComment(user, org, repo, thread, comment, results.emails, getURL(*repo, *thread, "slack"), title)
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

func notifyNewComment(ctx context.Context, repo sourcegraph.OrgRepo, thread sourcegraph.Thread, previousComments []*sourcegraph.Comment, comment sourcegraph.Comment, author sourcegraph.User, org sourcegraph.Org) (*commentResults, error) {
	commentURL := getURL(repo, thread, "email")
	if !notif.EmailIsConfigured() {
		return &commentResults{emails: []string{}, commentURL: commentURL}, nil
	}

	var first sourcegraph.Comment
	if len(previousComments) > 0 {
		first = *previousComments[0]
	} else {
		first = comment
	}

	emails, err := emailsToNotify(ctx, append(previousComments, &comment), author, org)
	if err != nil {
		return nil, err
	}

	repoName := repoNameFromURI(repo.RemoteURI)
	contents := strings.Replace(html.EscapeString(comment.Contents), "\n", "<br>", -1)
	mentions := usernamesFromMentions(comment.Contents)
	for _, m := range mentions {
		contents = strings.Replace(contents, "@"+m, `<b>@`+m+`</b>`, -1)
	}
	lineVars := []gochimp.Var{}
	if len(previousComments) == 0 && thread.Lines != nil {
		lines := thread.Lines.TextBefore + thread.Lines.Text
		lineVars = []gochimp.Var{
			gochimp.Var{Name: "CONTEXT_LINES", Content: html.EscapeString(lines)},
		}
	}
	for _, email := range emails {
		var branch string
		if thread.Branch != nil {
			branch = "@" + *thread.Branch
		}
		subject := fmt.Sprintf("[%s%s] %s (#%d)", repoName, branch, TitleFromContents(first.Contents), thread.ID)
		if len(previousComments) > 0 {
			subject = "Re: " + subject
		}
		config := &notif.EmailConfig{
			Template:  "new-comment",
			FromName:  author.DisplayName,
			FromEmail: "noreply@sourcegraph.com", // Remember to update this once we allow replies to these emails.
			ToName:    "",                        // We don't know names right now.
			ToEmail:   email,
			Subject:   subject,
		}

		notif.SendMandrillTemplate(config, []gochimp.Var{}, append([]gochimp.Var{
			gochimp.Var{Name: "CONTENTS", Content: contents},
			gochimp.Var{Name: "COMMENT_URL", Content: commentURL},
			gochimp.Var{Name: "LOCATION", Content: fmt.Sprintf("%s/%s:L%d", repoName, thread.File, thread.StartLine)},
		}, lineVars...))
	}
	return &commentResults{emails: emails, commentURL: commentURL}, nil
}

// emailsToNotify returns all emails that should be notified of activity given a list of comments.
func emailsToNotify(ctx context.Context, comments []*sourcegraph.Comment, author sourcegraph.User, org sourcegraph.Org) ([]string, error) {
	uniqueParticipants, uniqueMentions := map[string]struct{}{}, map[string]struct{}{}

	// Prevent the author from being mentioned
	uniqueParticipants[author.Auth0ID] = struct{}{}
	uniqueMentions[author.Username] = struct{}{}

	var participantIDs, mentions []string
	for i, c := range comments {
		// Notify participants in the conversation.
		participantIDs = appendUnique(uniqueParticipants, participantIDs, c.AuthorUserID)

		// Notify mentioned people in the conversation.
		usernames := usernamesFromMentions(c.Contents)

		// If first comment contains no mentions, notify entire org.
		if i == 0 && len(usernames) == 0 {
			usernames = []string{org.Name}
		}

		// Allow the user to mention themself in their latest comment.
		if i == len(comments)-1 {
			delete(uniqueMentions, author.Username)
		}
		mentions = appendUnique(uniqueMentions, mentions, usernames...)
	}

	_, selfMentioned := uniqueMentions[author.Username]
	_, atOrgMention := uniqueMentions["org"]
	_, orgNameMention := uniqueMentions[org.Name]

	if atOrgMention || orgNameMention {
		var exclude []string
		if !selfMentioned {
			exclude = []string{author.Auth0ID}
		}
		return allEmailsForOrg(ctx, org.ID, exclude)
	}

	users, err := store.Users.ListByOrg(ctx, org.ID, participantIDs, mentions)
	if err != nil {
		return nil, err
	}
	emails, uniqueEmails := []string{}, map[string]struct{}{}
	for _, u := range users {
		emails = appendUnique(uniqueEmails, emails, u.Email)
	}
	return emails, nil
}

var usernameMentionPattern = regexp.MustCompile("@" + store.UsernamePattern)

// usernamesFromMentions extracts usernames that are mentioned using a @username
// syntax within a comment.
func usernamesFromMentions(contents string) []string {
	matches := usernameMentionPattern.FindAll([]byte(contents), -1)
	var usernames []string
	for _, m := range matches {
		usernames = append(usernames, strings.TrimPrefix(string(m), "@"))
	}
	return usernames
}

func repoNameFromURI(remoteURI string) string {
	m := strings.SplitN(remoteURI, "/", 2)
	if len(m) < 2 {
		return remoteURI
	}
	return m[1]
}

func getURL(repo sourcegraph.OrgRepo, thread sourcegraph.Thread, utmSource string) string {
	aboutValues := url.Values{}
	if utmSource != "" {
		aboutValues.Set("utm_source", utmSource)
	}

	cloneURL := fmt.Sprintf("https://%s", repo.RemoteURI)
	values := url.Values{}
	values.Set("repo", cloneURL)
	values.Set("vcs", "git")
	values.Set("revision", thread.RepoRevision)
	values.Set("path", thread.File)
	values.Set("thread", strconv.FormatInt(int64(thread.ID), 10))
	return fmt.Sprintf("https://about.sourcegraph.com/open/?%s#open?%s", aboutValues.Encode(), values.Encode())
}

// appendUnique returns value appended to values if value is not a key in unique.
func appendUnique(unique map[string]struct{}, slice []string, values ...string) []string {
	for _, v := range values {
		if _, ok := unique[v]; ok {
			continue
		}
		unique[v] = struct{}{}
		slice = append(slice, v)
	}
	return slice
}
