package codeowners

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/api"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/paths"
)

type RulesetSource interface {
	// Used to type guard.
	rulesetSource()
}

// GitRulesetSource describes the specific codeowners file that was used from a
// git repository.
type GitRulesetSource struct {
	Repo   api.RepoID
	Commit api.CommitID
	Path   string
}

func (GitRulesetSource) rulesetSource() {}

// IngestedRulesetSource describes the codeowners file was taken from ingested
// data.
type IngestedRulesetSource struct {
	ID int32
}

func (IngestedRulesetSource) rulesetSource() {}

type Ruleset struct {
	proto        *codeownerspb.File
	rules        []*CompiledRule
	source       RulesetSource
	codeHostType string
}

func NewRuleset(source RulesetSource, proto *codeownerspb.File) *Ruleset {
	f := &Ruleset{
		proto:  proto,
		source: source,
	}
	for _, r := range proto.GetRule() {
		f.rules = append(f.rules, &CompiledRule{proto: r})
	}
	return f
}

func (r *Ruleset) GetFile() *codeownerspb.File {
	return r.proto
}

func (r *Ruleset) GetSource() RulesetSource {
	return r.source
}

func (r *Ruleset) GetCodeHostType() string {
	return r.codeHostType
}

func (r *Ruleset) SetCodeHostType(cht string) {
	r.codeHostType = cht
}

// Match returns the rule matching the given path as per this CODEOWNERS ruleset.
// Rules are evaluated in order: The returned rule is the rule which pattern matches
// the given path that is the furthest down the input file.
func (x *Ruleset) Match(path string) *codeownerspb.Rule {
	// For pattern matching, we expect paths to start with a `/`. Several internal
	// systems don't use leading `/` though, so we ensure it's always there here.
	if path[0] != '/' {
		path = "/" + path
	}
	for i := len(x.rules) - 1; i >= 0; i-- {
		rule := x.rules[i]
		if rule.match(path) {
			return rule.proto
		}
	}
	return nil
}

type CompiledRule struct {
	proto       *codeownerspb.Rule
	glob        *paths.GlobPattern
	compileOnce sync.Once
}

func (r *CompiledRule) match(filePath string) bool {
	r.compileOnce.Do(func() {
		// For now, we ignore errors.
		r.glob, _ = paths.Compile(r.proto.GetPattern())
	})
	// If we saw any error on compiling the glob, we just treat this as a no-match case.
	if r.glob == nil {
		return false
	}
	return r.glob.Match(filePath)
}
