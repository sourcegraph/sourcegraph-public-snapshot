package graphqlbackend

import (
	"context"
	"regexp"
	"strings"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

type threadResolver struct {
	thread *sourcegraph.Thread
}

func (t *threadResolver) ID() int32 {
	return t.thread.ID
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

func (r *rootResolver) Threads(ctx context.Context, args *struct {
	RemoteURI   string
	AccessToken string
	File        *string
}) ([]*threadResolver, error) {
	threads := []*threadResolver{}

	repo, err := store.LocalRepos.Get(ctx, args.RemoteURI, args.AccessToken)
	if err == store.ErrRepoNotFound {
		// Datastore is lazily populated when comments are created
		// so it isn't an error for a repo to not exist yet.
		return threads, nil
	}
	if err != nil {
		return nil, err
	}

	var ts []*sourcegraph.Thread
	if args.File != nil {
		ts, err = store.Threads.GetAllForFile(ctx, int64(repo.ID), *args.File)
	} else {
		ts, err = store.Threads.GetAllForRepo(ctx, int64(repo.ID))
	}
	if err != nil {
		return nil, err
	}

	for _, t := range ts {
		threads = append(threads, &threadResolver{thread: t})
	}
	return threads, nil
}

func (t *threadResolver) Comments(ctx context.Context) ([]*commentResolver, error) {
	cs, err := store.Comments.GetAllForThread(ctx, int64(t.thread.ID))
	if err != nil {
		return nil, err
	}
	comments := []*commentResolver{}
	for _, c := range cs {
		comments = append(comments, &commentResolver{comment: c})
	}
	return comments, nil
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
	repo, err := store.LocalRepos.Get(ctx, args.RemoteURI, args.AccessToken)
	if err == store.ErrRepoNotFound {
		repo, err = store.LocalRepos.Create(ctx, &sourcegraph.LocalRepo{
			RemoteURI:   args.RemoteURI,
			AccessToken: args.AccessToken,
		})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	newThread, err := store.Threads.Create(ctx, &sourcegraph.Thread{
		LocalRepoID:    repo.ID,
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

	comment, err := store.Comments.Create(ctx, newThread.ID, args.Contents, args.AuthorName, args.AuthorEmail)
	if err != nil {
		return nil, err
	}
	notifyThreadParticipants(repo, newThread, nil, comment)

	return &threadResolver{thread: newThread}, nil
}

func (*schemaResolver) UpdateThread(ctx context.Context, args *struct {
	RemoteURI   string
	AccessToken string
	ThreadID    int32
	Archived    *bool
}) (*threadResolver, error) {
	// 🚨 SECURITY: DO NOT REMOVE THIS CHECK! LocalRepos.Get is responsible for 🚨
	// ensuring the user has permissions to access the repository.
	repo, err := store.LocalRepos.Get(ctx, args.RemoteURI, args.AccessToken)
	if err != nil {
		return nil, err
	}

	thread, err := store.Threads.Update(ctx, int64(args.ThreadID), int64(repo.ID), args.Archived)
	if err != nil {
		return nil, err
	}
	return &threadResolver{thread: thread}, nil
}

// titleFromContents returns a title based on the first sentence or line of the content.
func titleFromContents(contents string) string {
	matchEndpoint := regexp.MustCompile(`[.!?]\s`)
	var title string
	if idxs := matchEndpoint.FindStringSubmatchIndex(contents); len(idxs) > 0 {
		title = contents[:idxs[0]+1]
	} else if i := strings.Index(contents, "\n"); i != -1 {
		title = contents[:i]
	} else {
		title = contents
	}
	if len(title) > 140 {
		title = title[:140] + "..."
	}
	return title
}
