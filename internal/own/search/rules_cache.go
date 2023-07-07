package search

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/own"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
)

type RulesKey struct {
	repoName api.RepoName
	commitID api.CommitID
}

type AssignedKey struct {
	repoID api.RepoID
}

type RulesCache struct {
	rules         map[RulesKey]*codeowners.Ruleset
	assigned      map[AssignedKey]own.AssignedOwners
	assignedTeams map[AssignedKey]own.AssignedTeams
	ownService    own.Service

	rulesMu         sync.RWMutex
	assignedMu      sync.RWMutex
	assignedTeamsMu sync.RWMutex
}

func NewRulesCache(gs gitserver.Client, db database.DB) RulesCache {
	return RulesCache{
		rules:         make(map[RulesKey]*codeowners.Ruleset),
		assigned:      make(map[AssignedKey]own.AssignedOwners),
		assignedTeams: make(map[AssignedKey]own.AssignedTeams),
		ownService:    own.NewService(gs, db),
	}
}

func (c *RulesCache) GetFromCacheOrFetch(ctx context.Context, repoName api.RepoName, repoID api.RepoID, commitID api.CommitID) (repoOwnershipData, error) {
	assigned, err := c.AssignedOwners(ctx, repoID, commitID)
	if err != nil {
		return repoOwnershipData{}, err
	}
	assignedTeams, err := c.AssignedTeams(ctx, repoID, commitID)
	if err != nil {
		return repoOwnershipData{}, err
	}
	codeowners, err := c.Codeowners(ctx, repoName, repoID, commitID)
	if err != nil {
		return repoOwnershipData{}, err
	}
	return repoOwnershipData{
		assigned:      assigned,
		assignedTeams: assignedTeams,
		codeowners:    codeowners,
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
			// Error is picked up on a call site and in most cases a search alert is created.
			return nil, err
		}
		c.assigned[key] = assigned
	}
	return c.assigned[key], nil
}

func (c *RulesCache) AssignedTeams(ctx context.Context, repoID api.RepoID, commitID api.CommitID) (own.AssignedTeams, error) {
	c.assignedTeamsMu.RLock()
	key := AssignedKey{repoID}
	if v, ok := c.assignedTeams[key]; ok {
		defer c.assignedTeamsMu.RUnlock()
		return v, nil
	}
	c.assignedTeamsMu.RUnlock()
	c.assignedTeamsMu.Lock()
	defer c.assignedTeamsMu.Unlock()
	if _, ok := c.assignedTeams[key]; !ok {
		assigned, err := c.ownService.AssignedTeams(ctx, repoID, commitID)
		if err != nil {
			// Error is picked up on a call site and in most cases a search alert is created.
			return nil, err
		}
		c.assignedTeams[key] = assigned
	}
	return c.assignedTeams[key], nil
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
	codeowners    *codeowners.Ruleset
	assigned      own.AssignedOwners
	assignedTeams own.AssignedTeams
}

func (o repoOwnershipData) Match(path string) fileOwnershipData {
	var rule *codeownerspb.Rule
	if o.codeowners != nil {
		rule = o.codeowners.Match(path)
	}
	return fileOwnershipData{
		rule:           rule,
		assignedOwners: o.assigned.Match(path),
		assignedTeams:  o.assignedTeams.Match(path),
	}
}

type fileOwnershipData struct {
	rule           *codeownerspb.Rule
	assignedOwners []database.AssignedOwnerSummary
	assignedTeams  []database.AssignedTeamSummary
}

func (d fileOwnershipData) References() []own.Reference {
	var rs []own.Reference
	for _, o := range d.rule.GetOwner() {
		rs = append(rs, own.Reference{Handle: o.Handle, Email: o.Email})
	}
	for _, o := range d.assignedOwners {
		rs = append(rs, own.Reference{UserID: o.OwnerUserID})
	}
	for _, o := range d.assignedTeams {
		rs = append(rs, own.Reference{TeamID: o.OwnerTeamID})
	}
	return rs
}

func (d fileOwnershipData) Empty() bool {
	return !d.NonEmpty()
}

func (d fileOwnershipData) NonEmpty() bool {
	if d.rule != nil && len(d.rule.Owner) > 0 {
		return true
	}
	if len(d.assignedOwners) > 0 {
		return true
	}
	if len(d.assignedTeams) > 0 {
		return true
	}
	return false
}

func (d fileOwnershipData) IsWithin(bag own.Bag) bool {
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
	for _, o := range d.assignedTeams {
		if bag.Contains(own.Reference{TeamID: o.OwnerTeamID}) {
			return true
		}
	}
	return false
}

func (d fileOwnershipData) String() string {
	var references []string
	for _, o := range d.rule.GetOwner() {
		if h := o.GetHandle(); h != "" {
			references = append(references, h)
		}
		if e := o.GetEmail(); e != "" {
			references = append(references, e)
		}
	}
	for _, o := range d.assignedOwners {
		references = append(references, fmt.Sprintf("#%d", o.OwnerUserID))
	}
	for _, o := range d.assignedTeams {
		references = append(references, fmt.Sprintf("#%d", o.OwnerTeamID))
	}
	return fmt.Sprintf("[%s]", strings.Join(references, ", "))
}
