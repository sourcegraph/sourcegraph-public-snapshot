package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/dlclark/regexp2"
	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/slack"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/txemail"
)

type commentResolver struct {
	org     *sourcegraph.Org
	repo    *sourcegraph.OrgRepo
	thread  *sourcegraph.Thread
	comment *sourcegraph.Comment
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
	user, err := db.Users.GetByID(ctx, c.comment.AuthorUserID)
	if err != nil {
		return nil, err
	}
	return &userResolver{user}, nil
}

func (s *schemaResolver) AddCommentToThreadShared(ctx context.Context, args *struct {
	ThreadID threadID
	Contents string
	ULID     string
}) (*threadResolver, error) {
	return s.addCommentToThread(ctx, args)
}

func (s *schemaResolver) AddCommentToThread(ctx context.Context, args *struct {
	ThreadID threadID
	Contents string
}) (*threadResolver, error) {
	return s.addCommentToThread(ctx, &struct {
		ThreadID threadID
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
	ThreadID threadID
	Contents string
	ULID     string
}) (*threadResolver, error) {
	thread, err := db.Threads.Get(ctx, args.ThreadID.int32Value)
	if err != nil {
		return nil, err
	}

	repo, err := db.OrgRepos.GetByID(ctx, thread.OrgRepoID)
	if err != nil {
		return nil, err
	}

	actor := actor.FromContext(ctx)
	if args.ULID == "" {
		// Plain case (adding a comment to a thread from the editor).
		// 🚨 SECURITY: Check that the current user is a member of the org.
		if err := backend.CheckCurrentUserIsOrgMember(ctx, repo.OrgID); err != nil {
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
		item, err := db.SharedItems.Get(ctx, args.ULID)
		if err != nil {
			return nil, err
		}
		if !item.Public {
			// 🚨 SECURITY: Check that the current user is a member of the org.
			if err := backend.CheckCurrentUserIsOrgMember(ctx, repo.OrgID); err != nil {
				return nil, err
			}
		}
	}

	user, err := db.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}
	email, _, err := db.UserEmails.GetEmail(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	org, err := db.Orgs.GetByID(ctx, repo.OrgID)
	if err != nil {
		return nil, err
	}

	// Query all comments so we can send a notification to all participants.
	comments, err := db.Comments.GetAllForThread(ctx, args.ThreadID.int32Value)
	if err != nil {
		return nil, err
	}

	comment, err := db.Comments.Create(ctx, args.ThreadID.int32Value, args.Contents, user.ID)
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
		commentURL := threadURL(thread.ID, &comment.ID, "slack")
		go client.NotifyOnComment(user, email, org, repo, thread, comment, results.emails, commentURL.String(), title)
	}

	return t, nil
}

// commentID can accept either a GraphQL ID! or Int! value. It is used to back-compatibly
// migrate from comment Int!-type IDs to ID!-type IDs.
type commentID struct {
	int32Value int32
}

func (commentID) ImplementsGraphQLType(name string) bool {
	return name == "ID"
}

func (v *commentID) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case string:
		var int32Value int32
		id := graphql.ID(input)
		if kind := relay.UnmarshalKind(id); kind != "Comment" {
			return fmt.Errorf("expected id with kind Comment; got %s", kind)
		}
		if err := relay.UnmarshalSpec(id, &int32Value); err != nil {
			return err
		}
		*v = commentID{int32Value: int32Value}
		return nil
	default:
		return fmt.Errorf("wrong type")
	}
}

func (s *schemaResolver) ShareComment(ctx context.Context, args *struct {
	CommentID commentID
}) (string, error) {
	u, err := s.shareCommentInternal(ctx, args.CommentID.int32Value, true)
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
	comment, err := db.Comments.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	thread, err := db.Threads.Get(ctx, comment.ThreadID)
	if err != nil {
		return nil, err
	}

	repo, err := db.OrgRepos.GetByID(ctx, thread.OrgRepoID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check that the current user is a member of the org.
	if err := backend.CheckCurrentUserIsOrgMember(ctx, repo.OrgID); err != nil {
		return nil, err
	}

	currentUser, err := db.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	return db.SharedItems.Create(ctx, &sourcegraph.SharedItem{
		AuthorUserID: currentUser.ID,
		Public:       public,
		ThreadID:     &thread.ID,
		CommentID:    &commentID,
	})
}

type commentResults struct {
	emails     []string
	commentURL string
}

var mockNotifyNewComment func() (*commentResults, error)

func (s *schemaResolver) notifyNewComment(ctx context.Context, repo sourcegraph.OrgRepo, thread sourcegraph.Thread, previousComments []*sourcegraph.Comment, comment sourcegraph.Comment, author sourcegraph.User, org sourcegraph.Org) (*commentResults, error) {
	if mockNotifyNewComment != nil {
		return mockNotifyNewComment()
	}

	commentURL := threadURL(thread.ID, &comment.ID, "email")
	if !conf.CanSendEmail() {
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
	var lines string
	if len(previousComments) == 0 && thread.Lines != nil {
		lines = strings.Join([]string{thread.Lines.TextBefore, thread.Lines.Text}, "\n")
	}

	// Send tx emails asynchronously.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		for _, email := range emails {
			var branch string
			if thread.Branch != nil {
				branch = "@" + *thread.Branch
			}
			location := fmt.Sprintf("%s/%s:L%d", repoName, thread.RepoRevisionPath, thread.StartLine)
			if err := txemail.Send(ctx, txemail.Message{
				FromName: author.DisplayName,
				To:       []string{email},
				Template: newCommentEmailTemplates,
				Data: struct {
					threadEmailTemplateCommonData
					Location     string
					ContextLines string
					Contents     string
				}{
					threadEmailTemplateCommonData: threadEmailTemplateCommonData{
						Reply:    len(previousComments) > 0,
						RepoName: repoName,
						Branch:   branch,
						Title:    TitleFromContents(first.Contents),
						Number:   thread.ID,
						URL:      commentURL.String(),
					},
					Location:     location,
					ContextLines: lines,
					Contents:     contents,
				},
			}); err != nil {
				log15.Error("error sending new-comment notifications", "to", email, "err", err)
			}
		}
	}()

	return &commentResults{emails: emails, commentURL: commentURL.String()}, nil
}

var (
	newCommentEmailTemplates = txemail.MustParseTemplate(txemail.Templates{
		Subject: threadEmailSubjectTemplate,
		Text: `
{{.Location}}

{{if .ContextLines}}
------------------------------------------------------------------------------
{{.ContextLines}}
------------------------------------------------------------------------------
{{end}}

{{.Contents}}

View discussion on Sourcegraph: {{.URL}}
`,
		HTML: `
{{if .ContextLines}}
<pre style="color:#555">{{.ContextLines}}</pre>
{{end}}

<p>{{.Contents}}</p>

<p>View discussion on Sourcegraph: <a href="{{.URL}}">{{.Location}}</a></p>
`,
	})
)

var mockEmailsToNotify func(ctx context.Context, comments []*sourcegraph.Comment, author sourcegraph.User, org sourcegraph.Org) ([]string, error)

// emailsToNotify returns all emails that should be notified of activity given a list of comments.
func emailsToNotify(ctx context.Context, comments []*sourcegraph.Comment, author sourcegraph.User, org sourcegraph.Org) ([]string, error) {
	if mockEmailsToNotify != nil {
		return mockEmailsToNotify(ctx, comments, author, org)
	}

	uniqueParticipants, uniqueMentions := map[int32]struct{}{}, map[string]struct{}{}
	orgName, authorUsername := strings.ToLower(org.Name), strings.ToLower(author.Username)

	// Prevent the author from being mentioned
	uniqueParticipants[author.ID] = struct{}{}
	uniqueMentions[authorUsername] = struct{}{}

	var participantIDs []int32
	var mentions []string
	for i, c := range comments {
		author, err := db.Users.GetByID(ctx, c.AuthorUserID)
		if err != nil {
			return nil, err
		}

		// Notify participants in the conversation.
		participantIDs = appendUniqueInt32(uniqueParticipants, participantIDs, author.ID)

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
		var exclude []int32
		if !selfMentioned {
			exclude = []int32{author.ID}
		}
		return allEmailsForOrg(ctx, org.ID, exclude)
	}

	users, err := db.Users.ListByOrg(ctx, org.ID, participantIDs, mentions)
	if err != nil {
		return nil, err
	}
	emails, uniqueEmails := []string{}, map[string]struct{}{}
	for _, u := range users {
		email, _, err := db.UserEmails.GetEmail(ctx, u.ID)
		if err != nil {
			return nil, err
		}

		emails = appendUnique(uniqueEmails, emails, email)
	}
	return emails, nil
}

var usernameMentionPattern = regexp2.MustCompile(`\B@`+db.UsernamePattern, 0)

// usernamesFromMentions extracts usernames that are mentioned using a @username
// syntax within a comment. Mentions are normalized to lowercase format.
func usernamesFromMentions(contents string) []string {
	var matches []string
	m, _ := usernameMentionPattern.FindStringMatch(contents)
	for m != nil {
		matches = append(matches, m.String())
		m, _ = usernameMentionPattern.FindNextMatch(m)
	}
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

// threadURL returns the URL to a thread (scrolled to an optional comment).
func threadURL(threadID int32, commentID *int32, utmSource string) *url.URL {
	u := globals.AppURL.ResolveReference(&url.URL{Path: path.Join("threads", string(marshalThreadID(threadID)))})
	if commentID != nil {
		q := u.Query()
		q.Set("id", fmt.Sprint(*commentID))
		u.RawQuery = q.Encode()
	}
	if utmSource != "" {
		u = withUTMSource(u, utmSource)
	}
	return u
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

// appendUniqueInt32 returns value appended to values if value is not a key in unique.
func appendUniqueInt32(unique map[int32]struct{}, slice []int32, values ...int32) []int32 {
	for _, v := range values {
		if _, ok := unique[v]; ok {
			continue
		}
		unique[v] = struct{}{}
		slice = append(slice, v)
	}
	return slice
}
