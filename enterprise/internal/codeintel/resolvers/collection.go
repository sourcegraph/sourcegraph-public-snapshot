package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type repositoryCollectionResolver struct {
	commitCollectionResolvers map[string]*commitCollectionResolver
}

// resolve returns a GitTreeEntryResolver for the given repository, commit, and path. This will cache
// the repository, commit, and path resolvers if they have been previously constructed with this same
// struct instance.
func (r *repositoryCollectionResolver) resolve(ctx context.Context, repoName, commit, path string) (*graphqlbackend.GitTreeEntryResolver, error) {
	commitCollectionResolver, err := r.resolveRepository(ctx, repoName)
	if err != nil {
		return nil, err
	}

	pathCollectionResolver, err := commitCollectionResolver.resolveCommit(ctx, commit)
	if err != nil {
		return nil, err
	}

	return pathCollectionResolver.resolvePath(ctx, path)
}

func (r *repositoryCollectionResolver) resolveRepository(ctx context.Context, repoName string) (*commitCollectionResolver, error) {
	if payload, ok := r.commitCollectionResolvers[repoName]; ok {
		return payload, nil
	}

	repositoryResolver, err := resolveRepository(ctx, repoName)
	if err != nil {
		return nil, err
	}

	payload := &commitCollectionResolver{
		repositoryResolver:      repositoryResolver,
		pathCollectionResolvers: map[string]*pathCollectionResolver{},
	}

	r.commitCollectionResolvers[repoName] = payload
	return payload, nil
}

type commitCollectionResolver struct {
	repositoryResolver      *graphqlbackend.RepositoryResolver
	pathCollectionResolvers map[string]*pathCollectionResolver
}

func (r *commitCollectionResolver) resolveCommit(ctx context.Context, commit string) (*pathCollectionResolver, error) {
	if resolver, ok := r.pathCollectionResolvers[commit]; ok {
		return resolver, nil
	}

	commitResolver, err := resolveCommitFrom(ctx, r.repositoryResolver, commit)
	if err != nil {
		return nil, err
	}

	resolver := &pathCollectionResolver{
		commitResolver: commitResolver,
		pathResolvers:  map[string]*graphqlbackend.GitTreeEntryResolver{},
	}

	r.pathCollectionResolvers[commit] = resolver
	return resolver, nil
}

type pathCollectionResolver struct {
	commitResolver *graphqlbackend.GitCommitResolver
	pathResolvers  map[string]*graphqlbackend.GitTreeEntryResolver
}

func (r *pathCollectionResolver) resolvePath(ctx context.Context, path string) (*graphqlbackend.GitTreeEntryResolver, error) {
	if resolver, ok := r.pathResolvers[path]; ok {
		return resolver, nil
	}

	resolver, err := resolvePathFrom(ctx, r.commitResolver, path)
	if err != nil {
		return nil, err
	}

	r.pathResolvers[path] = resolver
	return resolver, nil
}
