package codeownership

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
)

type RulesKey struct {
	repoName api.RepoName
	commitID api.CommitID
}

type RulesCache struct {
	rules      map[RulesKey]codeowners.Graph
	ownService backend.OwnService

	mu sync.RWMutex
}

func NewRulesCache(gs gitserver.Client, db database.DB) RulesCache {
	return RulesCache{
		rules:      make(map[RulesKey]codeowners.Graph),
		ownService: backend.NewOwnService(gs, db),
	}
}

func (c *RulesCache) GetFromCacheOrFetch(ctx context.Context, repoName api.RepoName, commitID api.CommitID) (codeowners.Graph, error) {
	c.mu.RLock()
	key := RulesKey{repoName, commitID}
	if _, ok := c.rules[key]; ok {
		defer c.mu.RUnlock()
		return c.rules[key], nil
	}
	c.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	// Recheck condition.
	if _, ok := c.rules[key]; !ok {
		graph, err := c.ownService.Ownership(ctx, repoName, commitID)
		if err != nil {
			nilGraph := codeowners.NilGraph{}
			c.rules[key] = nilGraph
			return nilGraph, err
		}
		c.rules[key] = graph
	}
	return c.rules[key], nil
}
