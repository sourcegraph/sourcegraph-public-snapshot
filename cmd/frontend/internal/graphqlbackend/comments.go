package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/url"
	"regexp"
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

func (c *commentResolver) RichHTML() (string, error) {
	return renderMarkdown(c.comment.Contents)
}

func (c *commentResolver) CreatedAt() string {
	return c.comment.CreatedAt.Format(time.RFC3339) // ISO
}

func (c *commentResolver) UpdatedAt() string {
	return c.comment.UpdatedAt.Format(time.RFC3339) // ISO
}

func (c *commentResolver) Author(ctx context.Context) (*userResolver, error) {
	user, err := store.Users.GetByAuth0ID(ctx, c.comment.AuthorUserID)
	if err != nil {
		return nil, err
	}
	return &userResolver{user}, nil
}

func (s *schemaResolver) AddCommentToThreadShared(ctx context.Context, args *struct {
	ThreadID int32
	Contents string
	ULID     string
}) (*threadResolver, error) {
	return s.addCommentToThread(ctx, args)
}

func (s *schemaResolver) AddCommentToThread(ctx context.Context, args *struct {
	ThreadID int32
	Contents string
}) (*threadResolver, error) {
	return s.addCommentToThread(ctx, &struct {
		ThreadID int32
		Contents string
		ULID     string
	}{
		args.ThreadID,
		args.Contents,
		"",
	})
}

// TODO(slimsag): expose only one addCommentToThread in the future (with
// nullable ULID string).
func (s *schemaResolver) addCommentToThread(ctx context.Context, args *struct {
	ThreadID int32
	Contents string
	ULID     string
}) (*threadResolver, error) {
	thread, err := store.Threads.Get(ctx, args.ThreadID)
	if err != nil {
		return nil, err
	}

	repo, err := store.OrgRepos.GetByID(ctx, thread.OrgRepoID)
	if err != nil {
		return nil, err
	}

	actor := actor.FromContext(ctx)
	if args.ULID == "" {
		// Plain case (adding a comment to a thread from the editor).
		// 🚨 SECURITY: verify that the user is in the org.
		_, err = store.OrgMembers.GetByOrgIDAndUserID(ctx, repo.OrgID, actor.UID)
		if err != nil {
			return nil, err
		}
	} else {
		// Web case (adding a comment to a thread from a shared URL).
		if !actor.IsAuthenticated() {
			// They must be signed in to comment from the web.
			return nil, errors.New("must be authenticated")
		}

		// 🚨 SECURITY: If the shared item is public, anyone can add comments
		// as long as the ULID is real. If the shared item is not public, only
		// org members can.
		item, err := store.SharedItems.Get(ctx, args.ULID)
		if err != nil {
			return nil, err
		}
		if !item.Public {
			// 🚨 SECURITY: verify that the user is in the org.
			_, err = store.OrgMembers.GetByOrgIDAndUserID(ctx, repo.OrgID, actor.UID)
			if err != nil {
				return nil, err
			}
		}
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

	results, err := s.notifyNewComment(ctx, *repo, *thread, comments, *comment, *user, *org)
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
	} else if results != nil {
		// TODO(Dan): replace sourcegraphOrgWebhookURL with any customer/org-defined webhook
		client := slack.New(org.SlackWebhookURL, true)
		commentURL, err := s.getURL(ctx, thread.ID, &comment.ID, "slack")
		if err != nil {
			log15.Error("graphqlbackend.AddCommentToThread: getURL failed", "error", err)
		} else {
			go client.NotifyOnComment(user, org, repo, thread, comment, results.emails, commentURL.String(), title)
		}
	}

	return t, nil
}

func (s *schemaResolver) ShareComment(ctx context.Context, args *struct {
	CommentID int32
}) (string, error) {
	u, err := s.shareCommentInternal(ctx, args.CommentID, true)
	if err != nil {
		return "", nil
	}

	// TODO(slimsag): future: move this to the client in case we ever have
	// other users of this method.
	q := u.Query()
	q.Set("utm_source", "share-comment-editor")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// TODO(slimsag): expose the public boolean as a graphql parameter and remove this internal function call
func (*schemaResolver) shareCommentInternal(ctx context.Context, commentID int32, public bool) (*url.URL, error) {
	comment, err := store.Comments.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	thread, err := store.Threads.Get(ctx, comment.ThreadID)
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
	return store.SharedItems.Create(ctx, &sourcegraph.SharedItem{
		AuthorUserID: actor.UID,
		Public:       public,
		ThreadID:     &thread.ID,
		CommentID:    &commentID,
	})
}

type commentResults struct {
	emails     []string
	commentURL string
}

func (s *schemaResolver) notifyNewComment(ctx context.Context, repo sourcegraph.OrgRepo, thread sourcegraph.Thread, previousComments []*sourcegraph.Comment, comment sourcegraph.Comment, author sourcegraph.User, org sourcegraph.Org) (*commentResults, error) {
	commentURL, err := s.getURL(ctx, thread.ID, &comment.ID, "email")
	if err != nil {
		return nil, err
	}
	if !notif.EmailIsConfigured() {
		return &commentResults{emails: []string{}, commentURL: commentURL.String()}, nil
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

	repoName := repoNameFromRemoteID(repo.CanonicalRemoteID)
	contents := strings.Replace(html.EscapeString(comment.Contents), "\n", "<br>", -1)
	mentions := usernamesFromMentions(comment.Contents)
	for _, m := range mentions {
		contents = strings.Replace(contents, "@"+m, `<b>@`+m+`</b>`, -1)
	}
	lineVars := []gochimp.Var{}
	if len(previousComments) == 0 && thread.Lines != nil {
		lines := strings.Join([]string{thread.Lines.TextBefore, thread.Lines.Text}, "\n")
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
			gochimp.Var{Name: "COMMENT_URL", Content: commentURL.String()},
			gochimp.Var{Name: "LOCATION", Content: fmt.Sprintf("%s/%s:L%d", repoName, thread.RepoRevisionPath, thread.StartLine)},
		}, lineVars...))
	}
	return &commentResults{emails: emails, commentURL: commentURL.String()}, nil
}

// emailsToNotify returns all emails that should be notified of activity given a list of comments.
func emailsToNotify(ctx context.Context, comments []*sourcegraph.Comment, author sourcegraph.User, org sourcegraph.Org) ([]string, error) {
	uniqueParticipants, uniqueMentions := map[string]struct{}{}, map[string]struct{}{}
	orgName, authorUsername := strings.ToLower(org.Name), strings.ToLower(author.Username)

	// Prevent the author from being mentioned
	uniqueParticipants[author.Auth0ID] = struct{}{}
	uniqueMentions[authorUsername] = struct{}{}

	var participantIDs, mentions []string
	for i, c := range comments {
		// Notify participants in the conversation.
		participantIDs = appendUnique(uniqueParticipants, participantIDs, c.AuthorUserID)

		// Notify mentioned people in the conversation.
		usernames := usernamesFromMentions(c.Contents)
		// Normalize usernames for case-insensitivity.
		for i, u := range usernames {
			usernames[i] = strings.ToLower(u)
		}

		// Allow the user to mention themself in their latest comment.
		if i == len(comments)-1 {
			delete(uniqueMentions, authorUsername)
		}
		mentions = appendUnique(uniqueMentions, mentions, usernames...)
	}

	_, selfMentioned := uniqueMentions[authorUsername]
	_, atOrgMention := uniqueMentions["org"]
	_, orgNameMention := uniqueMentions[orgName]

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

var usernameMentionPattern = regexp.MustCompile(`\B@` + store.UsernamePattern)

// usernamesFromMentions extracts usernames that are mentioned using a @username
// syntax within a comment. Mentions are normalized to lowercase format.
func usernamesFromMentions(contents string) []string {
	matches := usernameMentionPattern.FindAll([]byte(contents), -1)
	var usernames []string
	for _, m := range matches {
		usernames = append(usernames, strings.TrimPrefix(string(m), "@"))
	}
	return usernames
}

func repoNameFromRemoteID(canonicalRemoteID string) string {
	m := strings.SplitN(canonicalRemoteID, "/", 2)
	if len(m) < 2 {
		return canonicalRemoteID
	}
	return m[1]
}

func (s *schemaResolver) getURL(ctx context.Context, threadID int32, commentID *int32, utmSource string) (*url.URL, error) {
	var (
		url *url.URL
		err error
	)
	if commentID != nil {
		url, err = s.shareCommentInternal(ctx, *commentID, false)
	} else {
		url, err = s.shareThreadInternal(ctx, threadID, false)
	}
	if err != nil {
		return nil, err
	}
	return withUTMSource(url, utmSource), nil
}

func withUTMSource(u *url.URL, utmSource string) *url.URL {
	q := u.Query()
	q.Set("utm_source", utmSource)
	u.RawQuery = q.Encode()
	return u
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
