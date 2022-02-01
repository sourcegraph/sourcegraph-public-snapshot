// Package overridable provides data types representing values in batch
// specs that can be overridden for specific repositories.
package overridable

import (
	"encoding/json"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gobwas/glob"
)

// allPattern is used to define default rules for the simple scalar case.
const allPattern = "*"

// simpleRule creates the simplest of rules for the given value: `"*": value`.
func simpleRule(v interface{}) *rule {
	r, err := newRule(allPattern, v)
	if err != nil {
		// Since we control the pattern being compiled, an error should never
		// occur.
		panic(err)
	}

	return r
}

type complex []map[string]interface{}

type rule struct {
	// pattern is the glob-syntax pattern, such as "a/b/ceee-*"
	pattern string
	// patternSuffix is an optional suffix that can be appended to the pattern with "@"
	patternSuffix string

	compiled glob.Glob
	value    interface{}
}

// newRule builds a new rule instance, ensuring that the glob pattern
// is compiled.
func newRule(pattern string, value interface{}) (*rule, error) {
	var suffix string
	split := strings.SplitN(pattern, "@", 2)
	if len(split) > 1 {
		pattern = split[0]
		suffix = split[1]
	}

	compiled, err := glob.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return &rule{
		pattern:       pattern,
		patternSuffix: suffix,
		compiled:      compiled,
		value:         value,
	}, nil
}

func (a rule) Equal(b rule) bool {
	return a.pattern == b.pattern && a.value == b.value
}

type rules []*rule

// Match matches the given repository name against all rules, returning the rule value that matches at last, or nil if none match.
func (r rules) Match(name string) interface{} {
	// We want the last match to win, so we'll iterate in reverse order.
	for i := len(r) - 1; i >= 0; i-- {
		if r[i].compiled.Match(name) {
			return r[i].value
		}
	}
	return nil
}

// MatchWithSuffix matches the given repository name against all rules and the
// suffix against provided pattern suffix, returning the rule value that matches
// at last, or nil if none match.
func (r rules) MatchWithSuffix(name, suffix string) interface{} {
	// We want the last match to win, so we'll iterate in reverse order.
	for i := len(r) - 1; i >= 0; i-- {
		if r[i].compiled.Match(name) && (r[i].patternSuffix == "" || r[i].patternSuffix == suffix) {
			return r[i].value
		}
	}
	return nil
}

// MarshalJSON marshalls the bool into its JSON representation, which will
// either be a literal or an array of objects.
func (r rules) MarshalJSON() ([]byte, error) {
	if len(r) == 1 && r[0].pattern == allPattern {
		return json.Marshal(r[0].value)
	}

	rules := []map[string]interface{}{}
	for _, rule := range r {
		rules = append(rules, map[string]interface{}{
			rule.pattern: rule.value,
		})
	}
	return json.Marshal(rules)
}

// hydrateFromComplex builds an array of rules out of a complex value.
func (r *rules) hydrateFromComplex(c []map[string]interface{}) error {
	*r = make(rules, len(c))
	for i, rule := range c {
		if len(rule) != 1 {
			return errors.Errorf("unexpected number of elements in the array at entry %d: %d (must be 1)", i, len(rule))
		}
		for pattern, value := range rule {
			var err error
			(*r)[i], err = newRule(pattern, value)
			if err != nil {
				return errors.Wrapf(err, "building rule for array entry %d", i)
			}
		}
	}
	return nil
}

// Equal tests two rules for equality. Used in cmp.
func (r rules) Equal(other rules) bool {
	if len(r) != len(other) {
		return false
	}

	for i := range r {
		a := r[i]
		b := other[i]
		if !a.Equal(*b) {
			return false
		}
	}

	return true
}
