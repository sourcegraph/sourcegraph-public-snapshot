package codeowners

import (
	"log"

	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
)

type PatternTrie struct {
	rules    []*codeownerspb.Rule
	part     patternPart
	children PatternForist
}

// put associates given pattern with given rule in this sub-trie.
func (p *PatternTrie) put(pattern []patternPart, rule *codeownerspb.Rule) {
	if len(pattern) == 0 {
		p.rules = append(p.rules, rule)
		return
	}
	f := &p.children
	f.Add(pattern, rule)
}

// Forist is to Forest what Trie is to Tree.
type PatternForist []PatternTrie

// Add extends this forist with given pattern and associated rule.
// pattern must be non-empty
func (f *PatternForist) Add(pattern []patternPart, rule *codeownerspb.Rule) {
	if len(pattern) == 0 {
		log.Fatal("empty pattern in a forist")
	}
	part, rest := pattern[0], pattern[1:]
	var q *PatternTrie
	for i := range *f {
		if (*f)[i].part.Eq(part) {
			q = &((*f)[i])
			break
		}
	}
	if q == nil {
		*f = append(*f, PatternTrie{part: part})
		q = &((*f)[len(*f)-1])
	}
	q.put(rest, rule)
}

// Make a single step into a forist with a path chunk.
// Each trie responds to the step, and only matching children
// are returned, while ** introduce proper indeterminism.
//
// Invariant: if f contains ** then children of ** are also in f.
func (f PatternForist) step(pathChunk string) PatternForist {
	var next PatternForist
	for _, t := range f {
		if t.part.Eq(anySubPath{}) {
			next = append(next, t)
			continue
		}
		if t.part.Match(pathChunk) {
			next = append(next, t.children...)
		}
	}
	return next.includeSkipDoubleAsterisk()
}

func (f PatternForist) Find(path []string) *codeownerspb.Rule {
	var state PatternForist
	state = append(state, f...)
	state = state.includeSkipDoubleAsterisk()
	for _, p := range path {
		state = state.step(p)
	}
	var rule *codeownerspb.Rule
	for _, t := range state {
		for _, r := range t.rules {
			if rule == nil || rule.LineNumber < r.LineNumber {
				rule = r
			}
		}
	}
	return rule
}

func (f PatternForist) includeSkipDoubleAsterisk() PatternForist {
	var added PatternForist
	for _, t := range f {
		if t.part.Eq(anySubPath{}) {
			added = append(added, t.children...)
		}
	}
	if added == nil {
		return f
	}
	// NOTE: This can be made into a loop if recursive version does not perform.
	return append(added.includeSkipDoubleAsterisk(), f...)
}
