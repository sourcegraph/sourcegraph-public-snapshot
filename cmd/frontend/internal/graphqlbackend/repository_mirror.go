package graphqlbackend

import (
	"context"
	"sync"
	"time"

	graphql "github.com/neelance/graphql-go"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
)

func (r *repositoryResolver) MirrorInfo() *repositoryMirrorInfoResolver {
	return &repositoryMirrorInfoResolver{repository: r}
}

type repositoryMirrorInfoResolver struct {
	repository *repositoryResolver

	// memoize the gitserver RepoInfo call
	repoInfoOnce     sync.Once
	repoInfoResponse *protocol.RepoInfoResponse
	repoInfoErr      error
}

func (r *repositoryMirrorInfoResolver) gitserverRepoInfo(ctx context.Context) (*protocol.RepoInfoResponse, error) {
	r.repoInfoOnce.Do(func() {
		r.repoInfoResponse, r.repoInfoErr = gitserver.DefaultClient.RepoInfo(ctx, r.repository.repo.URI)
	})
	return r.repoInfoResponse, r.repoInfoErr
}

func (r *repositoryMirrorInfoResolver) RemoteURL(ctx context.Context) (string, error) {
	// 🚨 SECURITY: The remote URL might contain secret credentials in the URL userinfo, so
	// only allow site admins to see it.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return "", err
	}

	info, err := r.gitserverRepoInfo(ctx)
	if err != nil {
		return "", err
	}
	return info.URL, nil
}

func (r *repositoryMirrorInfoResolver) Cloned(ctx context.Context) (bool, error) {
	info, err := r.gitserverRepoInfo(ctx)
	if err != nil {
		return false, err
	}
	return info.Cloned, nil
}

func (r *repositoryMirrorInfoResolver) CloneInProgress(ctx context.Context) (bool, error) {
	info, err := r.gitserverRepoInfo(ctx)
	if err != nil {
		return false, err
	}
	return info.CloneInProgress, nil
}

func (r *repositoryMirrorInfoResolver) UpdatedAt(ctx context.Context) (*string, error) {
	info, err := r.gitserverRepoInfo(ctx)
	if err != nil {
		return nil, err
	}
	if info.LastFetched == nil {
		return nil, err
	}
	s := info.LastFetched.Format(time.RFC3339)
	return &s, nil
}

func (r *schemaResolver) CheckMirrorRepositoryConnection(ctx context.Context, args *struct {
	Repository graphql.ID
}) (*checkMirrorRepositoryConnectionResult, error) {
	// 🚨 SECURITY: This is an expensive operation and the errors may contain secrets,
	// so only site admins may run it.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	repo, err := repositoryByID(ctx, args.Repository)
	if err != nil {
		return nil, err
	}

	var result checkMirrorRepositoryConnectionResult
	if err := gitserver.DefaultClient.IsRepoCloneable(ctx, repo.repo.URI); err != nil {
		result.errorMessage = err.Error()
	}
	return &result, nil
}

type checkMirrorRepositoryConnectionResult struct {
	errorMessage string
}

func (r *checkMirrorRepositoryConnectionResult) Error() *string {
	if r.errorMessage == "" {
		return nil
	}
	return &r.errorMessage
}

func (r *schemaResolver) UpdateMirrorRepository(ctx context.Context, args *struct {
	Repository graphql.ID
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: There is no reason why non-site-admins would need to run this operation.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	repo, err := repositoryByID(ctx, args.Repository)
	if err != nil {
		return nil, err
	}

	gitserverRepo, err := backend.Repos.GitserverRepoInfo(ctx, repo.repo)
	if err != nil {
		return nil, err
	}
	if err := gitserver.DefaultClient.EnqueueRepoUpdate(ctx, gitserverRepo); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
