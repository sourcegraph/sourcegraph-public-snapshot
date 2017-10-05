package graphqlbackend

import (
	"context"
	"regexp"
	"strings"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/slack"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

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

func (t *threadResolver) Revision() string {
	return t.thread.Revision
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
	user, err := store.Users.GetByAuth0ID(t.thread.AuthorUserID)
	if err != nil {
		return nil, err
	}
	return &userResolver{actor: actor.FromContext(ctx), user: user}, nil
}

func (t *threadResolver) Lines() *threadLineResolver {
	return &threadLineResolver{t.thread}
}

type threadLineResolver struct {
	*sourcegraph.Thread
}

func (t *threadLineResolver) HTMLBefore() string {
	return t.Lines.HTMLBefore
}

func (t *threadLineResolver) HTML() string {
	return t.Lines.HTML
}

func (t *threadLineResolver) HTMLAfter() string {
	return t.Lines.HTMLAfter
}

func (t *threadLineResolver) TextBefore() string {
	return t.Lines.TextBefore
}

func (t *threadLineResolver) Text() string {
	return t.Lines.Text
}

func (t *threadLineResolver) TextAfter() string {
	return t.Lines.TextAfter
}

func (t *threadLineResolver) TextSelectionRangeStart() int32 {
	return t.Lines.TextSelectionRangeStart
}

func (t *threadLineResolver) TextSelectionRangeLength() int32 {
	return t.Lines.TextSelectionRangeLength
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
	return titleFromContents(cs[0].Contents()), nil
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
	OrgID          int32
	RemoteURI      string
	File           string
	Revision       string
	Branch         *string
	StartLine      int32
	EndLine        int32
	StartCharacter int32
	EndCharacter   int32
	RangeLength    int32
	Contents       string
	Lines          *threadLines
}) (*threadResolver, error) {
	// 🚨 SECURITY: verify that the current user is in the org.
	actor := actor.FromContext(ctx)
	member, err := store.OrgMembers.GetByOrgIDAndUserID(ctx, args.OrgID, actor.UID)
	if err != nil {
		return nil, err
	}

	repo, err := store.OrgRepos.GetByRemoteURI(ctx, args.OrgID, args.RemoteURI)
	if err == store.ErrRepoNotFound {
		repo, err = store.OrgRepos.Create(ctx, &sourcegraph.OrgRepo{
			RemoteURI: args.RemoteURI,
			OrgID:     args.OrgID,
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
		Revision:       args.Revision,
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

		results := notifyAllInOrg(ctx, repo, newThread, nil, comment, member.DisplayName)
		if user, err := currentUser(ctx); err != nil {
			// errors swallowed because user is only needed for Slack notifications
			log15.Error("graphqlbackend.CreateThread: currentUser failed", "error", err)
		} else {
			// TODO(Dan): replace sourcegraphOrgWebhookURL with any customer/org-defined webhook
			client := slack.New(sourcegraphOrgWebhookURL)
			go client.NotifyOnThread(user, org, repo, newThread, comment, results.emails, getURL(repo, newThread, comment, "slack"))
		}
	} /* else {
		// Creating a thread without Contents (a comment) means it is a code
		// snippet without any user comment.

		// TODO(dan): slack notifications for this case
	}*/
	return &threadResolver{org, repo, newThread}, nil
}

func (*schemaResolver) UpdateThread(ctx context.Context, args *struct {
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

	thread, err = store.Threads.Update(ctx, args.ThreadID, repo.ID, args.Archived)
	if err != nil {
		return nil, err
	}
	// TODO: send email notification that thread has been archived?
	return &threadResolver{org, repo, thread}, nil
}

// titleFromContents returns a title based on the first sentence or line of the content.
func titleFromContents(contents string) string {
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
