package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mattbaird/gochimp"
	"github.com/microcosm-cc/bluemonday"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/slack"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/notif"
)

type threadConnectionResolver struct {
	org    *sourcegraph.Org
	repo   *sourcegraph.OrgRepo
	file   *string
	branch *string
	limit  *int32
}

func (t *threadConnectionResolver) orgRepoArgs() (orgID *int32, repoID *int32) {
	if t.org != nil {
		orgID = &t.org.ID
	}
	if t.repo != nil {
		repoID = &t.repo.ID
		// repoID implies an orgID, avoid unnecessary join.
		orgID = nil
	}
	return orgID, repoID
}

const maxLimit = 1000

func (t *threadConnectionResolver) Nodes(ctx context.Context) ([]*threadResolver, error) {
	limit := int32(maxLimit)
	if t.limit != nil && *t.limit < maxLimit {
		limit = *t.limit
	}
	orgID, repoID := t.orgRepoArgs()
	threads, err := store.Threads.List(ctx, repoID, orgID, t.branch, t.file, limit)
	if err != nil {
		return nil, err
	}
	resolvers := []*threadResolver{}
	for _, thread := range threads {
		resolvers = append(resolvers, &threadResolver{t.org, t.repo, thread})
	}
	return resolvers, nil
}

func (t *threadConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	orgID, repoID := t.orgRepoArgs()
	return store.Threads.Count(ctx, repoID, orgID, t.branch, t.file)
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

func (t *threadResolver) Revision() string {
	return t.thread.RepoRevision // Deprecated. Using new repoRevision field data.
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

func (t *threadLineResolver) HTMLBefore() string {
	return sanitize(t.ThreadLines.HTMLBefore)
}

func (t *threadLineResolver) HTML() string {
	return sanitize(t.ThreadLines.HTML)
}

func (t *threadLineResolver) HTMLAfter() string {
	return sanitize(t.ThreadLines.HTMLAfter)
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

func (s *schemaResolver) CreateThread(ctx context.Context, args *struct {
	OrgID             int32
	RemoteURI         *string // DEPRECATED: to be replaced by CanonicalRemoteID.
	CanonicalRemoteID *string
	CloneURL          *string
	File              string
	RepoRevision      *string
	LinesRevision     *string
	Revision          *string
	Branch            *string
	StartLine         int32
	EndLine           int32
	StartCharacter    int32
	EndCharacter      int32
	RangeLength       int32
	Contents          string
	Lines             *threadLines
}) (*threadResolver, error) {
	// Sort out the revision args. This is temporary until args.Revision is phased out.
	if args.RepoRevision == nil && args.LinesRevision == nil {
		if args.Revision == nil {
			return nil, errors.New("no revision specified")
		}
		args.RepoRevision, args.LinesRevision = args.Revision, args.Revision
	} else if args.RepoRevision == nil || args.LinesRevision == nil {
		return nil, errors.New("both repoRevision and linesRevision required")
	}

	// 🚨 SECURITY: verify that the current user is in the org.
	actor := actor.FromContext(ctx)
	_, err := store.OrgMembers.GetByOrgIDAndUserID(ctx, args.OrgID, actor.UID)
	if err != nil {
		return nil, err
	}

	if args.RemoteURI == nil && args.CanonicalRemoteID == nil {
		return nil, errors.New("canonicalRemoteID required")
	}
	var cloneURL string
	if args.RemoteURI != nil {
		args.CanonicalRemoteID = args.RemoteURI
		// Convert remoteURI into the correct vendor-specific format.
		//
		// TODO client will be expected to supply CanonicalRemoteID in
		// the correct format and this logic can be removed.
		if strings.HasPrefix(*args.RemoteURI, "github.com") {
			cloneURL = "github://" + *args.RemoteURI
		} else if strings.HasPrefix(*args.RemoteURI, "bitbucket.org") {
			cloneURL = "bitbucketcloud://" + *args.RemoteURI
		} else {
			cloneURL = "https://" + *args.RemoteURI
		}
	} else {
		if args.CloneURL == nil {
			return nil, errors.New("cloneURL required")
		}
		cloneURL = *args.CloneURL
	}
	repo, err := store.OrgRepos.GetByCanonicalRemoteID(ctx, args.OrgID, *args.CanonicalRemoteID)
	if err == store.ErrRepoNotFound {
		repo, err = store.OrgRepos.Create(ctx, &sourcegraph.OrgRepo{
			CanonicalRemoteID: *args.CanonicalRemoteID,
			CloneURL:          cloneURL,
			OrgID:             args.OrgID,
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
		RepoRevision:   *args.RepoRevision,
		LinesRevision:  *args.LinesRevision,
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
		} else {
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
			gochimp.Var{Name: "THREAD_URL", Content: url},
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
