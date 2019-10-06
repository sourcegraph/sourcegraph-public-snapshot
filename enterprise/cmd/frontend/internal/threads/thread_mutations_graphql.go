package threads

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func (GraphQLResolver) CreateThread(ctx context.Context, arg *graphqlbackend.CreateThreadArgs) (graphqlbackend.Thread, error) {
	repo, err := graphqlbackend.RepositoryByID(ctx, arg.Input.Repository)
	if err != nil {
		return nil, err
	}

	author, err := comments.CommentActorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	comment := commentobjectdb.DBObjectCommentFields{Author: author}
	if arg.Input.Body != nil {
		comment.Body = *arg.Input.Body
	}

	data := &DBThread{
		RepositoryID: repo.DBID(),
		Title:        arg.Input.Title,
		IsDraft:      arg.Input.Draft != nil && *arg.Input.Draft,
		State:        string(graphqlbackend.ThreadStateOpen),
	}
	if arg.Input.BaseRef != nil {
		data.BaseRef = *arg.Input.BaseRef
	}
	if arg.Input.HeadRef != nil {
		data.HeadRef = *arg.Input.HeadRef
	}
	thread, err := dbThreads{}.Create(ctx, nil, data, comment)
	if err != nil {
		return nil, err
	}
	gqlThread := newGQLThread(thread)

	if arg.Input.RawDiagnostics != nil {
		if _, err := graphqlbackend.ThreadDiagnostics.AddDiagnosticsToThread(ctx, &graphqlbackend.AddDiagnosticsToThreadArgs{Thread: gqlThread.ID(), RawDiagnostics: *arg.Input.RawDiagnostics}); err != nil {
			return nil, err
		}
	}

	return gqlThread, nil
}

func (GraphQLResolver) ImportThreadsFromExternalService(ctx context.Context, arg *graphqlbackend.ImportThreadsFromExternalServiceArgs) ([]graphqlbackend.Thread, error) {
	threadsByDBIDs := func(ctx context.Context, threadIDs []int64) ([]graphqlbackend.Thread, error) {
		threads := make([]graphqlbackend.Thread, len(threadIDs))
		for i, threadID := range threadIDs {
			thread, err := threadByDBID(ctx, threadID)
			if err != nil {
				return nil, err
			}
			threads[i] = thread
		}
		return threads, nil
	}

	switch {
	case arg.Input.ByRepositoryAndNumber != nil:
		spec := arg.Input.ByRepositoryAndNumber
		repoName := spec.RepositoryName
		// TODO!(sqs): hacky
		if !strings.HasPrefix(repoName, "github.com/") {
			repoName = "github.com/" + repoName
		}
		repo, err := backend.Repos.GetByName(ctx, api.RepoName(repoName))
		if err != nil {
			return nil, err
		}
		_ = repo
		_ = threadsByDBIDs
		panic("TODO!(sqs): removed bc not yet implemented for bitbucket server")
		// threadID, err := createOrGetExistingGitHubIssueOrPullRequest(ctx, repo.ID, repo.ExternalRepo, spec.Number)
		// if err != nil {
		// 	return nil, err
		// }
		// return threadsByDBIDs(ctx, []int64{threadID})

	case arg.Input.ByQuery != nil:
		panic("TODO!(sqs): removed bc not yet implemented for bitbucket server")
		// threadIDs, err := createOrGetExistingGitHubThreadsByQuery(ctx, *arg.Input.ByQuery)
		// if err != nil {
		// 	return nil, err
		// }
		// return threadsByDBIDs(ctx, threadIDs)

	default:
		return nil, errors.New("no threads specified to import from external service")
	}
}

func (GraphQLResolver) UpdateThread(ctx context.Context, arg *graphqlbackend.UpdateThreadArgs) (graphqlbackend.Thread, error) {
	l, err := threadByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	thread, err := dbThreads{}.Update(ctx, l.db.ID, dbThreadUpdate{
		Title: arg.Input.Title,
		// TODO!(sqs): handle body update
		BaseRef: arg.Input.BaseRef,
		HeadRef: arg.Input.HeadRef,
	})
	if err != nil {
		return nil, err
	}
	return newGQLThread(thread), nil
}

func (GraphQLResolver) MarkThreadAsReady(ctx context.Context, arg *graphqlbackend.MarkThreadAsReadyArgs) (graphqlbackend.Thread, error) {
	t, err := threadByID(ctx, arg.Thread)
	if err != nil {
		return nil, err
	}
	if !t.db.IsDraft {
		return nil, errors.New("thread is not a draft thread")
	}
	tmp := false
	thread, err := dbThreads{}.Update(ctx, t.db.ID, dbThreadUpdate{
		IsDraft: &tmp,
	})
	if err != nil {
		return nil, err
	}
	return newGQLThread(thread), nil
}

func (GraphQLResolver) PublishThreadToExternalService(ctx context.Context, arg *graphqlbackend.PublishThreadToExternalServiceArgs) (graphqlbackend.Thread, error) {
	t, err := threadByID(ctx, arg.Thread)
	if err != nil {
		return nil, err
	}
	if err := PublishThreadToExternalService(ctx, t); err != nil {
		return nil, err
	}
	return threadByID(ctx, arg.Thread)
}

func PublishThreadToExternalService(ctx context.Context, thread_ graphqlbackend.Thread) error {
	thread := thread_.(*gqlThread).db
	if !thread.IsPendingExternalCreation {
		return errors.New("thread is not pending external creation")
	}

	repo, err := thread_.Repository(ctx)
	if err != nil {
		return err
	}
	threadBody, err := thread_.Body(ctx)
	if err != nil {
		return err
	}

	// TODO!(sqs): use actual campaign name, but its weird because a thread can be in >1 campaigns
	campaignName := thread_.Title()
	_, err = CreateOnExternalService(ctx, thread.ID, thread_.Title(), threadBody, campaignName, repo, []byte(thread.PendingPatch))
	return err
}

func (GraphQLResolver) DeleteThread(ctx context.Context, arg *graphqlbackend.DeleteThreadArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlThread, err := threadByID(ctx, arg.Thread)
	if err != nil {
		return nil, err
	}
	return nil, dbThreads{}.DeleteByID(ctx, gqlThread.db.ID)
}
