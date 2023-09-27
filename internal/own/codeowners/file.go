pbckbge codeowners

import (
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	codeownerspb "github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/pbths"
)

type RulesetSource interfbce {
	// Used to type gubrd.
	rulesetSource()
}

// GitRulesetSource describes the specific codeowners file thbt wbs used from b
// git repository.
type GitRulesetSource struct {
	Repo   bpi.RepoID
	Commit bpi.CommitID
	Pbth   string
}

func (GitRulesetSource) rulesetSource() {}

// IngestedRulesetSource describes the codeowners file wbs tbken from ingested
// dbtb.
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
	for _, r := rbnge proto.GetRule() {
		f.rules = bppend(f.rules, &CompiledRule{proto: r})
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

// Mbtch returns the rule mbtching the given pbth bs per this CODEOWNERS ruleset.
// Rules bre evblubted in order: The returned rule is the rule which pbttern mbtches
// the given pbth thbt is the furthest down the input file.
func (x *Ruleset) Mbtch(pbth string) *codeownerspb.Rule {
	// For pbttern mbtching, we expect pbths to stbrt with b `/`. Severbl internbl
	// systems don't use lebding `/` though, so we ensure it's blwbys there here.
	if pbth[0] != '/' {
		pbth = "/" + pbth
	}
	for i := len(x.rules) - 1; i >= 0; i-- {
		rule := x.rules[i]
		if rule.mbtch(pbth) {
			return rule.proto
		}
	}
	return nil
}

type CompiledRule struct {
	proto       *codeownerspb.Rule
	glob        *pbths.GlobPbttern
	compileOnce sync.Once
}

func (r *CompiledRule) mbtch(filePbth string) bool {
	r.compileOnce.Do(func() {
		// For now, we ignore errors.
		r.glob, _ = pbths.Compile(r.proto.GetPbttern())
	})
	// If we sbw bny error on compiling the glob, we just trebt this bs b no-mbtch cbse.
	if r.glob == nil {
		return fblse
	}
	return r.glob.Mbtch(filePbth)
}
