package proto

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// FindOwners returns the Owners associated with given path as per this CODEOWNERS file.
// Rules are evaluated in order: Returned owners come from the rule which pattern matches
// given path, that is the furthest down the file.
func (f *File) FindOwners(path string) []*Owner {
	var owners []*Owner
	for _, r := range f.GetRule() {
		m, err := compile(r.GetPattern())
		if err != nil {
			continue
		}
		if m.match(path) {
			owners = r.GetOwner()
		}
	}
	return owners
}

const separator = "/"

// globPattern implements a pattern for matching file paths,
// which can use directory/file names, * and ** wildcards,
// and may or may not be anchored to the root directory.
type globPattern []patternPart

// compile translates a text representation of a glob pattern
// to an executable one that can `match` file paths.
func compile(pattern string) (globPattern, error) {
	var p []patternPart
	// No leading `/` is equivalent to prefixing with `/**/`.
	// The pattern matches arbitrarily down the directory tree.
	if !strings.HasPrefix(pattern, separator) {
		p = append(p, anySubPath{})
	}
	for _, part := range strings.Split(strings.Trim(pattern, separator), separator) {
		switch {
		case part == "":
			return nil, errors.New("two consecutive forward slashes")
		case part == "*":
			p = append(p, anyMatch{})
		case part == "**":
			p = append(p, anySubPath{})
		default:
			p = append(p, exactMatch(part))
		}
	}
	// Trailing `/` is equivalent with trailing `/**/*`.
	// Such pattern matches any files within the directory sub-tree
	// anchored at the director that the pattern describes.
	if strings.HasSuffix(pattern, separator) {
		p = append(p, anySubPath{}, anyMatch{})
	}
	return p, nil
}

// match iterates over `filePath` separated by `/`. It uses a bit vector
// to track which prefixes of glob pattern match the file path prefix so far.
// Bit vector indices correspond to separators between pattern parts.
// Visualized matching of `/src/java/test/UnitTest.java`:
// / ** / src / java / test / ** / *Test.java   | Glob pattern
// 0    1     2      3      4    5            6 | Bit vector index
// *    *                                       | / (starting state)
// *    *     *                                 | /src
// *    *            *                          | /src/java
// *    *                   *    *              | /src/java/test
// *    *                   *    *            * | /src/java/test/UnitTest.java
// The match is successful if after iterating throght the whole file path,
// full pattern matches, that is there is a bit at the end of the glob.
func (p globPattern) match(filePath string) bool {
	currentState := big.NewInt(0)
	p.markEmptyMatches(currentState)
	parts := strings.Split(strings.Trim(filePath, separator), separator)
	nextState := big.NewInt(0)
	for _, part := range parts {
		nextState.SetInt64(0)
		p.consume(part, currentState, nextState)
		currentState, nextState = nextState, currentState
	}
	return p.matchesWhole(currentState)
}

// markEmptyMathes initializes a matching state with positions that are
// matches for an empty input (`/“). This is most often just bit 0, but in case
// there are sub-path wild-card **, it is expanded to all indices past the
// wild-cards, since they match empty path.
func (p globPattern) markEmptyMatches(state *big.Int) {
	state.SetBit(state, 0, 1)
	for i, p := range p {
		if _, ok := p.(anySubPath); !ok {
			break
		}
		state.SetBit(state, i+1, 1)
	}
}

// matchesWhole returns true if given state indicates whole glob being matched.
func (p globPattern) matchesWhole(state *big.Int) bool {
	return state.Bit(len(p)) == 1
}

// consume advances matching algorithm by a single part of a file path.
// The `current` bit vector is the matching state for up until, but excluding
// given `part` of the file path. The result - next set of states - is written
// to bit vector `next`, which is assumed to be zero when passed in.
func (p globPattern) consume(part string, current, next *big.Int) {
	// Since `**` or `anySubPath` can match any number of times, we hold
	// an invariant: If a bit vector has 1 at the state preceding `**`,
	// then that bit vector also has 1 at the state following `**`.
	for i := 0; i < len(p); i++ {
		if current.Bit(i) == 0 {
			continue
		}
		// Case 1: `current` matches before i-th part of the pattern,
		// so set the i+1-th position of the `next` state to whether
		// the i-th pattern matches (consumes) `part`.
		bit := uint(0)
		if p[i].Match(part) {
			bit = uint(1)
		}
		next.SetBit(next, i+1, bit)
		// Keep the invariant: if there is `**` afterwards, set it
		// to the same bit. This will not be overridden in the next
		// loop turns as `**` always matches.
		if i+1 < len(p) {
			if _, ok := p[i+1].(anySubPath); ok {
				next.SetBit(next, i+2, bit)
			}
		}
		// Case 2: To allow `**` to consume subsequent parts of the file path,
		// we keep the i-th bit - which precedes `**` - set.
		if _, ok := p[i].(anySubPath); ok {
			next.SetBit(next, i, 1)
		}
	}
}

// debugString prints out given state for this glob pattern
// where glob is printed, but instead of `/` separators,
// there is either X or _ which indicate bit set or unset
// in state. Very helpful for debugging.
func (p globPattern) debugString(state *big.Int) string {
	var s strings.Builder
	for i, p := range p {
		if state.Bit(i) != 0 {
			s.WriteByte('X')
		} else {
			s.WriteByte('_')
		}
		fmt.Fprint(&s, p.String())
	}
	if state.Bit(len(p)) != 0 {
		s.WriteByte('X')
	} else {
		s.WriteByte('_')
	}
	return s.String()
}

// patternPart implements matching for a single chunk of a glob pattern
// when separated by `/`.
type patternPart interface {
	String() string
	// Match is true if given file or directory name on the path matches
	// this part of the glob pattern.
	Match(string) bool
}

// anySubPath is indicated by ** in glob patterns, and matches arbitrary
// number of parts.
type anySubPath struct{}

func (p anySubPath) String() string      { return "**" }
func (p anySubPath) Match(_ string) bool { return true }

// exactMatch is indicated by an exact name of directory or a file within
// the glob pattern, and matches that exact part of the path only.
type exactMatch string

func (p exactMatch) String() string         { return string(p) }
func (p exactMatch) Match(part string) bool { return string(p) == part }

// anyMatch is indicated by * in a glob pattern, and matches any single file
// or directory on the path.
type anyMatch struct{}

func (p anyMatch) String() string         { return "*" }
func (p anyMatch) Match(part string) bool { return true }
