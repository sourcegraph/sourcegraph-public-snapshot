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
	"github.com/microcosm-cc/bluemonday"
	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/slack"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/notif"
)

type threadConnectionResolver struct {
	org                *sourcegraph.Org
	repos              []*sourcegraph.OrgRepo
	canonicalRemoteIDs []string
	file               *string
	branch             *string
	limit              *int32
}

func (t *threadConnectionResolver) orgRepoArgs() (orgID *int32, repoIDs []int32) {
	if t.org != nil {
		orgID = &t.org.ID
	}
	if len(t.repos) > 0 {
		for _, repo := range t.repos {
			repoIDs = append(repoIDs, repo.ID)
		}
		// repoIDs imply an orgID, avoid unnecessary join.
		orgID = nil
	} else if len(t.canonicalRemoteIDs) > 0 {
		// The query is for some repos but none of them exist.
		// This is not an error condition because we lazily populate org_repos.
		// Set an invalid repoID so no results are returned.
		repoIDs = []int32{-1}
	}
	return orgID, repoIDs
}

const maxLimit = 1000

func (t *threadConnectionResolver) Nodes(ctx context.Context) ([]*threadResolver, error) {
	limit := int32(maxLimit)
	if t.limit != nil && *t.limit < maxLimit {
		limit = *t.limit
	}
	orgID, repoIDs := t.orgRepoArgs()
	threads, err := store.Threads.List(ctx, orgID, repoIDs, t.branch, t.file, limit)
	if err != nil {
		return nil, err
	}
	repos := make(map[int32]*sourcegraph.OrgRepo)
	for _, repo := range t.repos {
		repos[repo.ID] = repo
	}
	resolvers := []*threadResolver{}
	for _, thread := range threads {
		repo := repos[thread.OrgRepoID]
		resolvers = append(resolvers, &threadResolver{t.org, repo, thread})
	}
	return resolvers, nil
}

func (t *threadConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	orgID, repoIDs := t.orgRepoArgs()
	return store.Threads.Count(ctx, orgID, repoIDs, t.branch, t.file)
}

type threadResolver struct {
	org    *sourcegraph.Org
	repo   *sourcegraph.OrgRepo
	thread *sourcegraph.Thread
}

func (t *threadResolver) ID() int32 {
	return t.thread.ID
}

func (t *threadResolver) Repo(ctx context.Context) (*orgRepoResolver, error) {
	var err error
	if t.repo == nil {
		t.repo, err = store.OrgRepos.GetByID(ctx, t.thread.OrgRepoID)
		if err != nil {
			return nil, err
		}
	}
	return &orgRepoResolver{t.org, t.repo}, nil
}

func (t *threadResolver) File() string {
	return t.thread.File
}

func (t *threadResolver) Branch() *string {
	return t.thread.Branch
}

func (t *threadResolver) RepoRevision() string {
	return t.thread.RepoRevision
}

func (t *threadResolver) LinesRevision() string {
	return t.thread.LinesRevision
}

func (t *threadResolver) StartLine() int32 {
	return t.thread.StartLine
}

func (t *threadResolver) EndLine() int32 {
	return t.thread.EndLine
}

func (t *threadResolver) StartCharacter() int32 {
	return t.thread.StartCharacter
}

func (t *threadResolver) EndCharacter() int32 {
	return t.thread.EndCharacter
}

func (t *threadResolver) RangeLength() int32 {
	return t.thread.RangeLength
}

func (t *threadResolver) CreatedAt() string {
	return t.thread.CreatedAt.Format(time.RFC3339) // ISO
}

func (t *threadResolver) Author(ctx context.Context) (*userResolver, error) {
	user, err := store.Users.GetByAuth0ID(ctx, t.thread.AuthorUserID)
	if err != nil {
		return nil, err
	}
	return &userResolver{actor: actor.FromContext(ctx), user: user}, nil
}

func (t *threadResolver) Lines() *threadLineResolver {
	if t.thread.Lines == nil {
		return nil
	}
	return &threadLineResolver{t.thread.Lines}
}

type threadLineResolver struct {
	*sourcegraph.ThreadLines
}

func (t *threadLineResolver) HTMLBefore(args *struct {
	IsLightTheme bool
}) string {
	return sanitize(colorSwap(t.ThreadLines.HTMLBefore, args.IsLightTheme))
}

func (t *threadLineResolver) HTML(args *struct {
	IsLightTheme bool
}) string {
	return sanitize(colorSwap(t.ThreadLines.HTML, args.IsLightTheme))
}

func (t *threadLineResolver) HTMLAfter(args *struct {
	IsLightTheme bool
}) string {
	return sanitize(colorSwap(t.ThreadLines.HTMLAfter, args.IsLightTheme))
}

var (
	// Matches exactly a string "color: #aaaaaa;" and NOTHING else. "aaaaaa"
	// can be any alphanumeric (upper or lowercase) characters.
	//
	// Be *VERY* careful modifying this, as it matching anything but the above
	// would introduce XSS vulnerabilies.
	onlyColorStyle = regexp.MustCompile(`^color: #[[:alnum:]]{6};$`)
)

func sanitize(html string) string {
	policy := bluemonday.UGCPolicy()
	policy.AllowAttrs("style").Matching(onlyColorStyle).OnElements("span")
	return policy.Sanitize(html)
}

// colorSwap takes some pre-highlighted HTML from the editor and replaces
// colors to match our theme.
//
// TODO(slimsag): We currently assume the editor will only produce HTML in one
// theme color.
func colorSwap(html string, isLightTheme bool) string {
	if !isLightTheme {
		return html
	}

	// lightTheme is a map of old colors to new (light theme) colors.
	lightTheme := map[string]string{
		"#ffa8a8": "#2aa198",
		"#72c3fc": "#657b83",
		"#fff3bf": "#268bd2",
		"#2b8a3e": "#93a1a1",
		"#c9d4e3": "#657b83",
		"#d4d4d4": "#859900",
		"#268bd2": "#b58900",
		"#d3f9d8": "#6c71c4",
		"#ffc078": "#fc8e00",
		"#f2f4f8": "#39496a"}
	for oldColor, newColor := range lightTheme {
		oldColor = fmt.Sprintf("color: %s;", oldColor)
		newColor = fmt.Sprintf("color: %s;", newColor)
		html = strings.Replace(html, oldColor, newColor, -1)
	}
	return html
}

func (t *threadLineResolver) TextBefore() string {
	return t.ThreadLines.TextBefore
}

func (t *threadLineResolver) Text() string {
	return t.ThreadLines.Text
}

func (t *threadLineResolver) TextAfter() string {
	return t.ThreadLines.TextAfter
}

func (t *threadLineResolver) TextSelectionRangeStart() int32 {
	return t.ThreadLines.TextSelectionRangeStart
}

func (t *threadLineResolver) TextSelectionRangeLength() int32 {
	return t.ThreadLines.TextSelectionRangeLength
}

func (t *threadResolver) ArchivedAt() *string {
	if t.thread.ArchivedAt == nil {
		return nil
	}
	a := t.thread.ArchivedAt.Format(time.RFC3339) // ISO
	return &a
}

func (t *threadResolver) Title(ctx context.Context) (string, error) {
	cs, err := t.Comments(ctx)
	if err != nil {
		return "", err
	}
	if len(cs) == 0 {
		return "", nil
	}
	return TitleFromContents(cs[0].Contents()), nil
}

func (t *threadResolver) Comments(ctx context.Context) ([]*commentResolver, error) {
	comments, err := store.Comments.GetAllForThread(ctx, t.thread.ID)
	if err != nil {
		return nil, err
	}
	commentResolvers := []*commentResolver{}
	for _, comment := range comments {
		commentResolvers = append(commentResolvers, &commentResolver{t.org, t.repo, t.thread, comment})
	}
	return commentResolvers, nil
}

type threadLines struct {
	HTMLBefore, HTML, HTMLAfter string
	TextBefore, Text, TextAfter string
	TextSelectionRangeStart     int32
	TextSelectionRangeLength    int32
}

// orgInt32OrID allows a GraphQL resolver arg to be specified either as
// an ID! or an Int!.
type orgInt32OrID struct {
	int32Value int32
}

func (orgInt32OrID) ImplementsGraphQLType(name string) bool {
	return name == "ID" || name == "Int"
}

func (v *orgInt32OrID) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case string: // graphql.ID
		var int32Value int32
		id := graphql.ID(input)
		if err := relay.UnmarshalSpec(id, &int32Value); err != nil {
			return err
		}
		*v = orgInt32OrID{int32Value: int32Value}
		return nil
	case int32:
		*v = orgInt32OrID{int32Value: input}
		return nil
	default:
		return fmt.Errorf("wrong type")
	}
}

func (s *schemaResolver) CreateThread(ctx context.Context, args *struct {
	OrgID             orgInt32OrID // accept int32 and org graphql.ID
	CanonicalRemoteID string
	CloneURL          string
	File              string
	RepoRevision      string
	LinesRevision     string
	Branch            *string
	StartLine         int32
	EndLine           int32
	StartCharacter    int32
	EndCharacter      int32
	RangeLength       int32
	Contents          string
	Lines             *threadLines
}) (*threadResolver, error) {
	// 🚨 SECURITY: verify that the current user is in the org.
	actor := actor.FromContext(ctx)
	_, err := store.OrgMembers.GetByOrgIDAndUserID(ctx, args.OrgID.int32Value, actor.UID)
	if err != nil {
		return nil, err
	}

	repo, err := store.OrgRepos.GetByCanonicalRemoteID(ctx, args.OrgID.int32Value, args.CanonicalRemoteID)
	if err == store.ErrRepoNotFound {
		repo, err = store.OrgRepos.Create(ctx, &sourcegraph.OrgRepo{
			CanonicalRemoteID: args.CanonicalRemoteID,
			CloneURL:          args.CloneURL,
			OrgID:             args.OrgID.int32Value,
		})
	}
	if err != nil {
		return nil, err
	}

	org, err := store.Orgs.GetByID(ctx, repo.OrgID)
	if err != nil {
		return nil, err
	}

	// TODO(nick): transaction
	thread := &sourcegraph.Thread{
		OrgRepoID:      repo.ID,
		File:           args.File,
		RepoRevision:   args.RepoRevision,
		LinesRevision:  args.LinesRevision,
		Branch:         args.Branch,
		StartLine:      args.StartLine,
		EndLine:        args.EndLine,
		StartCharacter: args.StartCharacter,
		EndCharacter:   args.EndCharacter,
		RangeLength:    args.RangeLength,
		AuthorUserID:   actor.UID,
	}
	if args.Lines != nil {
		thread.Lines = &sourcegraph.ThreadLines{
			HTMLBefore:               args.Lines.HTMLBefore,
			HTML:                     args.Lines.HTML,
			HTMLAfter:                args.Lines.HTMLAfter,
			TextBefore:               args.Lines.TextBefore,
			Text:                     args.Lines.Text,
			TextAfter:                args.Lines.TextAfter,
			TextSelectionRangeStart:  args.Lines.TextSelectionRangeStart,
			TextSelectionRangeLength: args.Lines.TextSelectionRangeLength,
		}
	}
	newThread, err := store.Threads.Create(ctx, thread)
	if err != nil {
		return nil, err
	}

	if args.Contents != "" {
		comment, err := store.Comments.Create(ctx, newThread.ID, args.Contents, "", actor.Email, actor.UID)
		if err != nil {
			return nil, err
		}
		var results *commentResults
		err = func() error {
			user, err := store.Users.GetByCurrentAuthUser(ctx)
			if err != nil {
				return err
			}
			results, err = s.notifyNewComment(ctx, *repo, *newThread, nil, *comment, *user, *org)
			if err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			log15.Error("notifyNewComment failed", "error", err)
		}
		if uResolver, err := currentUser(ctx); err != nil {
			// errors swallowed because user is only needed for Slack notifications
			log15.Error("graphqlbackend.CreateThread: currentUser failed", "error", err)
		} else if results != nil {
			// TODO(Dan): replace sourcegraphOrgWebhookURL with any customer/org-defined webhook
			client := slack.New(org.SlackWebhookURL, true)
			commentURL, err := s.getURL(ctx, newThread.ID, &comment.ID, "slack")
			if err != nil {
				log15.Error("graphqlbackend.CreateThread: getURL failed", "error", err)
			} else {
				go client.NotifyOnThread(uResolver, org, repo, newThread, comment, results.emails, commentURL.String())
			}
		}
	} /* else {
		// Creating a thread without Contents (a comment) means it is a code
		// snippet without any user comment.
		// TODO(dan): slack notifications for this case
	}*/
	return &threadResolver{org, repo, newThread}, nil
}

func (s *schemaResolver) UpdateThread(ctx context.Context, args *struct {
	ThreadID int32
	Archived *bool
}) (*threadResolver, error) {
	thread, err := store.Threads.Get(ctx, args.ThreadID)
	if err != nil {
		return nil, err
	}

	repo, err := store.OrgRepos.GetByID(ctx, thread.OrgRepoID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: verify that the current user is in the org.
	actor := actor.FromContext(ctx)
	_, err = store.OrgMembers.GetByOrgIDAndUserID(ctx, repo.OrgID, actor.UID)
	if err != nil {
		return nil, err
	}

	org, err := store.Orgs.GetByID(ctx, repo.OrgID)
	if err != nil {
		return nil, err
	}

	wasArchived := thread.ArchivedAt
	thread, err = store.Threads.Update(ctx, args.ThreadID, repo.ID, args.Archived)
	if err != nil {
		return nil, err
	}

	if wasArchived == nil && thread.ArchivedAt != nil {
		user, err := store.Users.GetByAuth0ID(ctx, actor.UID)
		if err != nil {
			return nil, err
		}
		comments, err := store.Comments.GetAllForThread(ctx, args.ThreadID)
		if err != nil {
			return nil, err
		}
		s.utilNotifyThreadArchived(ctx, *repo, *thread, comments, *user)
	}

	return &threadResolver{org, repo, thread}, nil
}

func (s *schemaResolver) ShareThread(ctx context.Context, args *struct {
	ThreadID int32
}) (string, error) {
	u, err := s.shareThreadInternal(ctx, args.ThreadID, true)
	if err != nil {
		return "", nil
	}

	// TODO(slimsag): future: move this to the client in case we ever have
	// other users of this method.
	q := u.Query()
	q.Set("utm_source", "share-snippet-editor")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// TODO(slimsag): expose the public boolean as a graphql parameter and remove this internal function call
func (*schemaResolver) shareThreadInternal(ctx context.Context, threadID int32, public bool) (*url.URL, error) {
	thread, err := store.Threads.Get(ctx, threadID)
	if err != nil {
		return nil, err
	}

	repo, err := store.OrgRepos.GetByID(ctx, thread.OrgRepoID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: verify that the current user is in the org.
	actor := actor.FromContext(ctx)
	_, err = store.OrgMembers.GetByOrgIDAndUserID(ctx, repo.OrgID, actor.UID)
	if err != nil {
		return nil, err
	}
	return store.SharedItems.Create(ctx, &sourcegraph.SharedItem{
		AuthorUserID: actor.UID,
		Public:       public,
		ThreadID:     &threadID,
	})
}

func (s *schemaResolver) utilNotifyThreadArchived(ctx context.Context, repo sourcegraph.OrgRepo, thread sourcegraph.Thread, previousComments []*sourcegraph.Comment, archiver sourcegraph.User) error {
	url, err := s.getURL(ctx, thread.ID, nil, "email")
	if err != nil {
		return err
	}
	if !notif.EmailIsConfigured() {
		return nil
	}

	var first *sourcegraph.Comment
	if len(previousComments) > 0 {
		first = previousComments[0]
	}

	org, err := store.Orgs.GetByID(ctx, repo.OrgID)
	if err != nil {
		return err
	}
	emails, err := emailsToNotify(ctx, previousComments, archiver, *org)
	if err != nil {
		return err
	}

	repoName := repoNameFromRemoteID(repo.CanonicalRemoteID)
	for _, email := range emails {
		var subject string
		if first != nil {
			var branch string
			if thread.Branch != nil {
				branch = "@" + *thread.Branch
			}
			subject = fmt.Sprintf("[%s%s] %s (#%d)", repoName, branch, TitleFromContents(first.Contents), thread.ID)
		}
		if len(previousComments) > 0 {
			subject = "Re: " + subject
		}
		config := &notif.EmailConfig{
			Template:  "thread-archived",
			FromName:  archiver.DisplayName,
			FromEmail: "noreply@sourcegraph.com", // Remember to update this once we allow replies to these emails.
			ToName:    "",                        // We don't know names right now.
			ToEmail:   email,
			Subject:   subject,
		}

		notif.SendMandrillTemplate(config, []gochimp.Var{}, []gochimp.Var{
			gochimp.Var{Name: "THREAD_ID", Content: strconv.Itoa(int(thread.ID))},
			gochimp.Var{Name: "THREAD_URL", Content: url.String()},
		})
	}
	return nil
}

// TitleFromContents returns a title based on the first sentence or line of the content.
func TitleFromContents(contents string) string {
	matchEndpoint := regexp.MustCompile(`[.!?]\s`)
	if idxs := matchEndpoint.FindStringSubmatchIndex(contents); len(idxs) > 0 {
		contents = contents[:idxs[0]+1]
	}
	if i := strings.Index(contents, "\n"); i != -1 {
		contents = contents[:i]
	}
	contents = strings.TrimSpace(contents)
	if len(contents) > 140 {
		contents = contents[:137] + "..."
	}
	return contents
}
