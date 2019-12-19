package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type treeCollectionResolver struct {
	repositoryCollectionResolvers map[string]*repositoryCollectionResolver
}

type repositoryCollectionResolver struct {
	repositoryResolver        *graphqlbackend.RepositoryResolver
	commitCollectionResolvers map[string]*commitCollectionResolver
}

type commitCollectionResolver struct {
	repositoryResolver      *graphqlbackend.RepositoryResolver
	pathCollectionResolvers map[string]*pathCollectionResolver
}

type pathCollectionResolver struct {
	commitResolver *graphqlbackend.GitCommitResolver
	pathResolvers  map[string]*graphqlbackend.GitTreeEntryResolver
}

func (r *treeCollectionResolver) resolve(ctx context.Context, repoName, commit, path string) (*graphqlbackend.GitTreeEntryResolver, error) {
	repositoryCollectionResolver, err := r.repositoryCollectionResolverFor(ctx, repoName)
	if err != nil {
		return nil, err
	}

	commitCollectionResolver, err := repositoryCollectionResolver.commitCollectionResolverFor(ctx, commit)
	if err != nil {
		return nil, err
	}

	pathCollectionResolver, err := commitCollectionResolver.pathCollectionResolverFor(ctx, path)
	if err != nil {
		return nil, err
	}

	return pathCollectionResolver.pathResolverFor(ctx, path)
}

func (r *treeCollectionResolver) repositoryCollectionResolverFor(ctx context.Context, repoName string) (*repositoryCollectionResolver, error) {
	if payload, ok := r.repositoryCollectionResolvers[repoName]; ok {
		return payload, nil
	}

	repositoryResolver, err := resolveRepository(ctx, repoName)
	if err != nil {
		return nil, err
	}

	payload := &repositoryCollectionResolver{
		repositoryResolver:        repositoryResolver,
		commitCollectionResolvers: map[string]*commitCollectionResolver{},
	}

	r.repositoryCollectionResolvers[repoName] = payload
	return payload, nil
}

func (r *repositoryCollectionResolver) commitCollectionResolverFor(ctx context.Context, commit string) (*commitCollectionResolver, error) {
	if resolver, ok := r.commitCollectionResolvers[commit]; ok {
		return resolver, nil
	}

	resolver := &commitCollectionResolver{
		repositoryResolver:      r.repositoryResolver,
		pathCollectionResolvers: map[string]*pathCollectionResolver{},
	}

	r.commitCollectionResolvers[commit] = resolver
	return resolver, nil
}

func (r *commitCollectionResolver) pathCollectionResolverFor(ctx context.Context, commit string) (*pathCollectionResolver, error) {
	if payload, ok := r.pathCollectionResolvers[commit]; ok {
		return payload, nil
	}

	commitResolver, err := resolveCommitFrom(ctx, r.repositoryResolver, commit)
	if err != nil {
		return nil, err
	}

	payload := &pathCollectionResolver{
		commitResolver: commitResolver,
		pathResolvers:  map[string]*graphqlbackend.GitTreeEntryResolver{},
	}

	r.pathCollectionResolvers[commit] = payload
	return payload, nil
}

func (r *pathCollectionResolver) pathResolverFor(ctx context.Context, path string) (*graphqlbackend.GitTreeEntryResolver, error) {
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
