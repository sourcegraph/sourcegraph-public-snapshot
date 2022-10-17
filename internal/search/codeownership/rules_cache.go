package codeownership

import (
	"context"
	"sync"

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
	ruleset, ok := c.rules[key]
	var err error
	if !ok {
		ruleset, err = NewRuleset(ctx, gitserver, repoName, commitID)
		if err != nil {
			emptyRuleset := Ruleset{}
			c.rules[key] = emptyRuleset
			return emptyRuleset, err
		}
		c.rules[key] = ruleset
	}

	return ruleset, nil
}
