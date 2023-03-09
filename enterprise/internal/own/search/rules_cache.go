package search

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/own"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type RulesKey struct {
	repoName api.RepoName
	commitID api.CommitID
}

type RulesCache struct {
	rules      map[RulesKey]*codeowners.Ruleset
	ownService own.Service

	mu sync.RWMutex
}

func NewRulesCache(gs gitserver.Client, db database.DB) RulesCache {
	return RulesCache{
		rules:      make(map[RulesKey]*codeowners.Ruleset),
		ownService: own.NewService(gs, db),
	}
}

func (c *RulesCache) GetFromCacheOrFetch(ctx context.Context, repoName api.RepoName, repoID api.RepoID, commitID api.CommitID) (*codeowners.Ruleset, error) {
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
		file, err := c.ownService.RulesetForRepo(ctx, repoName, repoID, commitID)
		if err != nil || file == nil {
			emptyRuleset := codeowners.NewRuleset(&codeownerspb.File{})
			c.rules[key] = emptyRuleset
			return emptyRuleset, err
		}
		c.rules[key] = file
	}
	return c.rules[key], nil
}
