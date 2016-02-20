package plan

import "sourcegraph.com/sourcegraph/makex"

// ruleSort sorts rules by target name, alphabetically. It is used to
// enforce stable ordering so that Makefiles are consistently
// reproducible given the same set of inputs to CreateMakefile.
type ruleSort struct {
	Rules []makex.Rule
}

func (s ruleSort) Len() int { return len(s.Rules) }
func (s ruleSort) Less(i, j int) bool {
	return s.Rules[i].Target() < s.Rules[j].Target()
}
func (s ruleSort) Swap(i, j int) {
	s.Rules[i], s.Rules[j] = s.Rules[j], s.Rules[i]
}
