package graphqlbackend

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/slack"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/txemail"
)

type threadConnectionResolver struct {
	org                *types.Org
	repos              []*types.OrgRepo
	canonicalRemoteIDs []api.RepoURI
	file               *string
	branch             *string
	limit              *int32
}

func (t *threadConnectionResolver) orgRepoArgs() (orgID *int32, repos []api.RepoID) {
	if t.org != nil {
		orgID = &t.org.ID
	}
	if len(t.repos) > 0 {
		for _, repo := range t.repos {
			repos = append(repos, repo.ID)
		}
		// repos imply an orgID, avoid unnecessary join.
		orgID = nil
	} else if len(t.canonicalRemoteIDs) > 0 {
		// The query is for some repos but none of them exist.
		// This is not an error condition because we lazily populate org_repos.
		// Set an invalid repo so no results are returned.
		repos = []api.RepoID{-1}
	}
	return orgID, repos
}

const maxLimit = 1000

func (t *threadConnectionResolver) Nodes(ctx context.Context) ([]*threadResolver, error) {
	limit := int32(maxLimit)
	if t.limit != nil && *t.limit < maxLimit {
		limit = *t.limit
	}
	orgID, repoIDs := t.orgRepoArgs()
	threads, err := db.Threads.ListByFile(ctx, orgID, repoIDs, t.branch, t.file, limit)
	if err != nil {
		return nil, err
	}
	repos := make(map[api.RepoID]*types.OrgRepo)
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
	orgID, repos := t.orgRepoArgs()
	count, err := db.Threads.CountByFile(ctx, orgID, repos, t.branch, t.file)
	return int32(count), err
}

type threadResolver struct {
	org    *types.Org
	repo   *types.OrgRepo
	thread *types.Thread
}

func (t *threadResolver) Repo(ctx context.Context) (*orgRepoResolver, error) {
	var err error
	if t.repo == nil {
		t.repo, err = db.OrgRepos.GetByID(ctx, t.thread.OrgRepoID)
		if err != nil {
			return nil, err
		}
	}
	return &orgRepoResolver{t.org, t.repo}, nil
}

// TODO(nick): deprecated
func (t *threadResolver) File() string {
	return t.thread.RepoRevisionPath
}

func (t *threadResolver) RepoRevisionPath() string {
	return t.thread.RepoRevisionPath
}

func (t *threadResolver) LinesRevisionPath() string {
	return t.thread.LinesRevisionPath
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
	user, err := db.Users.GetByID(ctx, t.thread.AuthorUserID)
	if err != nil {
		return nil, err
	}
	return &userResolver{user: user}, nil
}

func (t *threadResolver) Lines() *threadLineResolver {
	if t.thread.Lines == nil {
		return nil
	}
	return &threadLineResolver{t.thread.Lines}
}

type threadLineResolver struct {
	*types.ThreadLines
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
		"#ffa8a8": "#f08c00",
		"#72c3fc": "#495057",
		"#fff3bf": "#1862ab",
		"#2b8a3e": "#adb5bd",
		"#c9d4e3": "#657b83",
		"#d4d4d4": "#859900",
		"#d3f9d8": "#6c71c4",
		"#ffc078": "#fc8e00",
		"#f2f4f8": "#39496a",
		"#87cefa": "#7443ad",
		"#4ec9b0": "#20c997"}
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
	comments, err := db.Comments.GetAllForThread(ctx, t.thread.ID)
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

type orgID struct {
	int32Value int32
}

func (orgID) ImplementsGraphQLType(name string) bool {
	return name == "ID"
}

func (v *orgID) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case string:
		var int32Value int32
		id := graphql.ID(input)
		if kind := relay.UnmarshalKind(id); kind != "Org" {
			return fmt.Errorf("expected id with kind Org; got %s", kind)
		}
		if err := relay.UnmarshalSpec(id, &int32Value); err != nil {
			return err
		}
		*v = orgID{int32Value: int32Value}
		return nil
	default:
		return fmt.Errorf("wrong type")
	}
}

// threadID can accept either a GraphQL ID! or Int! value. It is used to back-compatibly
// migrate from thread Int!-type IDs to ID!-type IDs.
type threadID struct {
	int32Value int32
}

func (threadID) ImplementsGraphQLType(name string) bool {
	return name == "ID"
}

func (v *threadID) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case string:
		var int32Value int32
		id := graphql.ID(input)
		if kind := relay.UnmarshalKind(id); kind != "Thread" {
			return fmt.Errorf("expected id with kind Thread; got %s", kind)
		}
		if err := relay.UnmarshalSpec(id, &int32Value); err != nil {
			return err
		}
		*v = threadID{int32Value: int32Value}
		return nil
	default:
		return fmt.Errorf("wrong type")
	}
}

type createThreadInput struct {
	OrgID             orgID // accept int32 and org graphql.ID
	CanonicalRemoteID string
	CloneURL          string
	RepoRevisionPath  string
	LinesRevisionPath string
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
}

func (s *schemaResolver) CreateThread(ctx context.Context, args *struct {
	Input *createThreadInput
}) (*threadResolver, error) {
	return s.createThreadInput(ctx, args.Input)
}

// DEPRECATED
func (s *schemaResolver) CreateThread2(ctx context.Context, args *struct {
	Input *createThreadInput
}) (*threadResolver, error) {
	return s.createThreadInput(ctx, args.Input)
}

func (s *schemaResolver) createThreadInput(ctx context.Context, args *createThreadInput) (*threadResolver, error) {
	// 🚨 SECURITY: Check that the current user is a member of the org.
	if err := backend.CheckCurrentUserIsOrgMember(ctx, args.OrgID.int32Value); err != nil {
		return nil, err
	}

	currentUser, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	email, _, err := db.UserEmails.GetEmail(ctx, currentUser.SourcegraphID())
	if err != nil {
		return nil, err
	}

	repo, err := db.OrgRepos.GetByCanonicalRemoteID(ctx, args.OrgID.int32Value, api.RepoURI(args.CanonicalRemoteID))
	if errcode.IsNotFound(err) {
		repo, err = db.OrgRepos.Create(ctx, &types.OrgRepo{
			CanonicalRemoteID: api.RepoURI(args.CanonicalRemoteID),
			CloneURL:          args.CloneURL,
			OrgID:             args.OrgID.int32Value,
		})
	}
	if err != nil {
		return nil, err
	}

	org, err := db.Orgs.GetByID(ctx, repo.OrgID)
	if err != nil {
		return nil, err
	}

	// TODO(nick): transaction
	thread := &types.Thread{
		OrgRepoID:         repo.ID,
		RepoRevisionPath:  args.RepoRevisionPath,
		LinesRevisionPath: args.LinesRevisionPath,
		RepoRevision:      args.RepoRevision,
		LinesRevision:     args.LinesRevision,
		Branch:            args.Branch,
		StartLine:         args.StartLine,
		EndLine:           args.EndLine,
		StartCharacter:    args.StartCharacter,
		EndCharacter:      args.EndCharacter,
		RangeLength:       args.RangeLength,
		AuthorUserID:      currentUser.SourcegraphID(),
	}
	if args.Lines != nil {
		thread.Lines = &types.ThreadLines{
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
	newThread, err := db.Threads.Create(ctx, thread)
	if err != nil {
		return nil, err
	}

	if args.Contents != "" {
		comment, err := db.Comments.Create(ctx, newThread.ID, args.Contents, currentUser.SourcegraphID())
		if err != nil {
			return nil, err
		}
		var results *commentResults
		err = func() error {
			user, err := db.Users.GetByCurrentAuthUser(ctx)
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
		if results != nil {
			// TODO(Dan): replace sourcegraphOrgWebhookURL with any customer/org-defined webhook
			client := slack.New(org.SlackWebhookURL, true)
			commentURL := threadURL(newThread.ID, &comment.ID, "slack")
			go slack.NotifyOnThread(client, currentUser, email, org, repo, newThread, comment, results.emails, commentURL.String())
		}
	} /* else {
		// Creating a thread without Contents (a comment) means it is a code
		// snippet without any user comment.
		// TODO(dan): slack notifications for this case
	}*/
	return &threadResolver{org, repo, newThread}, nil
}

func (s *schemaResolver) UpdateThread(ctx context.Context, args *struct {
	ThreadID threadID
	Archived *bool
}) (*threadResolver, error) {
	thread, err := db.Threads.Get(ctx, args.ThreadID.int32Value)
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

	org, err := db.Orgs.GetByID(ctx, repo.OrgID)
	if err != nil {
		return nil, err
	}

	wasArchived := thread.ArchivedAt
	thread, err = db.Threads.Update(ctx, args.ThreadID.int32Value, repo.ID, args.Archived)
	if err != nil {
		return nil, err
	}

	if wasArchived == nil && thread.ArchivedAt != nil {
		user, err := db.Users.GetByCurrentAuthUser(ctx)
		if err != nil {
			return nil, err
		}
		comments, err := db.Comments.GetAllForThread(ctx, args.ThreadID.int32Value)
		if err != nil {
			return nil, err
		}
		if err := s.utilNotifyThreadArchived(ctx, *repo, *thread, comments, *user); err != nil {
			return nil, err
		}
	}

	return &threadResolver{org, repo, thread}, nil
}

func (s *schemaResolver) ShareThread(ctx context.Context, args *struct {
	ThreadID threadID
}) (string, error) {
	u, err := s.shareThreadInternal(ctx, args.ThreadID.int32Value, true)
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
	thread, err := db.Threads.Get(ctx, threadID)
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

	return db.SharedItems.Create(ctx, &types.SharedItem{
		AuthorUserID: currentUser.ID,
		Public:       public,
		ThreadID:     &threadID,
	})
}

func (s *schemaResolver) utilNotifyThreadArchived(ctx context.Context, repo types.OrgRepo, thread types.Thread, previousComments []*types.Comment, archiver types.User) error {
	url := threadURL(thread.ID, nil, "email")

	var first *types.Comment
	if len(previousComments) > 0 {
		first = previousComments[0]
	}

	org, err := db.Orgs.GetByID(ctx, repo.OrgID)
	if err != nil {
		return err
	}
	emails, err := emailsToNotify(ctx, previousComments, archiver, *org)
	if err != nil {
		return err
	}

	// Send tx emails asynchronously.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		repoName := repoNameFromRemoteID(repo.CanonicalRemoteID)
		for _, email := range emails {
			var branch string
			if thread.Branch != nil {
				branch = "@" + *thread.Branch
			}
			var title string
			if first != nil {
				title = TitleFromContents(first.Contents)
			}
			if err := txemail.Send(ctx, txemail.Message{
				FromName: archiver.DisplayName,
				To:       []string{email},
				Template: threadArchivedEmailTemplates,
				Data: threadEmailTemplateCommonData{
					Reply:    len(previousComments) > 0,
					RepoName: repoName,
					Branch:   branch,
					Title:    title,
					Number:   thread.ID,
					URL:      url.String(),
				},
			}); err != nil {
				log15.Error("error sending archived-thread email", "to", email, "err", err)
			}
		}
	}()

	return nil
}

type threadEmailTemplateCommonData struct {
	Reply    bool
	RepoName string
	Branch   string
	Title    string
	Number   int32
	URL      string
}

const threadEmailSubjectTemplate = `{{if .Reply}}Re: {{end}}[{{.RepoName}}{{.Branch}}] {{.Title}} (#{{.Number}})`

var (
	threadArchivedEmailTemplates = txemail.MustParseTemplate(txemail.Templates{
		Subject: threadEmailSubjectTemplate,
		Text: `
Archived #{{.Number}}

View discussion on Sourcegraph: {{.URL}}
`,
		HTML: `
<p>Archived <a href="{{.URL}}">#{{.Number}}</a></p>
`,
	})
)

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
