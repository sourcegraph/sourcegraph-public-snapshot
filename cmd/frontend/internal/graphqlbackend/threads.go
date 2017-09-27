package graphqlbackend

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/tracking/slack"

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

// Deprecated root resolver.
func (r *rootResolver) Threads(ctx context.Context, args *struct {
	RemoteURI   string
	AccessToken string
	File        *string
	Limit       *int32
}) ([]*threadResolver, error) {
	threads := []*threadResolver{}
	repo, err := store.OrgRepos.GetByAccessToken(ctx, args.RemoteURI, args.AccessToken)
	if err == store.ErrRepoNotFound {
		// Datastore is lazily populated when comments are created
		// so it isn't an error for a repo to not exist yet.
		return threads, nil
	}
	if err != nil {
		return nil, err
	}

	limit := int32(1000)
	if args.Limit != nil && *args.Limit < limit {
		limit = *args.Limit
	}

	var ts []*sourcegraph.Thread
	if args.File != nil {
		ts, err = store.Threads.GetAllForFile(ctx, repo.ID, *args.File, limit)
	} else {
		ts, err = store.Threads.GetAllForRepo(ctx, repo.ID, limit)
	}
	if err != nil {
		return nil, err
	}

	for _, thread := range ts {
		threads = append(threads, &threadResolver{nil, repo, thread})
	}
	return threads, nil
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

func (*schemaResolver) CreateThread(ctx context.Context, args *struct {
	RemoteURI      string
	AccessToken    string
	File           string
	Revision       string
	StartLine      int32
	EndLine        int32
	StartCharacter int32
	EndCharacter   int32
	Contents       string
	AuthorName     string
	AuthorEmail    string
}) (*threadResolver, error) {
	actor := actor.FromContext(ctx)
	repo, err := store.OrgRepos.GetByAccessToken(ctx, args.RemoteURI, args.AccessToken)
	if err == store.ErrRepoNotFound {
		repo, err = store.OrgRepos.Create(ctx, &sourcegraph.OrgRepo{
			RemoteURI:   args.RemoteURI,
			AccessToken: args.AccessToken,
		})
	}
	if err != nil {
		return nil, err
	}

	newThread, err := store.Threads.Create(ctx, &sourcegraph.Thread{
		OrgRepoID:      repo.ID,
		File:           args.File,
		Revision:       args.Revision,
		StartLine:      args.StartLine,
		EndLine:        args.EndLine,
		StartCharacter: args.StartCharacter,
		EndCharacter:   args.EndCharacter,
	})
	if err != nil {
		return nil, err
	}

	comment, err := store.Comments.Create(ctx, newThread.ID, args.Contents, args.AuthorName, args.AuthorEmail, actor.UID)
	if err != nil {
		return nil, err
	}
	results := notifyThreadParticipants(repo, newThread, nil, comment, comment.AuthorName)
	err = slack.NotifyOnThread(args.AuthorName, args.AuthorEmail, fmt.Sprintf("%s (%d)", repo.RemoteURI, repo.ID), strings.Join(results.emails, ", "), results.commentURL)
	if err != nil {
		log15.Error("slack.NotifyOnThread failed", "error", err)
	}

	return &threadResolver{nil, repo, newThread}, nil
}

func (*schemaResolver) CreateThread2(ctx context.Context, args *struct {
	OrgID          int32
	RemoteURI      string
	File           string
	Revision       string
	StartLine      int32
	EndLine        int32
	StartCharacter int32
	EndCharacter   int32
	RangeLength    int32
	Contents       string
}) (*threadResolver, error) {
	// 🚨 SECURITY: verify that the current user is in the org.
	actor := actor.FromContext(ctx)
	member, err := store.OrgMembers.GetByOrgIDAndUserID(ctx, args.OrgID, actor.UID)
	fmt.Println(err)
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
	newThread, err := store.Threads.Create(ctx, &sourcegraph.Thread{
		OrgRepoID:      repo.ID,
		File:           args.File,
		Revision:       args.Revision,
		StartLine:      args.StartLine,
		EndLine:        args.EndLine,
		StartCharacter: args.StartCharacter,
		EndCharacter:   args.EndCharacter,
		RangeLength:    args.RangeLength,
	})
	if err != nil {
		return nil, err
	}

	comment, err := store.Comments.Create(ctx, newThread.ID, args.Contents, "", "", actor.UID)
	if err != nil {
		return nil, err
	}

	results := notifyThreadParticipants(repo, newThread, nil, comment, member.DisplayName)
	err = slack.NotifyOnThread(member.DisplayName, member.Email, fmt.Sprintf("%s (%d)", repo.RemoteURI, repo.ID), strings.Join(results.emails, ", "), results.commentURL)
	if err != nil {
		log15.Error("slack.NotifyOnThread failed", "error", err)
	}

	return &threadResolver{org, repo, newThread}, nil
}

func (*schemaResolver) UpdateThread(ctx context.Context, args *struct {
	RemoteURI   string
	AccessToken string
	ThreadID    int32
	Archived    *bool
}) (*threadResolver, error) {
	// 🚨 SECURITY: DO NOT REMOVE THIS CHECK! LocalRepos.Get is responsible for 🚨
	// ensuring the user has permissions to access the repository.
	repo, err := store.OrgRepos.GetByAccessToken(ctx, args.RemoteURI, args.AccessToken)
	if err != nil {
		return nil, err
	}

	thread, err := store.Threads.Update(ctx, args.ThreadID, repo.ID, args.Archived)
	if err != nil {
		return nil, err
	}
	return &threadResolver{nil, repo, thread}, nil
}

func (*schemaResolver) UpdateThread2(ctx context.Context, args *struct {
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
