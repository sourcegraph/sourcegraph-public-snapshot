package codeownership

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type RulesKey struct {
	repoName api.RepoName
	commitID api.CommitID
}

type RulesCache struct {
	rules map[RulesKey]Ruleset

	mu sync.Mutex
}

func NewRulesCache() RulesCache {
	return RulesCache{rules: make(map[RulesKey]Ruleset)}
}

func (c *RulesCache) GetFromCacheOrFetch(ctx context.Context, gitserver gitserver.Client, repoName api.RepoName, commitID api.CommitID) (Ruleset, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := RulesKey{repoName, commitID}
	if _, ok := c.rules[key]; !ok {
		file, err := backend.NewOwnService(gitserver).OwnersFile(ctx, repoName, commitID)
		if err != nil {
			emptyRuleset := Ruleset{}
			c.rules[key] = emptyRuleset
			return emptyRuleset, err
		}
		c.rules[key] = Ruleset{file}
	}
	return c.rules[key], nil
}
