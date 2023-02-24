package codeownership

import (
	"context"
	"fmt"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
)

type cacheKey struct {
	repoID   api.RepoID
	commitID api.CommitID
}

type ResolvedOwnersFile struct {
	resCtx         backend.OwnerResolutionContext
	file           *codeownerspb.File
	resolvedOwners map[backend.OwnerKey]codeowners.ResolvedOwner
}

func (r *ResolvedOwnersFile) FindOwners(path string) (ret []codeowners.ResolvedOwner) {
	owners := r.file.FindOwners(path)
	for _, o := range owners {
		ro, ok := r.resolvedOwners[backend.NewOwnerKey(o.Handle, o.Email, r.resCtx)]
		if !ok {
			fmt.Printf("OH NO THIS OWNER DIDN'T EXIST\n")
			continue
		}
		ret = append(ret, ro)
	}
	return ret
}

type Cache struct {
	entries    map[cacheKey]*ResolvedOwnersFile
	ownService backend.OwnService

	mu sync.RWMutex
}

func NewCache(ownService backend.OwnService) *Cache {
	return &Cache{
		entries:    make(map[cacheKey]*ResolvedOwnersFile),
		ownService: ownService,
	}
}

func (c *Cache) GetFromCacheOrFetch(ctx context.Context, repoID api.RepoID, repoName api.RepoName, commitID api.CommitID) (*ResolvedOwnersFile, error) {
	resCtx := backend.OwnerResolutionContext{
		RepoID:   repoID,
		RepoName: repoName,
	}

	c.mu.RLock()
	key := cacheKey{repoID, commitID}
	if _, ok := c.entries[key]; ok {
		defer c.mu.RUnlock()
		return c.entries[key], nil
	}
	c.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	// Recheck condition.
	if _, ok := c.entries[key]; !ok {
		file, err := c.ownService.OwnersFile(ctx, repoName, commitID)
		if err != nil {
			emptyRuleset := &codeownerspb.File{}
			r := &ResolvedOwnersFile{
				file:           emptyRuleset,
				resolvedOwners: make(map[backend.OwnerKey]codeowners.ResolvedOwner),
			}
			c.entries[key] = r
			return r, err
		}

		resolvedOwners, err := c.ownService.ResolveOwnersWithType(ctx, ownersForFile(file), resCtx)
		// Warning: Failure here means that might always exercise a costly path, we
		// should store the error or keep a fallback.
		if err != nil {
			return nil, err
		}

		r := &ResolvedOwnersFile{
			file:           file,
			resolvedOwners: resolvedOwners,
			resCtx:         resCtx,
		}

		c.entries[key] = r
	}
	return c.entries[key], nil
}

func ownersForFile(file *codeownerspb.File) (allOwners []*codeownerspb.Owner) {
	seenOwners := make(map[string]struct{})

	for _, rule := range file.GetRule() {
		for _, owner := range rule.Owner {
			key := owner.Handle + ":" + owner.Email
			if _, ok := seenOwners[key]; ok {
				continue
			}
			allOwners = append(allOwners, owner)
			seenOwners[key] = struct{}{}
		}
	}

	return allOwners
}
