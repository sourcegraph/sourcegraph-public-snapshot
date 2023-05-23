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

type AssignedKey struct {
	repoID api.RepoID
}

type RulesCache struct {
	rules      map[RulesKey]*codeowners.Ruleset
	assigned   map[AssignedKey]own.AssignedOwners
	ownService own.Service

	rulesMu    sync.RWMutex
	assignedMu sync.RWMutex
}

func NewRulesCache(gs gitserver.Client, db database.DB) RulesCache {
	return RulesCache{
		rules:      make(map[RulesKey]*codeowners.Ruleset),
		assigned:   make(map[AssignedKey]own.AssignedOwners),
		ownService: own.NewService(gs, db),
	}
}

func (c *RulesCache) GetFromCacheOrFetch(ctx context.Context, repoName api.RepoName, repoID api.RepoID, commitID api.CommitID) (repoOwnershipData, error) {
	assigned, err := c.AssignedOwners(ctx, repoID, commitID)
	if err != nil {
		return repoOwnershipData{}, err
	}
	codeowners, err := c.Codeowners(ctx, repoName, repoID, commitID)
	if err != nil {
		return repoOwnershipData{}, err
	}
	return repoOwnershipData{
		assigned:   assigned,
		codeowners: codeowners,
	}, nil
}

func (c *RulesCache) AssignedOwners(ctx context.Context, repoID api.RepoID, commitID api.CommitID) (own.AssignedOwners, error) {
	c.assignedMu.RLock()
	key := AssignedKey{repoID}
	if v, ok := c.assigned[key]; ok {
		defer c.assignedMu.RUnlock()
		return v, nil
	}
	c.assignedMu.RUnlock()
	c.assignedMu.Lock()
	defer c.assignedMu.Unlock()
	if _, ok := c.assigned[key]; !ok {
		assigned, err := c.ownService.AssignedOwnership(ctx, repoID, commitID)
		if err != nil {
			// TODO: Consider error condition
			return nil, err
		}
		c.assigned[key] = assigned
	}
	return c.assigned[key], nil
}

func (c *RulesCache) Codeowners(ctx context.Context, repoName api.RepoName, repoID api.RepoID, commitID api.CommitID) (*codeowners.Ruleset, error) {
	c.rulesMu.RLock()
	key := RulesKey{repoName, commitID}
	if _, ok := c.rules[key]; ok {
		defer c.rulesMu.RUnlock()
		return c.rules[key], nil
	}
	c.rulesMu.RUnlock()
	c.rulesMu.Lock()
	defer c.rulesMu.Unlock()
	// Recheck condition.
	if _, ok := c.rules[key]; !ok {
		file, err := c.ownService.RulesetForRepo(ctx, repoName, repoID, commitID)
		if err != nil || file == nil {
			// TODO: Ideally we wouldn't use an empty ruleset here, and instead
			// check if this returns a nil ruleset.
			emptyRuleset := codeowners.NewRuleset(nil, &codeownerspb.File{})
			c.rules[key] = emptyRuleset
			return emptyRuleset, err
		}
		c.rules[key] = file
	}
	return c.rules[key], nil
}

type repoOwnershipData struct {
	codeowners *codeowners.Ruleset
	assigned   own.AssignedOwners
}

func (o repoOwnershipData) Match(path string) fileOwnershipData {
	var rule *codeownerspb.Rule
	if o.codeowners != nil {
		rule = o.codeowners.Match(path)
	}
	return fileOwnershipData{
		rule:           rule,
		assignedOwners: o.assigned.Match(path),
	}
}

type fileOwnershipData struct {
	rule           *codeownerspb.Rule
	assignedOwners []database.AssignedOwnerSummary
}

func (d fileOwnershipData) NonEmpty() bool {
	if d.rule != nil && len(d.rule.Owner) > 0 {
		return true
	}
	if len(d.assignedOwners) > 0 {
		return true
	}
	return false
}

func (d fileOwnershipData) Contains(bag own.Bag) bool {
	for _, o := range d.rule.GetOwner() {
		if bag.Contains(own.Reference{
			Handle: o.Handle,
			Email:  o.Email,
		}) {
			return true
		}
	}
	for _, o := range d.assignedOwners {
		if bag.Contains(own.Reference{UserID: o.OwnerUserID}) {
			return true
		}
	}
	return false
}
