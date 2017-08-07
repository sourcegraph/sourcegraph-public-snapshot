package graphqlbackend

import (
	"context"
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

func (r *rootResolver) Threads(ctx context.Context, args *struct {
	RemoteURI   string
	AccessToken string
	File        string
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

	ts, err := store.Threads.GetAllForFile(ctx, int64(repo.ID), args.File)
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

	comment, err := store.Comments.Create(ctx, &sourcegraph.Comment{
		ThreadID:    newThread.ID,
		Contents:    args.Contents,
		AuthorName:  args.AuthorName,
		AuthorEmail: args.AuthorEmail,
	})
	if err != nil {
		return nil, err
	}
	notifyThreadParticipants(newThread, nil, comment)

	return &threadResolver{thread: newThread}, nil
}
