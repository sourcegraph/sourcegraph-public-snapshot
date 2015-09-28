package makex

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"sourcegraph.com/sourcegraph/rwvfs"
)

// Makefile represents a set of rules, each describing how to build a target.
type Makefile struct {
	Rules []Rule
}

// BasicRule implements Rule.
//
// Use BasicRule for rules that you don't need to introspect
// programmatically. If you need to store additional metadata about
// rules, create a separate type that implements Rule and holds the
// metadata.
type BasicRule struct {
	TargetFile  string
	PrereqFiles []string
	RecipeCmds  []string
}

// Target implements Rule.
func (r *BasicRule) Target() string { return r.TargetFile }

// Prereqs implements Rule.
func (r *BasicRule) Prereqs() []string { return r.PrereqFiles }

// Recipes implements rule.
func (r *BasicRule) Recipes() []string { return r.RecipeCmds }

// Rule returns the rule to make the specified target if it exists, or nil
// otherwise.
//
// TODO(sqs): support multiple rules for one target
// (http://www.gnu.org/software/make/manual/html_node/Multiple-Rules.html).
func (mf *Makefile) Rule(target string) Rule {
	for _, rule := range mf.Rules {
		if rule.Target() == target {
			return rule
		}
	}
	return nil
}

// A Rule describes a target file, a list of commands (recipes) used
// to create the target output file, and the files (which may also
// have corresponding rules) that must exist prior to running the
// recipes.
//
// It is a slightly simplified representation of a standard "make"
// rule.
type Rule interface {
	Target() string
	Prereqs() []string
	Recipes() []string
}

// DefaultRule is the first rule whose name does not begin with a ".", or nil if
// no such rule exists.
func (mf *Makefile) DefaultRule() Rule {
	for _, rule := range mf.Rules {
		target := rule.Target()
		if !strings.HasPrefix(target, ".") {
			return rule
		}
	}
	return nil
}

// Expand returns a clone of mf with Prereqs filepath globs expanded. If rules
// contain globs, they are replaced with BasicRules with the globs expanded.
//
// Only globs containing "*" are detected.
func (c *Config) Expand(orig *Makefile) (*Makefile, error) {
	var mf Makefile
	mf.Rules = make([]Rule, len(orig.Rules))
	for i, rule := range orig.Rules {
		expandedPrereqs, err := c.globs(rule.Prereqs())
		if err != nil {
			return nil, err
		}
		mf.Rules[i] = &BasicRule{
			TargetFile:  rule.Target(),
			PrereqFiles: expandedPrereqs,
			RecipeCmds:  rule.Recipes(),
		}
	}
	return &mf, nil
}

// globs returns all files in the filesystem that match any of the glob patterns
// (using path/filepath.Match glob syntax). The
func (c *Config) globs(patterns []string) (matches []string, err error) {
	for _, pattern := range patterns {
		if strings.ContainsAny(pattern, "*?[]") {
			files, err := c.glob(pattern)
			if err != nil {
				return nil, err
			}
			matches = append(matches, files...)
		} else {
			matches = append(matches, pattern)
		}
	}
	return
}

// glob returns all files in the filesystem that match the glob pattern (using
// path/filepath.Match glob syntax).
func (c *Config) glob(pattern string) (matches []string, err error) {
	return rwvfs.Glob(walkableRWVFS{c.fs()}, globPrefix(pattern), pattern)
}

// globPrefix returns all path components up to (not including) the first path
// component that contains a "*".
func globPrefix(path string) string {
	cs := strings.Split(path, string(filepath.Separator))
	var prefix []string
	for _, c := range cs {
		if strings.Contains(c, "*") {
			break
		}
		prefix = append(prefix, c)
	}
	return filepath.Join(prefix...)
}

// ExpandAutoVars expands the automatic variables $@ (the current target path)
// and $^ (the space-separated list of prereqs) in s.
func ExpandAutoVars(rule Rule, s string) string {
	s = strings.Replace(s, "$@", Quote(rule.Target()), -1)
	s = strings.Replace(s, "$^", strings.Join(QuoteList(rule.Prereqs()), " "), -1)

	var firstPrereq string
	if len(rule.Prereqs()) > 0 {
		firstPrereq = Quote(rule.Prereqs()[0])
	}
	s = strings.Replace(s, "$<", firstPrereq, -1)

	return s
}

// Marshal returns the textual representation of the Makefile, in the
// usual format:
//
//   target: prereqs
//   	recipes
//
//   ...
//
func Marshal(mf *Makefile) ([]byte, error) {
	var b bytes.Buffer

	for i, rule := range mf.Rules {
		if i != 0 {
			fmt.Fprintln(&b)
		}

		ruleName := rule.Target()
		fmt.Fprintf(&b, "%s:", ruleName)
		for _, prereq := range rule.Prereqs() {
			fmt.Fprintf(&b, " %s", prereq)
		}
		fmt.Fprintln(&b)
		for _, recipe := range rule.Recipes() {
			fmt.Fprintf(&b, "\t%s\n", recipe)
		}
	}

	return b.Bytes(), nil
}

var cleanRE = regexp.MustCompile(`^[\w\d_/.-]+$`)

// Quote IS NOT A SAFE WAY TO ESCAPE USER INPUT. It hackily escapes
// special characters in s and surrounds it with quotation marks if
// needed, so that the shell interprets it as a single argument equal
// to s. DON'T RELY ON THIS FOR SECURITY.
//
// TODO(sqs): come up with a safe way of escaping user input
func Quote(s string) string {
	if cleanRE.MatchString(s) {
		return s
	}
	q := strconv.Quote(s)
	return "'" + strings.Replace(q[1:len(q)-1], "'", "", -1) + "'"
}

// QuoteList IS NOT A SAFE WAY TO ESCAPE USER INPUT. It returns a list
// whose elements are the escaped elements of ss (using Quote). DON'T
// RELY ON THIS FOR SECURITY.
//
// TODO(sqs): come up with a safe way of escaping user input
func QuoteList(ss []string) []string {
	q := make([]string, len(ss))
	for i, s := range ss {
		q[i] = Quote(s)
	}
	return q
}

// Targets returns the list of targets defined by rules.
func Targets(rules []Rule) []string {
	targets := make([]string, len(rules))
	for i, rule := range rules {
		targets[i] = rule.Target()
	}
	return targets
}
