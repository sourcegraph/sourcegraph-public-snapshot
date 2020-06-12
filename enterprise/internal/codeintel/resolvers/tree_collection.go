package resolvers

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type repositoryCollectionResolver struct {
	m                         sync.RWMutex
	commitCollectionResolvers map[api.RepoID]*commitCollectionResolver
}

// resolve returns a GitTreeEntryResolver for the given repository, commit, and path. This will cache
// the repository, commit, and path resolvers if they have been previously constructed with this same
// struct instance. If the commit resolver cannot be constructed, a nil resolver is returned.
func (r *repositoryCollectionResolver) resolve(ctx context.Context, repoID api.RepoID, commit, path string) (*graphqlbackend.GitTreeEntryResolver, error) {
	commitCollectionResolver, err := r.resolveRepository(ctx, repoID)
	if err != nil {
		return nil, err
	}

	pathCollectionResolver, err := commitCollectionResolver.resolveCommit(ctx, commit)
	if err != nil {
		return nil, err
	}

	return pathCollectionResolver.resolvePath(ctx, path)
}

// resolveRepository returns a commitCollectionResolver with the given resolved repository.
func (r *repositoryCollectionResolver) resolveRepository(ctx context.Context, repoID api.RepoID) (*commitCollectionResolver, error) {
	r.m.RLock()
	if payload, ok := r.commitCollectionResolvers[repoID]; ok {
		r.m.RUnlock()
		return payload, nil
	}
	r.m.RUnlock()

	r.m.Lock()
	defer r.m.Unlock()
	if payload, ok := r.commitCollectionResolvers[repoID]; ok {
		return payload, nil
	}

	repositoryResolver, err := resolveRepository(ctx, repoID)
	if err != nil {
		return nil, err
	}

	payload := &commitCollectionResolver{
		repositoryResolver:      repositoryResolver,
		pathCollectionResolvers: map[string]*pathCollectionResolver{},
	}

	r.commitCollectionResolvers[repoID] = payload
	return payload, nil
}

type commitCollectionResolver struct {
	repositoryResolver *graphqlbackend.RepositoryResolver

	m                       sync.RWMutex
	pathCollectionResolvers map[string]*pathCollectionResolver
}

// resolveCommit returns a pathCollectionResolver with the given resolved commit.
func (r *commitCollectionResolver) resolveCommit(ctx context.Context, commit string) (*pathCollectionResolver, error) {
	r.m.RLock()
	if resolver, ok := r.pathCollectionResolvers[commit]; ok {
		r.m.RUnlock()
		return resolver, nil
	}
	r.m.RUnlock()

	r.m.Lock()
	defer r.m.Unlock()
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

	m             sync.RWMutex
	pathResolvers map[string]*graphqlbackend.GitTreeEntryResolver
}

// pathCollectionResolver returns a GitTreeEntryResolver with the given path. If the
// commit resolver could not be constructed, a nil resolver is returned.
func (r *pathCollectionResolver) resolvePath(ctx context.Context, path string) (*graphqlbackend.GitTreeEntryResolver, error) {
	r.m.RLock()
	if resolver, ok := r.pathResolvers[path]; ok {
		r.m.RUnlock()
		return resolver, nil
	}
	r.m.RUnlock()

	r.m.Lock()
	defer r.m.Unlock()
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
