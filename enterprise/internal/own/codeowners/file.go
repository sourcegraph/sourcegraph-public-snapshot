package codeowners

import (
	"sync"

	codeownerspb "github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/paths"
)

type Ruleset struct {
	proto *codeownerspb.File
	rules []*CompiledRule
}

func NewRuleset(proto *codeownerspb.File) *Ruleset {
	f := &Ruleset{
		proto: proto,
	}
	for _, r := range proto.GetRule() {
		f.rules = append(f.rules, &CompiledRule{proto: r})
	}
	return f
}

func (r *Ruleset) GetFile() *codeownerspb.File {
	return r.proto
}

// FindOwners returns the Owners associated with given path as per this CODEOWNERS file.
// Rules are evaluated in order: Returned owners come from the rule which pattern matches
// given path, that is the furthest down the file.
func (x *Ruleset) FindOwners(path string) []*codeownerspb.Owner {
	// For pattern matching, we expect paths to start with a `/`. Several internal
	// systems don't use leading `/` though, so we ensure it's always there here.
	if path[0] != '/' {
		path = "/" + path
	}
	for i := len(x.rules) - 1; i >= 0; i-- {
		rule := x.rules[i]
		if rule.match(path) {
			return rule.GetOwner()
		}
	}
	return nil
}

type CompiledRule struct {
	proto       *codeownerspb.Rule
	glob        paths.GlobPattern
	compileOnce sync.Once
}

func (r *CompiledRule) match(filePath string) bool {
	r.compileOnce.Do(func() {
		// For now, we ignore errors.
		r.glob, _ = paths.Compile(r.proto.GetPattern())
	})
	return r.glob.Match(filePath)
}

func (r *CompiledRule) GetOwner() []*codeownerspb.Owner {
	return r.proto.Owner
}
