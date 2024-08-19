// package ignore provides helpers to support ignore-files similar to .gitignore
package ignore

import (
	"bufio"
	"io"
	"strings"

	"github.com/gobwas/glob"
)

var (
	lineComment = "#"
	IgnoreFile  = ".sourcegraph/ignore"
)

type Matcher struct {
	ignoreList []glob.Glob
}

// ParseIgnoreFile parses an ignore-file according to the following rules
//
// - each line represents a glob-pattern relative to the root of the repository
// - for patterns without any glob-characters, a trailing ** is implicit
// - lines starting with # are ignored
// - empty lines are ignored
func ParseIgnoreFile(r io.Reader) (matcher *Matcher, error error) {
	var patterns []glob.Glob
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// ignore empty lines
		if line == "" {
			continue
		}
		// ignore comments
		if strings.HasPrefix(line, lineComment) {
			continue
		}
		line = strings.TrimPrefix(line, "/")
		// implicit ** for patterns without glob-characters
		if !strings.ContainsAny(line, ".][*?") {
			line += "**"
		}
		// with separators = '/', * becomes path-aware
		pattern, err := glob.Compile(line, '/')
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, pattern)
	}
	return &Matcher{ignoreList: patterns}, scanner.Err()
}

// Match returns true if path has a prefix in common with any item in m.ignoreList
func (m *Matcher) Match(path string) bool {
	if len(m.ignoreList) == 0 {
		return false
	}
	for _, pattern := range m.ignoreList {
		if pattern.Match(path) {
			return true
		}
	}
	return false
}
