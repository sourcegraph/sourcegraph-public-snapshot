// Package ignorefiles deals with the ".terraformignore" file format, which
// is a convention similar to ".gitignore" that specifies path patterns that
// match files Terraform should discard or ignore when interpreting a package
// fetched from a remote location.
package ignorefiles

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// A Ruleset is the result of reading, parsing, and compiling a
// ".terraformignore" file.
type Ruleset struct {
	rules []rule
}

// ParseIgnoreFileContent takes a reader over the content of a .terraformignore
// file and returns the Ruleset described by that file, or an error if the
// file is invalid.
func ParseIgnoreFileContent(r io.Reader) (*Ruleset, error) {
	rules, err := readRules(r)
	if err != nil {
		return nil, err
	}
	return &Ruleset{rules: rules}, nil
}

// LoadPackageIgnoreRules implements reasonable default behavior for finding
// ignore rules for a particular package root directory: if .terraformignore is
// present then use it, or otherwise just return DefaultRuleset.
//
// This function will return an error only if an ignore file is present but
// unreadable, or if an ignore file is present but contains invalid syntax.
func LoadPackageIgnoreRules(packageDir string) (*Ruleset, error) {
	file, err := os.Open(filepath.Join(packageDir, ".terraformignore"))
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultRuleset, nil
		}
		return nil, fmt.Errorf("cannot read .terraformignore: %s", err)
	}
	defer file.Close()

	ret, err := ParseIgnoreFileContent(file)
	if err != nil {
		// The parse errors already mention that they were parsing ignore rules,
		// so don't need an additional prefix added.
		return nil, err
	}
	return ret, nil
}

// Excludes tests whether the given path matches the set of paths that are
// excluded by the rules in the ruleset.
//
// If any of the rules in the ruleset have invalid syntax then Excludes will
// return an error, but it will also still return a boolean result which
// considers all of the remaining valid rules, to support callers that want to
// just ignore invalid exclusions. Such callers can safely ignore the error
// result:
//
//     exc, _ = ruleset.Excludes(path)
func (r *Ruleset) Excludes(path string) (bool, error) {
	if r == nil {
		return false, nil
	}

	var retErr error
	foundMatch := false
	for _, rule := range r.rules {
		match, err := rule.match(path)
		if err != nil {
			// We'll remember the first error we encounter, but continue
			// matching anyway to support callers that want to ignore invalid
			// lines and just match with whatever's left.
			if retErr == nil {
				retErr = fmt.Errorf("invalid ignore rule %q", rule.val)
			}
		}
		if match {
			foundMatch = !rule.excluded
		}
	}
	return foundMatch, retErr
}

// Includes is the inverse of [Ruleset.Excludes].
func (r *Ruleset) Includes(path string) (bool, error) {
	notRet, err := r.Excludes(path)
	return !notRet, err
}

var DefaultRuleset *Ruleset

func init() {
	DefaultRuleset = &Ruleset{rules: defaultExclusions}
}
